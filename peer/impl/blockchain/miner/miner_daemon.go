package miner

import (
	"fmt"
	"go.dedis.ch/cs438/peer/impl/blockchain/block"
	"go.dedis.ch/cs438/peer/impl/blockchain/transaction"
	"go.dedis.ch/cs438/types"
	"sort"
	"time"
)

func (m *Miner) txProcessingDaemon() {
	defer m.wg.Done()
	for {
		select {
		case <-(*m.GetContext()).Done():
			return
		default:
			// 1. reset miner's tmp world state before processing txs and forming a new block
			m.mu.Lock()
			m.resetTmpWorldState()
			prevHash := m.chain.Tail.BlockHash
			m.mu.Unlock()

			// 2. process txs until a block is formed or timeout or new block from others is appended
			err := m.processTxs()
			if err != nil {
				m.logger.Error().Msg("failed to process tx : " + err.Error())
				continue
			}

			// 3. form a new block
			b := m.formBlock(prevHash)
			if b == nil {
				continue
			}

			// 4. proof of work
			err = b.ProofOfWork(m.GetConf().BlockchainDifficulty, m.GetContext())
			if err != nil {
				continue
			}

			// 5. try to append new block to blockchain and broadcast it
			ok := m.chain.CheckNewBlock(b)
			if !ok {
				m.revertBlock(b)
				continue
			} else {
				blockMsg := types.BlockMessage{Block: *b}
				blockTransMsg, err := m.message.GetConf().MessageRegistry.MarshalMessage(blockMsg)
				if err != nil {
					m.logger.Error().Err(err).Msg("failed to marshal block message")
					continue
				}

				err = m.message.Broadcast(blockTransMsg)
				if err != nil {
					m.logger.Error().Err(err).Msg("failed to broadcast block message")
					continue
				}
			}
		}
	}
}

func (m *Miner) processTxs() error {
	start := time.Now()

	// process transactions until
	// 1. BlockSize is reached  2. timeout is reached
	for m.txProcessed.Len() < m.message.GetConf().BlockchainBlockSize ||
		time.Now().Sub(start) < m.message.GetConf().BlockchainBlockTimeout {

		select {
		case <-(*m.GetContext()).Done():
			return nil
		default:
			err := m.processOneTx()
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func (m *Miner) processOneTx() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.txPending.IsEmpty() {
		return nil
	}

	err := m.executeTransaction(m.txPending.Dequeue())
	if err != nil {
		return err
	}

	return nil
}

func (m *Miner) formBlock(prevHash string) *block.Block {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check prevHash: If a new block from other is appended, invalid this miner's own block
	if prevHash != m.chain.Tail.BlockHash || m.txProcessed.IsEmpty() {
		m.cleanTxPool()
		return nil
	}

	// Form a new block
	b := m.chain.NextBlock()

	// Add all processed txs
	for !m.txProcessed.IsEmpty() {
		tx := m.txProcessed.Dequeue()
		b.TXs = append(b.TXs, tx)
	}

	b.State = m.tmpWorldState.Copy()

	return b
}

// cleanTxPool will clean the transaction pool based on current blockchain.
// It must be called under an outlier protection of mutex
func (m *Miner) cleanTxPool() {

	tmp := make([]*transaction.SignedTransaction, 0)

	for !m.txPending.IsEmpty() {
		tmp = append(tmp, m.txPending.Dequeue())
	}
	for !m.txProcessed.IsEmpty() {
		tmp = append(tmp, m.txProcessed.Dequeue())
	}
	for !m.txInvalid.IsEmpty() {
		tmp = append(tmp, m.txInvalid.Dequeue())
	}

	cleaned := make([]*transaction.SignedTransaction, len(tmp))
	for _, tx := range tmp {
		if !m.chain.HasTransaction(tx.HashCode()) {
			cleaned = append(cleaned, tx)
		}
	}

	// Sort the cleaned txs based on their timestamps
	sort.Slice(cleaned, func(i, j int) bool {
		return cleaned[i].TX.Timestamp < cleaned[j].TX.Timestamp
	})

	// Put all cleaned txs back to txPending
	for _, tx := range cleaned {
		m.txPending.Enqueue(tx)
	}
}

func (m *Miner) revertBlock(b *block.Block) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Put all processed txs in this block back to txPending
	for _, tx := range b.TXs {
		if !m.chain.HasTransaction(tx.HashCode()) {
			m.txPending.Enqueue(tx)
		}
	}

	// Clean the txs to be prepared for next block generation
	m.cleanTxPool()
}

//func (m *Miner) blockProcessingDaemon() {
//	for {
//		select {
//		case <-m.CTX.Done():
//			return
//		case blockMsg := <-m.blockInCh:
//			err := m.processBlock(blockMsg)
//			if err != nil {
//				m.logger.Error().Err(err).Msg("failed to process a block")
//			}
//		}
//	}
//}

func (m *Miner) processBlock(blockMsg *types.BlockMessage) error {
	// Move the locking to handler function
	//m.mu.Lock()
	//defer m.mu.Unlock()

	// Add the new block to the buffer
	m.blockBuffer.Store(blockMsg.Block.ID, blockMsg)

	// Append blocks as much as possible
	for {
		// Try to retrieve the next block from the buffer
		nextID := m.chain.Tail.ID + 1
		nextBlockMsg, ok := m.blockBuffer.Load(nextID)
		if !ok {
			return nil
		}
		m.blockBuffer.Delete(nextID)

		// Try to append the next block
		nextBlock := nextBlockMsg.(*types.BlockMessage).Block
		ok = m.chain.CheckNewBlock(&nextBlock)
		if !ok {
			return fmt.Errorf("check new block failed")
		}

		// Append the new block
		err := m.chain.AppendBlock(&nextBlock)
		if err != nil {
			return err
		}

		m.cleanTxPool()
	}
}

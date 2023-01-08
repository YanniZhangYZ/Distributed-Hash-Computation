package miner

import (
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
			// 1. Reset miner's tmp world state before processing txs and forming a new block
			m.mu.Lock()
			m.resetTmpWorldState()
			preparingBlockID := m.chain.Tail.ID + 1
			notifyCh := m.blockNotificationCh[int(preparingBlockID)]
			m.mu.Unlock()

			// 2. Process txs until a block is formed or timeout or new block from others is appended
			m.processTxs(notifyCh)

			// 3. Form a new block
			b := m.formBlock(preparingBlockID)
			if b == nil {
				continue
			}

			// 4. Proof of work
			start := time.Now()
			m.logger.Debug().Msg("block formed, begin Proof of Work...")

			err := b.ProofOfWork(m.GetConf().BlockchainDifficulty, m.GetContext(), notifyCh)
			if err != nil {
				m.logger.Debug().Str("reason", err.Error()).Msg("Proof of Work failed")
				continue
			}
			m.logger.Debug().Msgf("Proof of Work finished in %f seconds", time.Now().Sub(start).Seconds())

			// 5. check the new block and broadcast it if valid
			err = m.chain.CheckNewBlock(b)
			if err != nil {
				m.logger.Debug().Err(err).Uint32("blockID", b.ID).Msg("discard an invalid mined block")
				m.revertBlock(b)
				continue
			} else {
				blockMsg := types.BlockMessage{TransBlock: *b.GetTransBlock()}
				blockTransMsg, err := m.message.GetConf().MessageRegistry.MarshalMessage(blockMsg)
				if err != nil {
					m.logger.Error().Err(err).Msg("fail to marshal block message")
					continue
				}

				err = m.message.Broadcast(blockTransMsg)
				if err != nil {
					m.logger.Error().Err(err).Msg("fail to broadcast block message")
					continue
				}
				m.logger.Debug().Uint32("blockID", b.ID).Msg("mined block is valid and broadcast")
			}
		}
	}
}

func (m *Miner) processTxs(notifyCh chan struct{}) {
	start := time.Now()

	// process transactions until
	// 1. BlockSize is reached  2. timeout is reached
	for m.txProcessed.Len() < m.message.GetConf().BlockchainBlockSize &&
		time.Now().Sub(start) < m.message.GetConf().BlockchainBlockTimeout {
		select {
		case <-(*m.GetContext()).Done():
			return
		case <-notifyCh:
			return
		default:
			m.processOneTx()
		}
	}
}

// processOneTx retrieve one transaction from miner's txPending and verify and execute it
func (m *Miner) processOneTx() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.txPending.IsEmpty() {
		return
	}

	tx := m.txPending.Dequeue()
	err := transaction.VerifyAndExecuteTransaction(tx, &(m.tmpWorldState))

	if err == nil {
		m.txProcessed.Enqueue(tx)
		m.logger.Debug().
			Int("type", tx.TX.Type).
			Str("src", tx.TX.Src.String()).
			Str("dst", tx.TX.Dst.String()).
			Int("nonce", tx.TX.Nonce).
			Int64("value", tx.TX.Value).
			Uint64("timestamp", tx.TX.Timestamp).
			Str("code", tx.TX.Code).
			Str("data", tx.TX.Data).
			Msg("enqueue a confirmed transaction")
	} else {
		m.txInvalid.Enqueue(tx)
		m.logger.Debug().
			Err(err).
			Int("type", tx.TX.Type).
			Str("src", tx.TX.Src.String()).
			Str("dst", tx.TX.Dst.String()).
			Int("nonce", tx.TX.Nonce).
			Int64("value", tx.TX.Value).
			Uint64("timestamp", tx.TX.Timestamp).
			Str("code", tx.TX.Code).
			Str("data", tx.TX.Data).
			Msg("discard an invalid transaction")
	}
	return
}

func (m *Miner) formBlock(preparingBlockID uint32) *block.Block {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check prevHash: If a new block from other is appended, invalid this miner's own block
	if preparingBlockID != m.chain.Tail.ID+1 || m.txProcessed.IsEmpty() {
		if !m.txProcessed.IsEmpty() {
			m.cleanTxPool()
		}
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

	cleaned := make([]*transaction.SignedTransaction, 0)
	for _, tx := range tmp {
		if !m.chain.HasTransaction(tx.HashCode()) {
			cleaned = append(cleaned, tx)
		}
	}

	// Sort the cleaned txs based on their timestamp and nonce
	sort.Slice(cleaned, func(i, j int) bool {
		if cleaned[i].TX.Src.String() == cleaned[j].TX.Src.String() {
			return cleaned[i].TX.Nonce < cleaned[j].TX.Nonce
		} else {
			return cleaned[i].TX.Timestamp < cleaned[j].TX.Timestamp
		}
	})

	// Put all cleaned txs back to txPending
	for _, tx := range cleaned {
		m.txPending.Enqueue(tx)
	}

	m.resetTmpWorldState()

	m.logger.Debug().Uint("#txPending", m.txPending.Len()).Msg("transaction pool cleaned")
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

func (m *Miner) processBlock(blockMsg *types.BlockMessage) error {
	// m.mu.Lock() must be done by the caller

	// Recover the real block from TransBlock
	b := blockMsg.TransBlock.GetBlock()

	// Add the new block to the buffer
	m.blockBuffer.Store(b.ID, b)
	m.logger.Debug().Uint32("blockID", b.ID).Msgf("buffered a received block")

	// Append blocks as much as possible
	for {
		// Try to retrieve the next block from the buffer
		nextID := m.chain.Tail.ID + 1
		nextBlockAny, ok := m.blockBuffer.Load(nextID)
		if !ok {
			return nil
		}
		m.blockBuffer.Delete(nextID)

		// Try to append the next block
		nextBlock := nextBlockAny.(*block.Block)
		err := m.chain.CheckNewBlock(nextBlock)
		if err != nil {
			m.logger.Debug().Err(err).Uint32("blockID", nextBlock.ID).
				Msg("appending a buffered block failed")
			return err
		}

		// Append the new block
		err = m.chain.AppendBlock(nextBlock)
		if err != nil {
			return err
		}
		m.logger.Debug().Uint32("blockID", nextBlock.ID).Msg("new block appended")

		// Notify the completion of this block
		close(m.blockNotificationCh[int(nextBlock.ID)])

		// Create the channel for the next block
		m.blockNotificationCh[int(nextBlock.ID+1)] = make(chan struct{})

		m.cleanTxPool()
	}
}

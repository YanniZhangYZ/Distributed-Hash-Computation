package passwordcracker

import (
	"encoding/hex"
	"github.com/rs/zerolog/log"
	"go.dedis.ch/cs438/transport"
	"go.dedis.ch/cs438/types"
	"golang.org/x/xerrors"
	"time"
)

// execPasswordCrackerRequestMessage is the callback function to handle PasswordCrackerRequestMessage
func (p *PasswordCracker) execPasswordCrackerRequestMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	passwordCrackerRequestMsg, ok := msg.(*types.PasswordCrackerRequestMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	crackPasswordAndReply := func() {

		password := p.crackPassword(passwordCrackerRequestMsg.Hash, passwordCrackerRequestMsg.Salt)

		// Execute the smart contract to earn the reward
		if passwordCrackerRequestMsg.ContractAddress.String() != "" {
			_ = p.blockchain.ExecuteContract(password,
				hex.EncodeToString(passwordCrackerRequestMsg.Hash),
				hex.EncodeToString(passwordCrackerRequestMsg.Salt),
				passwordCrackerRequestMsg.ContractAddress.String(),
				time.Second*600)
		}

		passwordCrackerReplyMsg := types.PasswordCrackerReplyMessage{
			Hash:     passwordCrackerRequestMsg.Hash,
			Salt:     passwordCrackerRequestMsg.Salt,
			Password: password,
		}
		passwordCrackerReplyMsgTrans, err := p.conf.MessageRegistry.MarshalMessage(passwordCrackerReplyMsg)
		if err != nil {
			log.Error().Err(err).Msg("execPasswordCrackerRequestMessage MarshalMessage")
		}
		err = p.message.SendDirectMsg(pkt.Header.Source, pkt.Header.Source, passwordCrackerReplyMsgTrans)
		if err != nil {
			log.Error().Err(err).Msg("execPasswordCrackerRequestMessage SendDirectMsg")
		}
	}
	go crackPasswordAndReply()
	return nil
}

// execPasswordCrackerReplyMessage is the callback function to handle PasswordCrackerReplyMessage
func (p *PasswordCracker) execPasswordCrackerReplyMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	passwordCrackerReplyMsg, ok := msg.(*types.PasswordCrackerReplyMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	/* Record the result into our task pool */
	taskKey := hex.EncodeToString(append(passwordCrackerReplyMsg.Hash, passwordCrackerReplyMsg.Salt...))
	taskResult := map[string]string{"password": passwordCrackerReplyMsg.Password}
	p.tasks.Store(taskKey, taskResult)
	return nil
}

// execPasswordCrackerUpdDictRangeMessage is the callback function to handle PasswordCrackerUpdDictRangeMessage
func (p *PasswordCracker) execPasswordCrackerUpdDictRangeMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	passwordCrackerUpdDictRangeMsg, ok := msg.(*types.PasswordCrackerUpdDictRangeMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	// Update the range
	p.updDictRange(passwordCrackerUpdDictRangeMsg.Start, passwordCrackerUpdDictRangeMsg.End)
	return nil
}

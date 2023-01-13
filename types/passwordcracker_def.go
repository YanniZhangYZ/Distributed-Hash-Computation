package types

import "go.dedis.ch/cs438/peer/impl/blockchain/common"

// PasswordCrackerRequestMessage request a peer to crack the specified hash + salt
//
// - implements types.Message
type PasswordCrackerRequestMessage struct {
	// Hash is the hash to crack
	Hash []byte

	// Salt is the salt to compute the hash
	Salt []byte

	// ContractAddress is the smart contract account address corresponding to this request
	ContractAddress common.Address
}

// PasswordCrackerReplyMessage replies with the answer to the password cracking request
//
// - implements types.Message
type PasswordCrackerReplyMessage struct {
	// Hash is the hash to crack
	Hash []byte

	// Salt is the salt to compute the hash
	Salt []byte

	// Password is the cracked password, if any
	Password string
}

// PasswordCrackerUpdDictRangeMessage is sent by the Chord component to update about the salt range that the
// node is responsible for the password cracker
type PasswordCrackerUpdDictRangeMessage struct {
	// Start range of the salt value
	Start uint

	// End range of the salt value
	End uint
}

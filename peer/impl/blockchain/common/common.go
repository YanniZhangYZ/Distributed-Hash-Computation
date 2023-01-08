package common

import (
	"container/list"
	"sync"
)

type KVStore[TV any] interface {
	Len() int

	Get(key string) (value TV, ok bool)
	Set(key string, value TV)
	Delete(key string)

	Copy() KVStore[TV]
	Hash() []byte
	HashCode() string
	ForEach(fn func(key string, value TV) bool) bool
	GetSimpleMap() map[string]TV
	Keys() []string
	Equal(other *KVStore[TV]) bool
}

// SafeQueue is a thread-safe version of queue
type SafeQueue[T any] struct {
	q  *list.List
	mu sync.Mutex
}

type Address struct {
	//Addr      [8]byte
	HexString string
}

type State struct {
	// Nonce – A counter that indicates the number of transactions sent from the account.
	// This ensures transactions are only processed once.
	// In a contract account, this number represents the number of contracts created by the account.
	Nonce int

	// Balance – The number of money owned by this address.
	Balance int64

	// CodeHash – This is a hash refers to the code of an account on the Ethereum virtual machine (EVM).
	// CodeHash is DISABLED for Externally owned account (EOA). This field is set to an empty string for EOAs.
	// Contract accounts have code fragments programmed in that can perform different operations.
	// This code gets executed if the account gets a message call. It cannot be changed.
	// All such code fragments are contained in the state database under their corresponding hashes for later retrieval.
	// This hash value is known as a codeHash.
	CodeHash string

	// StorageRoot – Sometimes known as a storage hash.
	// StorageRoot is DISABLED for Externally owned account (EOA). This field is set to an empty string for EOAs.
	// A 256-bit hash of the root node of a Merkle Patricia trie (or a simple KVStore) that
	// encodes the storage contents of the account (a mapping between 256-bit integer values),
	// encoded into the trie as a mapping from the Keccak 256-bit hash of the 256-bit integer keys to the RLP-encoded 256-bit integer values.
	// This trie encodes the hash of the storage contents of this account, and is empty by default.
	StorageRoot string
}

type WorldState = KVStore[State]

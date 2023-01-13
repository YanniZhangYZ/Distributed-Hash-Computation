package common

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type State struct {
	// Nonce – A counter that indicates the number of transactions sent from the account.
	// This ensures transactions are only processed once.
	// In a contract account, this number represents the number of contracts created by the account.
	Nonce int

	// Balance – The number of money owned by this address.
	Balance int64

	CodeHash string

	// Contract – This is the code of an account on the Ethereum virtual machine (EVM).
	// DISABLED for Externally owned account (EOA). This field is set to an empty string for EOAs.
	// Contract accounts have code fragments programmed in that can perform different operations.
	// This code gets executed if the account gets a message call. It cannot be changed.
	// All such code fragments are contained in the state database under their corresponding hashes for later retrieval.
	// This hash value is known as a codeHash.
	Contract []byte

	// StorageRoot – Sometimes known as a storage hash.
	// StorageRoot is DISABLED for Externally owned account (EOA). This field is set to an empty string for EOAs.
	// A 256-bit hash of the root node of a Merkle Patricia trie (or a simple KVStore) that
	// encodes the storage contents of the account (a mapping between 256-bit integer values),
	// encoded into the trie as a mapping from the Keccak 256-bit hash of the 256-bit integer keys to the RLP-encoded 256-bit integer values.
	// This trie encodes the hash of the storage contents of this account, and is empty by default.
	StorageRoot string

	// Tasks – The map that keeps record of all password-cracking tasks that have been executed by this account.
	// hash -> [password, salt]
	Tasks map[string][2]string
}

func (a *State) String() string {
	s := ""
	s += fmt.Sprintf("Balance %d, Nonce %d, CodeHash %s, StorageRoot %s",
		a.Balance, a.Nonce, a.CodeHash, a.StorageRoot)
	for k, v := range a.Tasks {
		s += fmt.Sprintf(", Hash-%s_Password-%s_Salt-%s", k, v[0], v[1])
	}
	return s
}

func (a *State) Hash() []byte {
	hash := sha256.New()
	hash.Write([]byte(a.String()))
	return hash.Sum(nil)
}

func (a *State) HashCode() string {
	return hex.EncodeToString(a.Hash())
}

func (a *State) Equals(other State) bool {

	if a.Nonce != other.Nonce {
		return false
	}

	if a.Balance != other.Balance {
		return false
	}

	if a.CodeHash != other.CodeHash {
		return false
	}

	if len(a.Contract) != len(other.Contract) {
		return false
	}

	for i := 0; i < len(a.Contract); i++ {
		if a.Contract[i] != other.Contract[i] {
			return false
		}
	}

	if a.StorageRoot != other.StorageRoot {
		return false
	}

	if len(a.Tasks) != len(other.Tasks) {
		return false
	}

	for k, v1 := range a.Tasks {
		v2, ok := other.Tasks[k]
		if !ok {
			return false
		}
		if v1[0] != v2[0] || v1[1] != v2[1] {
			return false
		}
	}

	return true
}

func (a *State) Copy() State {
	cpy := State{}
	cpy.Nonce = a.Nonce
	cpy.Balance = a.Balance
	cpy.CodeHash = a.CodeHash
	cpy.StorageRoot = a.StorageRoot

	cpy.Contract = make([]byte, len(a.Contract))
	copy(cpy.Contract, a.Contract)

	cpy.Tasks = make(map[string][2]string)
	for k, v := range a.Tasks {
		cpy.Tasks[k] = [2]string{v[0], v[1]}
	}
	return cpy
}

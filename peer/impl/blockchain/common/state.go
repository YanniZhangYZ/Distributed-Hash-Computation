package common

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func (a *State) String() string {
	s := ""
	s += fmt.Sprintf("Balance %d, Nonce %d, CodeHash %s, StorageRoot %s",
		a.Balance, a.Nonce, a.CodeHash, a.StorageRoot)
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

func QuickWorldState(accounts int, balance int64) WorldState {
	worldState := NewKVStore[State]()
	for i := 0; i < accounts; i++ {
		worldState.Set(fmt.Sprintf("%d", i+1), State{
			Nonce:       0,
			Balance:     balance,
			CodeHash:    "",
			StorageRoot: "",
		})
	}

	return worldState
}

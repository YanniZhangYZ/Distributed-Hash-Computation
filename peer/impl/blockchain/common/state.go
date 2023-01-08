package common

import "fmt"

func (a *State) String() string {
	s := ""
	s += fmt.Sprintf("Balance %d, Nonce %d, CodeHash %s, StorageRoot %s",
		a.Balance, a.Nonce, a.CodeHash, a.StorageRoot)
	return s
}

func (a *State) Hash() []byte {
	//TODO implement me
	panic("implement me")
}

func (a *State) HashCode() string {
	//TODO implement me
	panic("implement me")
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

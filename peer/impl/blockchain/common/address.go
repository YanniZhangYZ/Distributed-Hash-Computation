package common

func (a *Address) String() string {
	return a.HexString
}

func (a *Address) Hash() []byte {
	//TODO implement me
	panic("implement me")
}

func (a *Address) HashCode() string {
	//TODO implement me
	panic("implement me")
}

func StringToAddress(s string) Address {
	return Address{HexString: s}
}

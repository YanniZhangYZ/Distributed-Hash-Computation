package common

import (
	"crypto/sha256"
	"encoding/hex"
)

func (a *Address) String() string {
	return a.HexString
}

func (a *Address) Hash() []byte {
	hash := sha256.New()
	hash.Write([]byte(a.HexString))
	return hash.Sum(nil)
}

func (a *Address) HashCode() string {
	return hex.EncodeToString(a.Hash())
}

func StringToAddress(s string) Address {
	return Address{HexString: s}
}

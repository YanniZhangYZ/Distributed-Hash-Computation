package password_cracker

import (
	"crypto"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/cs438/peer"
	"go.dedis.ch/cs438/storage/inmemory"
	"testing"
)

// Test_Hash_Password tests the hashPassword function
func Test_Hash_Password(t *testing.T) {
	p := PasswordCracker{}
	p.hashAlgo = crypto.SHA256

	password1 := "Password"
	salt1 := []byte{0x0, 0x0}
	passwordHash1 := []byte{0x48, 0x4f, 0x95, 0x73, 0x38, 0xd, 0x13, 0xc3, 0x4, 0x2d, 0x36, 0x1, 0xb2, 0x0,
		0x1b, 0x61, 0x1d, 0x2, 0xf4, 0xec, 0xc8, 0x8a, 0xf2, 0x23, 0x5e, 0xc3, 0x18, 0xd, 0xe7, 0xbd, 0x96, 0x2c}
	require.Equal(t, passwordHash1, p.hashPassword(password1, salt1))

	password2 := "apple"
	salt2 := []byte{0x0, 0x3c}
	passwordHash2 := []byte{0x6a, 0xd1, 0x8f, 0x94, 0xf, 0xfb, 0xd3, 0x4, 0x54, 0xe3, 0xc2, 0xec, 0xf6, 0x17,
		0x8c, 0x64, 0x92, 0xde, 0xb3, 0x3c, 0xd2, 0xfa, 0x14, 0x2d, 0xad, 0x3b, 0x41, 0x17, 0x62, 0xa5, 0x78, 0x60}
	require.Equal(t, passwordHash2, p.hashPassword(password2, salt2))

	//for i := 0x0f; i <= 0xff; i += 0x10 {
	//	fmt.Println(hex.EncodeToString(p.hashPassword("apple", []byte{byte(i)})))
	//	fmt.Println(hex.EncodeToString([]byte{byte(i)}))
	//}
}

// Test_Create_Dictionary tests the createDictionary function
func Test_Create_Dictionary(t *testing.T) {
	p := PasswordCracker{}
	p.hashAlgo = crypto.SHA256
	p.conf = &peer.Configuration{}
	p.conf.ChordBytes = 2
	p.conf.Storage = inmemory.NewPersistency()

	salt := uint(0)
	p.createDictionary(salt)
	require.Equal(t, 1, p.conf.Storage.GetDictionaryStore().Len())

	salt = uint(2)
	p.createDictionary(salt)
	require.Equal(t, 2, p.conf.Storage.GetDictionaryStore().Len())
}

// Test_Crack_Password tests the crackPassword function correctly finds the password
func Test_Crack_Password(t *testing.T) {
	p := PasswordCracker{}
	p.hashAlgo = crypto.SHA256
	p.conf = &peer.Configuration{}
	p.conf.ChordBytes = 2
	p.conf.Storage = inmemory.NewPersistency()

	salt := uint(60)
	saltBytes := []byte{0x0, 0x3c}
	p.createDictionary(salt)
	require.Equal(t, 1, p.conf.Storage.GetDictionaryStore().Len())

	hash := []byte{0x6a, 0xd1, 0x8f, 0x94, 0xf, 0xfb, 0xd3, 0x4, 0x54, 0xe3, 0xc2, 0xec, 0xf6, 0x17,
		0x8c, 0x64, 0x92, 0xde, 0xb3, 0x3c, 0xd2, 0xfa, 0x14, 0x2d, 0xad, 0x3b, 0x41, 0x17, 0x62, 0xa5, 0x78, 0x60}
	password := p.crackPassword(hash, saltBytes)
	require.Equal(t, "apple", password)

	// Test all passwords inside the dictionary can be correctly found
	salt = uint(1233)
	saltBytes = []byte{0x4, 0xD1}
	p.createDictionary(salt)
	require.Equal(t, 2, p.conf.Storage.GetDictionaryStore().Len())
	for _, word := range defaultDict {
		passwordHash := p.hashPassword(word, saltBytes)
		password = p.crackPassword(passwordHash, saltBytes)
		require.Equal(t, word, password)
	}
}

// Test_Upd_Dict_Range tests the updDictRange function
func Test_Upd_Dict_Range(t *testing.T) {
	p := PasswordCracker{}
	p.hashAlgo = crypto.SHA256
	p.conf = &peer.Configuration{}
	p.conf.ChordBytes = 2
	p.conf.Storage = inmemory.NewPersistency()

	// We are responsible for the full range
	p.updDictRange(0, 0)
	require.Equal(t, 65536, p.conf.Storage.GetDictionaryStore().Len())

	// We are crossing the upper bound boundary
	p.updDictRange(65535, 1)
	require.Equal(t, 2, p.conf.Storage.GetDictionaryStore().Len())

	// Normal case
	p.updDictRange(32, 34)
	require.Equal(t, 2, p.conf.Storage.GetDictionaryStore().Len())

	p.updDictRange(77, 1077)
	require.Equal(t, 1000, p.conf.Storage.GetDictionaryStore().Len())
}

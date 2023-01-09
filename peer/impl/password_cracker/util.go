package password_cracker

import (
	"encoding/hex"
	"encoding/json"
	"github.com/rs/zerolog/log"
	"math"
	"math/big"
)

// hashPassword combines password and salt then hash them using the configured hash algorithm and then
// return the hashed password bytes
func (p *PasswordCracker) hashPassword(password string, salt []byte) []byte {
	passwordBytes := []byte(password)
	h := p.hashAlgo.New()
	// Append salt to password
	passwordBytes = append(passwordBytes, salt...)
	h.Write(passwordBytes)
	hashedPasswordBytes := h.Sum(nil)
	return hashedPasswordBytes
}

// createDictionary creates a dictionary given a salt value, and stores the computed dictionary for later usage
func (p *PasswordCracker) createDictionary(salt uint) {
	// convert salt from uint into bytes array
	saltBytes := make([]byte, p.conf.ChordBytes)
	big.NewInt(int64(salt)).FillBytes(saltBytes)
	saltString := hex.EncodeToString(saltBytes)

	// If we already have this entry, returns directly
	unmarshaledDictionary := p.conf.Storage.GetDictionaryStore().Get(saltString)
	if unmarshaledDictionary != nil {
		return
	}

	dictionary := map[string]string{}
	for _, word := range defaultDict {
		wordHash := p.hashPassword(word, saltBytes)
		wordHashString := hex.EncodeToString(wordHash)
		dictionary[wordHashString] = word
	}
	dictionaryByte, err := json.Marshal(dictionary)
	if err != nil {
		log.Error().Err(err).Msg("PasswordCracker createDictionary Marshal")
	}
	p.conf.Storage.GetDictionaryStore().Set(saltString, dictionaryByte)
}

// deleteDictionary deletes a dictionary entry given a salt value
func (p *PasswordCracker) deleteDictionary(salt uint) {
	// convert salt from uint into bytes array
	saltBytes := make([]byte, p.conf.ChordBytes)
	big.NewInt(int64(salt)).FillBytes(saltBytes)
	saltString := hex.EncodeToString(saltBytes)
	p.conf.Storage.GetDictionaryStore().Delete(saltString)
}

// crackPassword cracks the password using the given hash and salt value, if it succeeds, it returns the
// cracked password
func (p *PasswordCracker) crackPassword(hash []byte, salt []byte) string {
	// Check that we have this dictionary, and look it up inside the dictionary
	saltString := hex.EncodeToString(salt)
	unmarshaledDictionary := p.conf.Storage.GetDictionaryStore().Get(saltString)
	if unmarshaledDictionary == nil {
		return ""
	}

	dictionary := map[string]string{}
	err := json.Unmarshal(unmarshaledDictionary, &dictionary)
	if err != nil {
		log.Error().Err(err).Msg("PasswordCracker crackPassword Unmarshal")
	}

	hashKey := hex.EncodeToString(hash)
	password := dictionary[hashKey]
	return password
}

// updDictRange updates the range of salted dictionary that this node stores
func (p *PasswordCracker) updDictRange(start uint, end uint) {
	p.dictUpdLock.Lock()
	defer p.dictUpdLock.Unlock()

	upperBound := uint(math.Pow(2, float64(p.conf.ChordBytes)*8))
	if start < end {
		for i := uint(0); i <= start; i++ {
			p.deleteDictionary(i)
		}
		for i := start + 1; i <= end; i++ {
			p.createDictionary(i)
		}
		for i := end + 1; i < upperBound; i++ {
			p.deleteDictionary(i)
		}
	} else {
		for i := uint(0); i <= end; i++ {
			p.createDictionary(i)
		}
		for i := end + 1; i <= start; i++ {
			p.deleteDictionary(i)
		}
		for i := start + 1; i < upperBound; i++ {
			p.createDictionary(i)
		}
	}
}

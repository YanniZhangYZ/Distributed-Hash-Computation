package password_cracker

// Combine password and salt then hash them using the configured hash algorithm and then
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

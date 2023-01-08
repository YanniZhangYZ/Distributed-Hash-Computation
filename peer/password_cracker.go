package peer

// PasswordCracker defines the functions for the basic password cracking operation
type PasswordCracker interface {
	// PasswordSubmitRequest submits a password cracking tasks to a remote peer, the hashStr is the password
	// hash that we would like to crack, and the saltStr is the salt value accompanying
	PasswordSubmitRequest(hashStr string, saltStr string) error

	// PasswordReceiveResult receives the result, corresponding to the SubmitRequest
	PasswordReceiveResult(hashStr string, saltStr string) string
}

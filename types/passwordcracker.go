package types

import "fmt"

// -----------------------------------------------------------------------------
// PasswordCrackerRequestMessage

// NewEmpty implements types.Message.
func (c PasswordCrackerRequestMessage) NewEmpty() Message {
	return &PasswordCrackerRequestMessage{}
}

// Name implements types.Message.
func (c PasswordCrackerRequestMessage) Name() string {
	return "passwordcrackerreq"
}

// String implements types.Message.
func (c PasswordCrackerRequestMessage) String() string {
	return fmt.Sprintf("{passwordcrackerreq hash [%x], salt [%x]}", c.Hash, c.Salt)
}

// HTML implements types.Message.
func (c PasswordCrackerRequestMessage) HTML() string {
	return c.String()
}

// -----------------------------------------------------------------------------
// PasswordCrackerReplyMessage

// NewEmpty implements types.Message.
func (c PasswordCrackerReplyMessage) NewEmpty() Message {
	return &PasswordCrackerReplyMessage{}
}

// Name implements types.Message.
func (c PasswordCrackerReplyMessage) Name() string {
	return "passwordcrackerreply"
}

// String implements types.Message.
func (c PasswordCrackerReplyMessage) String() string {
	return fmt.Sprintf("{passwordcrackerreply hash [%x], salt [%x], password [%s]}", c.Hash, c.Salt, c.Password)
}

// HTML implements types.Message.
func (c PasswordCrackerReplyMessage) HTML() string {
	return c.String()
}

// -----------------------------------------------------------------------------
// PasswordCrackerUpdDictRangeMessage

// NewEmpty implements types.Message.
func (c PasswordCrackerUpdDictRangeMessage) NewEmpty() Message {
	return &PasswordCrackerUpdDictRangeMessage{}
}

// Name implements types.Message.
func (c PasswordCrackerUpdDictRangeMessage) Name() string {
	return "passwordcrackerupddictrange"
}

// String implements types.Message.
func (c PasswordCrackerUpdDictRangeMessage) String() string {
	return fmt.Sprintf("{passwordcrackerupddictrange from %d to %d}", c.Start, c.End)
}

// HTML implements types.Message.
func (c PasswordCrackerUpdDictRangeMessage) HTML() string {
	return c.String()
}

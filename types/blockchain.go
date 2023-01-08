package types

// NewEmpty implements types.Message.
func (c TransactionMessage) NewEmpty() Message {
	return &TransactionMessage{}
}

// Name implements types.Message.
func (c TransactionMessage) Name() string {
	return "transaction message"
}

// String implements types.Message.
func (c TransactionMessage) String() string {
	return c.SignedTX.String()
}

// HTML implements types.Message.
func (c TransactionMessage) HTML() string {
	return c.String()
}

// NewEmpty implements types.Message.
func (c BlockMessage) NewEmpty() Message {
	return &BlockMessage{}
}

// Name implements types.Message.
func (c BlockMessage) Name() string {
	return "block message"
}

// String implements types.Message.
func (c BlockMessage) String() string {
	return c.TransBlock.String()
}

// HTML implements types.Message.
func (c BlockMessage) HTML() string {
	return c.String()
}

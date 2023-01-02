package chord

import (
	"go.dedis.ch/cs438/peer"
	"go.dedis.ch/cs438/peer/impl"
)

type Module struct {
	address string
	conf    *peer.Configuration // The configuration contains Socket and MessageRegistry
	message *impl.MessageModule
	// TODO
}

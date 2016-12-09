package inmem

import (
	"github.com/begizi/vch-server/tunnel"
)

/*
InMemQueue
----------

InMemQueue impliments the Queue interface in the most
basic way that allows the message sending to happen from
within the same process. This makes testing much easier and
allows the single vchd process to run without external
dependencies.

THIS DOES NOT SCALE. Only use for testing and single client
setups.

Since vchd processes can't communicate when they receive
a message, there is no guarantee that the message
will go to the process that the client is connected to.
*/

type InMemQueue struct {
	receivec tunnel.ReceiveC
}

func NewInMemQueue() tunnel.Queue {
	return InMemQueue{
		receivec: make(tunnel.ReceiveC),
	}
}

func (i InMemQueue) Broadcast(m *tunnel.QueueMessage) error {
	i.receivec <- m
	return nil
}

func (i InMemQueue) Listen() (tunnel.ReceiveC, error) {
	return i.receivec, nil
}

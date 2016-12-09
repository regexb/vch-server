package tunnel

/*
Queue Interface
---------------

The Queue interface describes a system for broadcasting
and receiving messages from any other running vchd
process. It acts as a message broker for allowing all
processes to know about incoming messages and
allows the process with the connected client to handle
sending the message down to the client.
*/

type NLPResponse struct {
	Body []byte `json:"body"`
}

type QueueMessage struct {
	NLPResponse NLPResponse
}

type ReceiveC chan *QueueMessage

type Queue interface {
	Broadcast(message *QueueMessage) error
	Listen() (ReceiveC, error)
}

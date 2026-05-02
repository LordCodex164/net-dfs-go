package p2p

import "net"

const (
	IncomingStream = 0x2
	IncomingMessage = 0x1
)

type RPC struct {
	From 	string
	Payload []byte
	Stream	bool
}

type HandshakeFunc func(Peer) error

type Peer interface {
	net.Conn
	Send(b []byte) (error)
	CloseStream()
}
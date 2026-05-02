package p2p

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

type TcpTransportOpts struct {
	Listen_address		string
	OnPeer			HandlePeer
}

type TcpTransport struct {
	TcpTransportOpts
	rpcch    		chan RPC
	mu       		sync.Mutex
	Listener		net.Listener
	Decoder			DefaultDecoder
}

type HandlePeer func(Peer) error

//create mock HandlePeer function

func NopHandlePeer(Peer) error {return nil}

type TcpPeer struct {
	net.Conn
	outbound 	bool
	wg 			*sync.WaitGroup
}

func NewTcpPeer(conn net.Conn, outbound bool) *TcpPeer {
	return &TcpPeer{
		Conn: conn,
		outbound: false,
		wg: &sync.WaitGroup{},
	}
}

func (tp *TcpPeer) Send(b []byte) error {
	_, err := tp.Conn.Write(b)
	return err
}

func (tP *TcpPeer) RemoteAdr() string {
	return tP.Conn.RemoteAddr().String()
}

func (tp *TcpPeer) CloseStream() {
	tp.wg.Done()
}

func NewTcpTransport(opts TcpTransportOpts) *TcpTransport {
	return &TcpTransport{
		TcpTransportOpts: opts,
		rpcch: make(chan RPC, 1024),
		Decoder: DefaultDecoder{},
	}
}

func (t *TcpTransport) Listen_Address() string {
	return t.Listen_address
}

func (t *TcpTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	go t.handleConn(conn, true)
	return nil
}

func (t *TcpTransport) Consume() <-chan RPC {
	return t.rpcch
}

func (t *TcpTransport) CloseChannel() {
	close(t.rpcch)
}


func (t *TcpTransport) ListenAndAccept() error {
	var err error

	t.Listener, err = net.Listen("tcp", t.Listen_address)
	if err != nil {
		return err
	}
	//this works after connection 
	go t.acceptHandShakeLoop()
	log.Printf("TCP transport listening on port: %s\n", t.Listen_address)
	return nil
}

func (t *TcpTransport) acceptHandShakeLoop() {

	for {
		conn, err := t.Listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}
		if err != nil {
			fmt.Println("tcp accept error:", err, conn)
		}
		//handle connections
		go t.handleConn(conn, false)
	}

}

func (t *TcpTransport) handleConn(conn net.Conn, outbound bool) {

	Peer := NewTcpPeer(conn, outbound)

	fmt.Printf("peer at remote address %+v and listener address %+s\n", Peer.RemoteAddr().String(), t.Listen_address)

	if err := t.OnPeer(Peer); err != nil {
		fmt.Println("peer error", err)
		return
	}

	for {
		rpc := RPC{}
		err := t.Decoder.Decode(conn, &rpc)
		if err != nil {
			return
		}
		rpc.From = conn.RemoteAddr().String()
		if rpc.Stream {
			Peer.wg.Add(1)
			fmt.Printf("%+v incoming stream...\n", conn.RemoteAddr().String())
			Peer.wg.Wait()
			fmt.Println("stream closed")
			continue
		}
		t.rpcch <- rpc
		fmt.Println("ssss")
		
	}
}

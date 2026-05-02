package main

import (
	"bytes"
	"fmt"
	//"io/ioutil"
	//"fmt"
	"io"
	"log"
	"time"
	"github.com/LordCodex164/net-protocol/p2p"
)

func makeServer(listenAddr string, nodes ...string) *Server {
	tcptransportOpts := p2p.TcpTransportOpts{
		Listen_address:    listenAddr,
		//HandshakeFunc: p2p.NOPHandshakeFunc, 
	}
	tcpTransport := p2p.NewTcpTransport(tcptransportOpts)

	fileServerOpts := ServerOpts{
		EncKey:           	newEncryptionKey(),
		StorageRoot:       	listenAddr + "_network",
		PathTransformFunc: 	DefaultPathTransformFunc,
		Transport:         	tcpTransport,
		Server_nodes:      	nodes,
		serverId: 			generateID(),
	}

	s := NewServer(fileServerOpts)

	tcpTransport.OnPeer = s.handlePeer

	return s
}

func main() {
	s1 := makeServer(":3030", "")
	s2 := makeServer(":7000", ":3030")
	time.Sleep(2 * time.Second)
	go func() { log.Fatal(s1.Start()) }()

	time.Sleep(2 * time.Second)
	go func() { log.Fatal(s2.Start()) }()

	for i := 0; i < 3; i++ {
		key := fmt.Sprintf("picture_%d", i)
		data := bytes.NewReader([]byte("hello world"))
		s2.Store(key, data)
		if err := s2.Storage.Delete(s2.serverId, key); err != nil {
			log.Fatal(err)
		}
		r, err := s2.Get(key)
		if err != nil {
			log.Fatal(err)
		}
		b, err := io.ReadAll(r)
		fmt.Println("r", string(b))
	}
	select{}
}

package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/LordCodex164/net-protocol/p2p"
)


type ServerOpts struct {
	Transport 			p2p.Transport
	EncKey    			[]byte
	Server_nodes		[]string
	ListenAddr			string
	PathTransformFunc	PathTransformFunc
	serverId			string
	StorageRoot			string
}

type Server struct {
	ServerOpts
	serverCh			chan struct{}
	peers               map[string]p2p.Peer
	Storage             *Store
	peerLock			sync.Mutex
}

func NewServer(opts ServerOpts) *Server {
	storeOpts := StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}

	return &Server{
		ServerOpts: opts,
		Storage:          NewStore(storeOpts),
		serverCh:        make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

func (s *Server) broadcast(msg *Message) error {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers {
		peer.Send([]byte{p2p.IncomingMessage})
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}

	return nil
}

type Message struct {
	Payload any
}


type MessageStoreFile struct {
	ID 		string
	Key 	string
	Size	int64
}

type MessageGetFile struct {
	ID		string
	Key	string
}

func (s *Server) Store(key string, r io.Reader) error {
	var (
		buf = new(bytes.Buffer)
		teeReader = io.TeeReader(r, buf)
	)

	n, err := s.Storage.Write(s.serverId, key, teeReader)

	if err != nil {
		return err
	}

	// //broadcast the message id, key and size over the network
	// //then the filebuffer

	msg := Message{
		Payload: MessageStoreFile{
			ID: s.serverId,
			Key: key,
			Size: n + 16,
		},
	}

	fmt.Printf("msg:%+v\n", msg)

	if err = s.broadcast(&msg); err != nil {
		return err
	}

	fmt.Println("successful broadcast over the network")

	time.Sleep(time.Millisecond * 5)

	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	mw.Write([]byte{p2p.IncomingStream})
	size, err := copyEncrypt(s.EncKey, buf, mw)
	if err != nil {
		return err
	}

	fmt.Printf("[%s] received and written (%d) bytes to disk\n", s.Transport.Listen_Address(), size)
	return nil
}

func (s * Server) Get(key string) (io.Reader, error) {

	if s.Storage.Has(s.serverId, key) {
		fmt.Printf("serving file disk with the key %v\n", key)
		_, r, err := s.Storage.Read(s.serverId, key)
		if rc, ok := r.(io.ReadCloser); ok {
			rc.Close()
		}
		if err != nil {
			return nil, err
		}
		return r, nil
	}

	fmt.Printf("[%s] file not found, need to serve file from the network\n", s.Transport.Listen_Address())

	msg := Message{
		Payload: MessageGetFile{
			ID: s.serverId,
			Key: key,
		},
	}

	if err := s.broadcast(&msg); err != nil {
		return nil, err
	}

	for _, peer := range(s.peers) {
		//get file size
		var fileSize int64
		//read the binary structured data of the peer (which is the file size) into the data(fileSize)
		binary.Read(peer, binary.LittleEndian, &fileSize)
		_, err := s.Storage.WriteDecrypt(s.EncKey, s.serverId, key, io.LimitReader(peer, fileSize))
		if err != nil {
			return nil, err
		}
		peer.CloseStream()
	}
	
	_, r, err := s.Storage.Read(s.serverId, key)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (s *Server) Close(){
	close(s.serverCh)
}

func (s *Server) handlePeer(peer p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()
	s.peers[peer.RemoteAddr().String()] = peer
	fmt.Println("server peer", s.peers)
	return nil
}


func (s *Server) loop() {
	defer func() {
		log.Println("file server stopped due to error or user quit action")
	}()

	for {
		select {
		case rpc := <-s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Println("decoding error: ", err)
			}
			if err := s.handleMessage(&msg, rpc.From); err != nil {
				log.Println("handle message error: ", err)
			}

		case <-s.serverCh:
			return
		}
	}
}

func (s *Server) handleMessage(msg *Message, from string) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return s.handleMessageStoreFile(from, v)
	case MessageGetFile:
		return s.handleMessageGetFile(from, v)
	}
	return nil
}

func (s *Server) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	fmt.Println("vvv")
	
	peer, ok := s.peers[from]

	if !ok {
		fmt.Println("the peer does not exist in the nwtwork")
	}

	n, err := s.Storage.Write(msg.ID, msg.Key, io.LimitReader(peer, msg.Size))

	if err != nil {
		return err
	}

	fmt.Printf("[%s] written %d bytes to disk\n", s.Transport.Listen_Address(), n)

	
	peer.CloseStream()

	return nil
}


func (s *Server) handleMessageGetFile(from string, msg MessageGetFile) error {

	//check if the file exists in the disk

	if !s.Storage.Has(msg.ID, msg.Key) {
		return fmt.Errorf("file can not found in the network [%s]", s.Transport.Listen_Address())
	}

	//it it exists 

	fs, f, err := s.Storage.Read(msg.ID, msg.Key)

	if err != nil {
		return err
	}

	//implement a closer for the file since it implements the reader interface

	peer := s.peers[from]

	//before sending data over the network

	//send the type of message to know we are streaming
	peer.Send([]byte{p2p.IncomingStream})
	binary.Write(peer, binary.LittleEndian, fs)
	n, err := io.Copy(peer, f)

	if rc, ok := f.(io.ReadCloser); ok {
		fmt.Println("closing the reader")
		rc.Close()
	}

	if err != nil {
		return err
	}

	time.Sleep(50 * time.Millisecond)

	fmt.Printf("written [%v] bytes over the network", n)

	//peer.CloseStream()
	return nil

}


func (s *Server) boostrap_nodes() error {
	for _, addr := range(s.Server_nodes) {
		if len(addr) == 0 {
			continue
		}
		go func(addr string) {
			fmt.Println("addr", addr)
			if err := s.Transport.Dial(addr); err != nil {
				fmt.Println("err", err)
			}
		fmt.Printf("port %+v dialing a connection to address %s\n", s.Transport.Listen_Address(), addr)
		}(addr)
	}
	return nil
}

func (s *Server) Start() error {
	fmt.Printf("[%s] starting fileserver...\n", s.Transport.Listen_Address())
	err := s.Transport.ListenAndAccept()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("listening on port %s \n", s.Transport.Listen_Address())
	s.boostrap_nodes()
	s.loop()
	return nil
}


//to-do let broadcast message of type messageStoreFile


func init() {
	gob.Register(MessageStoreFile{})
	gob.Register(MessageGetFile{})
}
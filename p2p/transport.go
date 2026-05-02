package p2p


type Transport interface {
	ListenAndAccept()	error
	Dial(string)		error
	Listen_Address()	string
	Consume()			<-chan RPC
}
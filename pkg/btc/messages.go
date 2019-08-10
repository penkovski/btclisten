package btc

// NetAddr represents a network address as per the Bitcoin documentation at
// https://en.bitcoin.it/wiki/Protocol_documentation#Network_address
type NetAddr struct {
	Time     uint32
	Services uint64
	IP       [16]byte
	Port     uint16
}

// Message describes the structure of a bitcoin protocol message.
// https://en.bitcoin.it/wiki/Protocol_documentation#Message_structure
type Message struct {
	Magic    uint32
	Command  [12]byte
	Length   uint32
	Checksum uint32
	Payload  []byte
}

// MsgVersion is the initial version message that is exchanged between a
// connecting node and its peer.
// https://en.bitcoin.it/wiki/Protocol_documentation#version
type MsgVersion struct {
	Version     uint32
	Services    uint64
	Timestamp   uint64
	AddrRecv    NetAddr
	AddrFrom    NetAddr
	Nonce       uint64
	UserAgent   string
	StartHeight int
	Relay       bool
}

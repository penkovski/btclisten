package btc

import (
	"bytes"
	"encoding/binary"
)

// NetAddr represents a network address as per the Bitcoin documentation at
// https://en.bitcoin.it/wiki/Protocol_documentation#Network_address
type NetAddr struct {
	Time     uint32
	Services uint64
	IP       [16]byte
	Port     uint16
}

// Serialize the network address before sending over the network.
//
// Most integers are encoded in little endian.
// Only IP or port number are encoded big endian.
// https://en.bitcoin.it/wiki/Protocol_documentation#Common_structures
func (addr NetAddr) Serialize() []byte {
	var buf bytes.Buffer

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, addr.Services)
	buf.Write(b)

	// most things are encoded in LittleEndian, but
	// IP and Port are encoded in BigEndian...
	buf.Write(addr.IP[:])

	binary.Write(&buf, binary.BigEndian, addr.Port)

	return buf.Bytes()
}

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

// MsgEnvelope represents the structure of a bitcoin protocol message.
// https://en.bitcoin.it/wiki/Protocol_documentation#Message_structure
type MsgEnvelope struct {
	Magic    uint32
	Command  [12]byte
	Length   uint32
	Checksum [4]byte
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
	StartHeight int32
	Relay       bool
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

// Serialize the protocol message envelope and add the payload
// to the serialized bytes slice.
func (m MsgEnvelope) Serialize() []byte {
	var buf bytes.Buffer

	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, m.Magic)
	buf.Write(b)

	buf.Write(m.Command[:])

	b = make([]byte, 4)
	binary.LittleEndian.PutUint32(b, m.Length)
	buf.Write(b)

	buf.Write(m.Checksum[:])
	buf.Write(m.Payload)

	return buf.Bytes()
}

// Serialize version protocol message. This is the
// payload of the MsgEnvelope.
func (mv MsgVersion) Serialize() (data []byte) {
	var buf bytes.Buffer

	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, mv.Version)
	buf.Write(b)

	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, mv.Services)
	buf.Write(b)

	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, mv.Timestamp)
	buf.Write(b)

	buf.Write(mv.AddrRecv.Serialize())
	buf.Write(mv.AddrFrom.Serialize())

	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, mv.Nonce)
	buf.Write(b)

	// TODO(penkovski): write user agent (optional)

	b = make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(mv.StartHeight))
	buf.Write(b)

	// TODO(penkovski): write Relay (optional)

	return buf.Bytes()
}

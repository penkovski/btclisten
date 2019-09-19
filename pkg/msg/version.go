package msg

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

// Version is the initial version message that is exchanged between a
// connecting node and its peer.
// https://en.bitcoin.it/wiki/Protocol_documentation#version
type Version struct {
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

// NewVersion returns a version message with IP and Port of the
// connecting peer, which is intended to be send to another peer as
// part of the initial peer-to-peer connection handshake.
func NewVersion(peerIP [16]byte, peerPort uint16) *Version {
	msgver := &Version{
		Version:   ProtocolVersion,
		Services:  1,
		Timestamp: uint64(time.Now().Unix()),
		AddrRecv: NetAddr{
			Services: 1,
			IP:       peerIP,
			Port:     peerPort,
		},
		AddrFrom: NetAddr{
			Services: 1,
			IP:       [16]byte{},
			Port:     0,
		},
		Nonce:       randomNonce(),
		UserAgent:   "github.com/penkovski/btclisten",
		StartHeight: 0,
		Relay:       false,
	}

	return msgver
}

// Serialize version protocol message.
func (v *Version) Serialize() (data []byte) {
	var buf bytes.Buffer

	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v.Version)
	buf.Write(b)

	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, v.Services)
	buf.Write(b)

	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, v.Timestamp)
	buf.Write(b)

	buf.Write(v.AddrRecv.Serialize(false))
	buf.Write(v.AddrFrom.Serialize(false))

	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, v.Nonce)
	buf.Write(b)

	// TODO(penkovski): write user agent (optional)
	buf.Write([]byte{0x00})

	b = make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(v.StartHeight))
	buf.Write(b)

	// TODO(penkovski): write Relay (optional)

	return buf.Bytes()
}

func (v *Version) Deserialize(r io.Reader) error {
	buf, ok := r.(*bytes.Buffer)
	if !ok {
		return fmt.Errorf("[Version.Deserialize] reader is not a *bytes.Buffer")
	}

	// protocol version
	b := make([]byte, 4)
	if _, err := io.ReadFull(buf, b); err != nil {
		return err
	}
	v.Version = binary.LittleEndian.Uint32(b)

	// services
	b = make([]byte, 8)
	if _, err := io.ReadFull(buf, b); err != nil {
		return err
	}
	v.Services = binary.LittleEndian.Uint64(b)

	// timestamp
	b = make([]byte, 8)
	if _, err := io.ReadFull(buf, b); err != nil {
		return err
	}
	v.Timestamp = binary.LittleEndian.Uint64(b)

	// addr peer
	v.AddrRecv.Deserialize(buf)

	// Protocol versions >= 106 added a from address, nonce, and user agent
	// field and they are only considered present if there are bytes
	// remaining in the message.
	if buf.Len() > 0 {
		v.AddrFrom.Deserialize(buf)
	}

	// nonce
	if buf.Len() > 0 {
		b = make([]byte, 8)
		if _, err := io.ReadFull(buf, b); err != nil {
			return err
		}
		v.Nonce = binary.LittleEndian.Uint64(b)
	}

	// user agent
	// TODO(penkovski): read user agent

	// last block height
	// TODO(penkovski): read last block height

	// relay
	// TODO(penkovski): read relay flag

	return nil
}

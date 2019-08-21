package btc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

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

func NewMsgVersion(peerIP [16]byte, peerPort uint16) *MsgVersion {
	msgver := &MsgVersion{
		Version:   Version,
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
		StartHeight: -1,
		Relay:       false,
	}

	return msgver
}

// Serialize version protocol message. This is the
// payload of the Msg.
func (mv *MsgVersion) Serialize() (data []byte) {
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

	buf.Write(mv.AddrRecv.Serialize(false))
	buf.Write(mv.AddrFrom.Serialize(false))

	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, mv.Nonce)
	buf.Write(b)

	// TODO(penkovski): write user agent (optional)
	buf.Write([]byte{0x00})
	//buf.Write([]byte{0x0F, 0x2F, 0x53, 0x61, 0x74, 0x6F, 0x73, 0x68, 0x69, 0x3A, 0x30, 0x2E, 0x37, 0x2E, 0x32, 0x2F})

	b = make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(mv.StartHeight))
	buf.Write(b)

	// TODO(penkovski): write Relay (optional)

	return buf.Bytes()
}

func (mv *MsgVersion) Deserialize(r io.Reader) error {
	buf, ok := r.(*bytes.Buffer)
	if !ok {
		return fmt.Errorf("[MsgVersion.Deserialize] reader is not a *bytes.Buffer")
	}

	// protocol version
	b := make([]byte, 4)
	if _, err := io.ReadFull(buf, b); err != nil {
		return err
	}
	mv.Version = binary.LittleEndian.Uint32(b)

	// services
	b = make([]byte, 8)
	if _, err := io.ReadFull(buf, b); err != nil {
		return err
	}
	mv.Services = binary.LittleEndian.Uint64(b)

	// timestamp
	b = make([]byte, 8)
	if _, err := io.ReadFull(buf, b); err != nil {
		return err
	}
	mv.Timestamp = binary.LittleEndian.Uint64(b)

	// addr peer
	mv.AddrRecv.Deserialize(buf)

	// Protocol versions >= 106 added a from address, nonce, and user agent
	// field and they are only considered present if there are bytes
	// remaining in the message.
	if buf.Len() > 0 {
		mv.AddrFrom.Deserialize(buf)
	}

	// nonce
	if buf.Len() > 0 {
		b = make([]byte, 8)
		if _, err := io.ReadFull(buf, b); err != nil {
			return err
		}
		mv.Nonce = binary.LittleEndian.Uint64(b)
	}

	// user agent
	// TODO(penkovski): read user agent

	// last block height
	// TODO(penkovski): read last block height

	// relay
	// TODO(penkovski): read relay flag

	return nil
}

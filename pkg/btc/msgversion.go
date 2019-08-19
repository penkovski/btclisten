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

func NewMsgVersion(peerIP [16]byte, peerPort uint16) MsgVersion {
	msgver := MsgVersion{
		Version:   Version,
		Services:  1,
		Timestamp: uint64(time.Now().Unix()),
		AddrRecv: NetAddr{
			IP:   peerIP,
			Port: peerPort,
		},
		AddrFrom: NetAddr{
			IP:   [16]byte{},
			Port: 0,
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

func (mv MsgVersion) Deserialize(r io.Reader) error {
	buf, ok := r.(*bytes.Buffer)
	if !ok {
		return fmt.Errorf("[MsgVersion.Deserialize] reader is not a *bytes.Buffer")
	}

	// protocol version

	// services

	// timestamp

	// addr peer

	// addr local

	// nonce

	// user agent

	// last block height

	// relay

	return nil
}

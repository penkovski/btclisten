package btc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
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

func (addr NetAddr) Deserialize(r io.Reader) error {
	buf, ok := r.(*bytes.Buffer)
	if !ok {
		return fmt.Errorf("[NetAddr.Deserialize] reader is not a *bytes.Buffer")
	}

	// time
	b := make([]byte, 4)
	if _, err := io.ReadFull(buf, b); err != nil {
		return err
	}
	addr.Time = binary.LittleEndian.Uint32(b)

	// services
	b = make([]byte, 8)
	if _, err := io.ReadFull(buf, b); err != nil {
		return err
	}
	addr.Services = binary.LittleEndian.Uint64(b)

	// ip
	if _, err := io.ReadFull(buf, addr.IP[:]); err != nil {
		return err
	}

	// port
	b = make([]byte, 2)
	if _, err := io.ReadFull(buf, b); err != nil {
		return err
	}
	addr.Port = binary.BigEndian.Uint16(b)

	return nil
}

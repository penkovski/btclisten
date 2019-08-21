package btc

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
)

// MaxMessagePayload is the maximum size of the message payload.
const MaxMessagePayload = (1024 * 1024 * 16) // 16MB

// Msg represents the structure of a bitcoin protocol message.
// https://en.bitcoin.it/wiki/Protocol_documentation#Message_structure
type Msg struct {
	Magic    uint32
	Command  [12]byte
	Length   uint32
	Checksum [4]byte
	Payload  []byte
}

func NewMsg(magic uint32, command string, payload []byte) *Msg {
	var cmd [12]byte
	copy(cmd[:], command)

	msg := &Msg{
		Magic:   magic,
		Command: cmd,
		Length:  uint32(len(payload)),
	}

	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])
	copy(msg.Checksum[:], second[0:4])
	msg.Payload = payload

	return msg
}

// Serialize the message and the payload to the
// bytes slice returned in the response.
func (m *Msg) Serialize() []byte {
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

func (m *Msg) Deserialize(r io.Reader) error {
	var headerBytes [24]byte
	_, err := io.ReadFull(r, headerBytes[:])
	if err != nil {
		return fmt.Errorf("error reading header bytes: %v", err)
	}

	header := bytes.NewReader(headerBytes[:])

	// read Magic
	buf := make([]byte, 4)
	if _, err := io.ReadFull(header, buf); err != nil {
		buf = nil
		return fmt.Errorf("error reading Magic: %v", err)
	}
	m.Magic = binary.LittleEndian.Uint32(buf)

	// read Command
	buf = make([]byte, 12)
	if _, err := io.ReadFull(header, buf); err != nil {
		buf = nil
		return fmt.Errorf("error reading command: %v", err)
	}
	copy(m.Command[:], buf[:])

	// read Payload Length
	buf = make([]byte, 4)
	if _, err := io.ReadFull(header, buf); err != nil {
		buf = nil
		return fmt.Errorf("error reading payload length: %v", err)
	}
	m.Length = binary.LittleEndian.Uint32(buf)
	if m.Length > MaxMessagePayload {
		return fmt.Errorf("error: message payload length is too big")
	}

	// read Checksum
	buf = make([]byte, 4)
	if _, err := io.ReadFull(header, buf); err != nil {
		buf = nil
		return fmt.Errorf("error reading checksum: %v", err)
	}
	copy(m.Checksum[:], buf[:])

	// read the Payload
	payload := make([]byte, m.Length)
	_, err = io.ReadFull(r, payload)
	if err != nil {
		return fmt.Errorf("error reading payload: %v", err)
	}
	m.Payload = payload

	return nil
}

package msg

import (
	"crypto/rand"
	"encoding/binary"
)

func randomNonce() uint64 {
	buf := make([]byte, 8)
	_, err := rand.Read(buf)
	if err != nil {
		return 0
	}
	return binary.BigEndian.Uint64(buf)
}

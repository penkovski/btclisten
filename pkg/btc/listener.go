package btc

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
)

const (
	Version = 31800

	MagicMainNet  = 0xD9B4BEF9
	MagicTestNet  = 0xDAB5BFFA
	MagicTestNet3 = 0x0709110B
	MagicNameCoin = 0xFEB4BEF9
)

type Listener struct {
	peerIP   [16]byte
	peerPort uint16

	conn net.Conn
	quit chan bool
}

func NewListener(conn net.Conn, done chan bool) (*Listener, error) {
	// extract host and port
	ip, port, err := addrIpPort(conn.RemoteAddr())
	if err != nil {
		return nil, err
	}

	var peerIP [16]byte
	copy(peerIP[:], ip)

	l := &Listener{
		conn:     conn,
		quit:     done,
		peerIP:   peerIP,
		peerPort: port,
	}

	return l, nil
}

func (l *Listener) Start() {
	defer func() { l.quit <- true }()

	err := l.Handshake()
	if err != nil {
		return
	}

	l.readVerAck()

	go l.Listen(l.conn)
}

// Handshake initiates the exchange of version messages between
// the connecting client/node and its peer.
//
// When a node creates an outgoing connection, it will immediately
// advertise its version. The remote node will respond with
// its version. No further communication is possible until both
// peers have exchanged their version.
// https://en.bitcoin.it/wiki/Protocol_documentation
func (l *Listener) Handshake() error {
	msgver := MsgVersion{
		Version:   Version,
		Services:  1,
		Timestamp: uint64(time.Now().Unix()),
		AddrRecv: NetAddr{
			IP:   l.peerIP,
			Port: l.peerPort,
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

	payload := msgver.serialize()

	var cmd [12]byte
	copy(cmd[:], "version")

	msg := MsgEnvelope{
		Magic:   MagicMainNet,
		Command: cmd,
		Length:  uint32(len(payload)),
	}

	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])
	copy(msg.Checksum[:], second[0:4])
	msg.Payload = payload

	_, err := l.conn.Write(msg.serialize())
	return err
}

func (l *Listener) readVerAck() error {
	var headerBytes [24]byte
	_, err := io.ReadFull(l.conn, headerBytes[:])
	if err != nil {
		return err
	}

	fmt.Println(headerBytes)

	return nil
}

func (l *Listener) ping() {

}

func (l *Listener) Listen(conn net.Conn) {
	const maxBufSize = 1 << 16
	buf := make([]byte, maxBufSize)
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(buf, maxBufSize)
	for scanner.Scan() {
		message := scanner.Text()
		fmt.Println(message)
	}
}

func addrIpPort(addr net.Addr) (ip string, port uint16, err error) {
	if tcpAddr, ok := addr.(*net.TCPAddr); ok {
		ip = tcpAddr.IP.String()
		port = uint16(tcpAddr.Port)
		return ip, port, nil
	}

	// For the most part, addr should be one of the two above cases, but
	// to be safe, fall back to trying to parse the information from the
	// address string as a last resort.
	host, portStr, err := net.SplitHostPort(addr.String())
	if err != nil {
		return "", 0, err
	}
	ipAddr := net.ParseIP(host)
	p, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return "", 0, err
	}
	ip = ipAddr.String()
	port = uint16(p)

	return ip, port, nil
}

func randomNonce() uint64 {
	buf := make([]byte, 8)
	_, err := rand.Read(buf)
	if err != nil {
		return 0
	}
	return binary.BigEndian.Uint64(buf)
}

package btc

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

const (
	Version = 60002

	MagicMainNet  = 0xD9B4BEF9
	MagicTestNet  = 0xDAB5BFFA
	MagicTestNet3 = 0x0709110B
	MagicNameCoin = 0xFEB4BEF9
)

type Listener struct {
	peerIP   [16]byte
	peerPort uint16

	conn net.Conn
}

func NewListener(conn net.Conn) (*Listener, error) {
	// extract host and port
	ip, port, err := addrIpPort(conn.RemoteAddr())
	if err != nil {
		return nil, err
	}

	var peerIP [16]byte
	copy(peerIP[:], ip)

	l := &Listener{
		conn:     conn,
		peerIP:   peerIP,
		peerPort: port,
	}

	return l, nil
}

// Start should be executed in its own go routine
// if you don't want to block when calling it.
func (l *Listener) Start(quit chan struct{}) {
	defer func() { quit <- struct{}{} }()

	err := l.Handshake()
	if err != nil {
		fmt.Println(err)
		return
	}

	l.Listen(l.conn)
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
	fmt.Println("initiate handshake...")

	// send version message
	msgver := NewMsgVersion(l.peerIP, l.peerPort)
	payload := msgver.Serialize()
	msg := NewMsg(MagicMainNet, "version", payload)
	if _, err := l.conn.Write(msg.Serialize()); err != nil {
		return err
	}

	// read peer version response
	verack, err := l.readVerAck()
	if err != nil {
		return err
	}
	if !bytes.EqualFold(verack.Command[:], []byte("version")) {
		return fmt.Errorf("expected version command, but received: %v", verack.Command)
	}

	// validate Checksum
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])
	if !bytes.Equal(verack.Checksum[0:4], second[0:4]) {
		return fmt.Errorf("invalid checksum: %v, expected = %v", verack.Checksum, second[0:4])
	}

	peerMsgVersion := &MsgVersion{}
	buf := bytes.NewBuffer(payload)
	err = peerMsgVersion.Deserialize(buf)
	if err != nil {
		return fmt.Errorf("error deserializing version message payload: %v", err)
	}

	fmt.Println("peer version =", peerMsgVersion)

	return err
}

func (l *Listener) readVerAck() (Msg, error) {
	msg := Msg{}
	err := msg.Deserialize(l.conn)
	if err != nil {
		return Msg{}, err
	}
	return msg, err
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

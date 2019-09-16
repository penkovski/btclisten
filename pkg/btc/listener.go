package btc

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

const (
	Version = 70015

	MagicMainNet  = 0xD9B4BEF9
	MagicTestNet  = 0xDAB5BFFA
	MagicTestNet3 = 0x0709110B
	MagicNameCoin = 0xFEB4BEF9
)

type Listener struct {
	peerIP   [16]byte
	peerPort uint16

	stop chan struct{}

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
		stop:     make(chan struct{}),
	}

	return l, nil
}

// Start should be executed in its own go routine
// if you don't want to block when calling it.
func (l *Listener) Start(notifyQuit chan struct{}) {
	defer func() { notifyQuit <- struct{}{} }()

	err := l.Handshake()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("listening...")

	l.Listen(l.conn)
}

func (l *Listener) Stop() {
	l.stop <- struct{}{}
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
	msg := NewMsgEnvelope(MagicMainNet, "version", payload)
	if _, err := l.conn.Write(msg.Serialize()); err != nil {
		return err
	}
	fmt.Println(" - send version")

	msg = &MsgEnvelope{}
	err := msg.Deserialize(l.conn)
	if err != nil {
		return err
	}

	// check that it's a version message
	cmdstr := string(bytes.TrimRight(msg.Command[:], string(0)))
	if cmdstr != "version" {
		return fmt.Errorf("expected version command, but received: %v", msg.Command)
	}

	// validate Checksum
	first := sha256.Sum256(msg.Payload)
	second := sha256.Sum256(first[:])
	if !bytes.Equal(msg.Checksum[0:4], second[0:4]) {
		return fmt.Errorf("invalid checksum: %v, expected = %v", msg.Checksum, second[0:4])
	}

	// deserialize version message payload
	receivedMsgVersion := &MsgVersion{}
	buf := bytes.NewBuffer(msg.Payload)
	err = receivedMsgVersion.Deserialize(buf)
	if err != nil {
		return fmt.Errorf("error deserializing version message payload: %v", err)
	}
	fmt.Printf(" - received peer version: %+v\n", receivedMsgVersion)

	// send version acknowledgement
	msgVerAck := NewMsgEnvelope(MagicMainNet, "verack", nil)
	if _, err := l.conn.Write(msgVerAck.Serialize()); err != nil {
		return err
	}
	fmt.Printf(" - send verack: %+v\n", msgVerAck.Serialize())

	msg = &MsgEnvelope{}
	err = msg.Deserialize(l.conn)
	if err != nil {
		return err
	}

	// check that it's a verack message
	cmdstr = string(bytes.TrimRight(msg.Command[:], string(0)))
	if cmdstr != "verack" {
		return fmt.Errorf("expected version command, but received: %v", msg.Command)
	}
	fmt.Printf(" - received verack: %+v\n", msg)

	return nil
}

func (l *Listener) Listen(conn net.Conn) {
	// listen for messages
	for {
		select {
		case <-l.stop:
			return
		default:
			msg := &MsgEnvelope{}
			err := msg.Deserialize(l.conn)
			if err != nil {
				return
			}

			fmt.Println("--- received message ---")
			fmt.Println(msg)
			fmt.Printf("magic = %x\n", msg.Magic)
			fmt.Printf("command = %s\n", msg.Command)
			fmt.Printf("length = %d\n", msg.Length)
			fmt.Printf("checksum = %x\n", msg.Checksum)
			fmt.Printf("payload = %x\n", msg.Payload)
		}
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

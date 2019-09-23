package btc

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"net"

	"github.com/penkovski/btclisten/pkg/msg"
)

const (
	MagicMainNet  = 0xD9B4BEF9
	MagicTestNet  = 0xDAB5BFFA
	MagicTestNet3 = 0x0709110B
	MagicNameCoin = 0xFEB4BEF9
)

type Listener struct {
	seedNode *Peer

	stop chan struct{}

	conn net.Conn
}

func NewListener(conn net.Conn) (*Listener, error) {
	seedNode, err := NewPeer(conn)
	if err != nil {
		return nil, err
	}

	l := &Listener{
		conn:     conn,
		seedNode: seedNode,
		stop:     make(chan struct{}),
	}

	return l, nil
}

// Start should be executed in its own go routine
// if you don't want to block when calling it.
func (l *Listener) Start(notifyDone chan struct{}) {
	defer func() { notifyDone <- struct{}{} }()

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
	msgver := msg.NewVersion(l.seedNode.IP, l.seedNode.Port)
	payload := msgver.Serialize()
	m := msg.New(MagicMainNet, "version", payload)
	if _, err := l.conn.Write(m.Serialize()); err != nil {
		return err
	}
	fmt.Println(" - send version")

	m = &msg.Envelope{}
	err := m.Deserialize(l.conn)
	if err != nil {
		return err
	}

	// check that it's a version message
	cmdstr := string(bytes.TrimRight(m.Command[:], string(0)))
	if cmdstr != "version" {
		return fmt.Errorf("expected version command, but received: %v", m.Command)
	}

	// validate Checksum
	first := sha256.Sum256(m.Payload)
	second := sha256.Sum256(first[:])
	if !bytes.Equal(m.Checksum[0:4], second[0:4]) {
		return fmt.Errorf("invalid checksum: %v, expected = %v", m.Checksum, second[0:4])
	}

	// deserialize version message payload
	receivedMsgVersion := &msg.Version{}
	buf := bytes.NewBuffer(m.Payload)
	err = receivedMsgVersion.Deserialize(buf)
	if err != nil {
		return fmt.Errorf("error deserializing version message payload: %v", err)
	}
	fmt.Printf(" - received peer version: %+v\n", receivedMsgVersion)

	// send version acknowledgement
	msgVerAck := msg.New(MagicMainNet, "verack", nil)
	if _, err := l.conn.Write(msgVerAck.Serialize()); err != nil {
		return err
	}
	fmt.Printf(" - send verack: %+v\n", msgVerAck.Serialize())

	m = &msg.Envelope{}
	err = m.Deserialize(l.conn)
	if err != nil {
		return err
	}

	// check that it's a verack message
	cmdstr = string(bytes.TrimRight(m.Command[:], string(0)))
	if cmdstr != "verack" {
		return fmt.Errorf("expected version command, but received: %v", m.Command)
	}
	fmt.Printf(" - received verack: %+v\n", m)

	return nil
}

func (l *Listener) Listen(conn net.Conn) {
	// listen for messages
	for {
		select {
		case <-l.stop:
			return
		default:
			msg := &msg.Envelope{}
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

package btc

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"log"
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

// Run should be executed in its own go routine
// if you don't want to block when calling it.
func (l *Listener) Run(notifyDone chan struct{}) {
	// upon exiting the function, the listener
	// will notify that it's quitting and the
	// program will terminate itself
	defer func() { notifyDone <- struct{}{} }()

	err := l.Handshake()
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("listening...")

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
	log.Println("initiate handshake...")

	// send version message
	msgver := msg.NewVersion(l.seedNode.IP, l.seedNode.Port)
	payload := msgver.Serialize()
	m := msg.New(MagicMainNet, "version", payload)
	if _, err := l.conn.Write(m.Serialize()); err != nil {
		return fmt.Errorf("error sending version message: %v", err)
	}
	log.Println(" ✓ send version")

	m = &msg.Message{}
	err := m.Deserialize(l.conn)
	if err != nil {
		return err
	}
	log.Println(" ✓ received peer version")

	// check that it's a version message
	cmdstr := string(bytes.TrimRight(m.Command[:], string(0)))
	if cmdstr != "version" {
		return fmt.Errorf("expected version message, but received: %v", m.Command)
	}

	// validate Checksum
	first := sha256.Sum256(m.Payload)
	second := sha256.Sum256(first[:])
	if !bytes.Equal(m.Checksum[0:4], second[0:4]) {
		return fmt.Errorf("invalid checksum: %v, expected = %v", m.Checksum, second[0:4])
	}
	log.Println(" ✓ validate peer version")

	// deserialize version message payload
	receivedMsgVersion := &msg.Version{}
	buf := bytes.NewBuffer(m.Payload)
	err = receivedMsgVersion.Deserialize(buf)
	if err != nil {
		return fmt.Errorf("error deserializing version message payload: %v", err)
	}

	// send version acknowledgement
	msgVerAck := msg.New(MagicMainNet, "verack", nil)
	if _, err := l.conn.Write(msgVerAck.Serialize()); err != nil {
		return err
	}
	log.Println(" ✓ send verack")

	m = &msg.Message{}
	err = m.Deserialize(l.conn)
	if err != nil {
		return err
	}

	// check that it's a verack message
	cmdstr = string(bytes.TrimRight(m.Command[:], string(0)))
	if cmdstr != "verack" {
		return fmt.Errorf("expected version command, but received: %v", m.Command)
	}
	log.Println(" ✓ received peer verack")

	log.Println(" ✓ handshake completed")

	return nil
}

func (l *Listener) Listen(conn net.Conn) {
	// listen for messages
	for {
		select {
		case <-l.stop:
			return
		default:
			m := &msg.Message{}
			err := m.Deserialize(l.conn)
			if err != nil {
				return
			}

			log.Println("--- received message ---")
			log.Println(m)
			log.Printf("magic = %x\n", m.Magic)
			log.Printf("command = %s\n", m.Command)
			log.Printf("length = %d\n", m.Length)
			log.Printf("checksum = %x\n", m.Checksum)
			log.Printf("payload = %x\n", m.Payload)
		}
	}
}

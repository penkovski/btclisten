package btc

import (
	"bufio"
	"fmt"
	"net"
)

const (
	Version = "31800"

	MagicMainNet  = 0xF9BEB4D9
	MagicTestNet  = 0xFABFB5DA
	MagicTestNet3 = 0x0B110907
	MagicNameCoin = 0xF9BEB4FE
)

type Listener struct {
	conn net.Conn
	done chan bool
}

func NewListener(conn net.Conn, done chan bool) *Listener {
	return &Listener{conn: conn, done: done}
}

func (l *Listener) Start() {
	go l.Listen(l.conn)
	l.Handshake()
}

// Handshake initiates the exchange of version messages between
// the connecting client/node and its peer.
//
// When a node creates an outgoing connection, it will immediately
// advertise its version. The remote node will respond with
// its version. No further communication is possible until both
// peers have exchanged their version.
// https://en.bitcoin.it/wiki/Protocol_documentation
func (l *Listener) Handshake() {

}

func (l *Listener) ping() {

}

func (l *Listener) Listen(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		message := scanner.Text()
		fmt.Println(message)
	}
	l.done <- true
}

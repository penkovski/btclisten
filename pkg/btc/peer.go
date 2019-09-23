package btc

import (
	"net"
	"strconv"
)

type Peer struct {
	IP   [16]byte
	Port uint16
}

func NewPeer(conn net.Conn) (*Peer, error) {
	peer := &Peer{}

	// extract host and port
	ip, port, err := addrIpPort(conn.RemoteAddr())
	if err != nil {
		return nil, err
	}
	copy(peer.IP[:], ip)

	peer.Port = port

	return peer, nil
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

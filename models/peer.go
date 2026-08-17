package models

import (
	"net"
	"sync"
	"time"
)

type Peer struct {
	Username  string
	Hostname  string
	IP        string
	TCPPort   int
	LastSeen  time.Time
	Connected bool
	conn      net.Conn
	mu        sync.Mutex
}

func NewPeer(username, hostname, ip string, tcpPort int) *Peer {
	return &Peer{
		Username:  username,
		Hostname:  hostname,
		IP:        ip,
		TCPPort:   tcpPort,
		LastSeen:  time.Now(),
		Connected: false,
	}
}

func (p *Peer) SetConn(conn net.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.conn = conn
	p.Connected = conn != nil
}

func (p *Peer) Conn() net.Conn {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.conn
}

func (p *Peer) IsExpired(timeout time.Duration) bool {
	return time.Since(p.LastSeen) > timeout
}

func (p *Peer) Touch() {
	p.LastSeen = time.Now()
}

func (p *Peer) Key() string {
	return p.IP + ":" + itoa(p.TCPPort)
}

func (p *Peer) String() string {
	return p.Username + "@" + p.Hostname
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [12]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

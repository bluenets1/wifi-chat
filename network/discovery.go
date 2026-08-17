package network

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/batspeed/wifi-chat/models"
)

const (
	DiscoveryPort    = 9999
	announceInterval = 5 * time.Second
	readTimeout      = 10 * time.Second
)

type AnnouncePacket struct {
	Type     string `json:"type"`
	Username string `json:"username"`
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	TCPPort  int    `json:"tcp_port"`
}

type Discovery struct {
	username    string
	hostname    string
	tcpPort     int
	eventCh     chan models.Event
	peerManager *PeerManager
	udpConn     *net.UDPConn
	stopCh      chan struct{}
	wg          sync.WaitGroup
	running     bool
	mu          sync.Mutex
}

func NewDiscovery(username string, tcpPort int, pm *PeerManager, eventCh chan models.Event) *Discovery {
	hostname, _ := os.Hostname()
	return &Discovery{
		username:    username,
		hostname:    hostname,
		tcpPort:     tcpPort,
		eventCh:     eventCh,
		peerManager: pm,
		stopCh:      make(chan struct{}),
	}
}

func (d *Discovery) Start() error {
	d.mu.Lock()
	if d.running {
		d.mu.Unlock()
		return nil
	}
	d.running = true
	d.mu.Unlock()

	addr := &net.UDPAddr{Port: DiscoveryPort}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		addr = &net.UDPAddr{Port: 0}
		conn, err = net.ListenUDP("udp4", addr)
		if err != nil {
			return fmt.Errorf("discovery listen: %w", err)
		}
	}
	d.udpConn = conn

	d.wg.Add(2)
	go d.listenLoop()
	go d.announceLoop()

	return nil
}

func (d *Discovery) Stop() {
	d.mu.Lock()
	if !d.running {
		d.mu.Unlock()
		return
	}
	d.running = false
	d.mu.Unlock()

	close(d.stopCh)
	if d.udpConn != nil {
		d.udpConn.Close()
	}
	d.wg.Wait()
}

func (d *Discovery) listenLoop() {
	defer d.wg.Done()
	buf := make([]byte, 2048)

	for {
		select {
		case <-d.stopCh:
			return
		default:
			d.udpConn.SetReadDeadline(time.Now().Add(readTimeout))
			n, remoteAddr, err := d.udpConn.ReadFromUDP(buf)
			if err != nil {
				if isTimeout(err) {
					continue
				}
				if isClosed(err) {
					return
				}
				log.Printf("UDP read error: %v", err)
				continue
			}

			data := make([]byte, n)
			copy(data, buf[:n])
			go d.handlePacket(data, remoteAddr)
		}
	}
}

func (d *Discovery) handlePacket(data []byte, remoteAddr *net.UDPAddr) {
	var packet AnnouncePacket
	if err := json.Unmarshal(data, &packet); err != nil {
		return
	}

	if packet.Username == d.username && packet.TCPPort == d.tcpPort {
		return
	}

	if packet.IP == "" {
		packet.IP = remoteAddr.IP.String()
	}

	peer := models.NewPeer(packet.Username, packet.Hostname, packet.IP, packet.TCPPort)
	peer.Touch()

	existing := d.peerManager.GetPeer(peer.Key())
	if existing == nil {
		d.peerManager.AddPeer(peer)
		d.eventCh <- models.NewPeerJoinedEvent(peer)
	} else {
		existing.Touch()
		if !existing.Connected {
			d.peerManager.ConnectToPeer(existing)
		}
	}
}

func (d *Discovery) announceLoop() {
	defer d.wg.Done()

	d.broadcastAnnounce()

	ticker := time.NewTicker(announceInterval)
	defer ticker.Stop()

	for {
		select {
		case <-d.stopCh:
			return
		case <-ticker.C:
			d.broadcastAnnounce()
		}
	}
}

func (d *Discovery) broadcastAnnounce() {
	localIP := getLocalIP()
	if localIP == "" {
		return
	}

	packet := AnnouncePacket{
		Type:     models.MsgTypeAnnounce,
		Username: d.username,
		Hostname: d.hostname,
		IP:       localIP,
		TCPPort:  d.tcpPort,
	}

	data, err := json.Marshal(packet)
	if err != nil {
		return
	}

	if d.udpConn == nil {
		return
	}

	d.udpConn.SetWriteDeadline(time.Now().Add(2 * time.Second))

	targets := getBroadcastTargets()
	for _, addr := range targets {
		d.udpConn.WriteTo(data, addr)
	}
}

func getBroadcastTargets() []*net.UDPAddr {
	var targets []*net.UDPAddr
	targets = append(targets, &net.UDPAddr{IP: net.IPv4bcast, Port: DiscoveryPort})

	addrs, err := getBroadcastAddrs()
	if err == nil {
		for _, ip := range addrs {
			targets = append(targets, &net.UDPAddr{IP: ip, Port: DiscoveryPort})
		}
	}
	return targets
}

func getLocalIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ipv4 := ipnet.IP.To4(); ipv4 != nil {
					return ipv4.String()
				}
			}
		}
	}

	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipv4 := ipnet.IP.To4(); ipv4 != nil {
					return ipv4.String()
				}
			}
		}
	}
	return ""
}

func getBroadcastAddrs() ([]net.IP, error) {
	var addrs []net.IP
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		ifaceAddrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range ifaceAddrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ipv4 := ipnet.IP.To4(); ipv4 != nil {
					broadcast := net.IP(make([]byte, 4))
					for i := 0; i < 4; i++ {
						broadcast[i] = ipnet.IP[i] | ^ipnet.Mask[i]
					}
					addrs = append(addrs, broadcast)
				}
			}
		}
	}
	return addrs, nil
}

func isTimeout(err error) bool {
	if ne, ok := err.(net.Error); ok {
		return ne.Timeout()
	}
	return false
}

func isClosed(err error) bool {
	return err != nil && (fmt.Sprintf("%v", err) == "use of closed network connection" ||
		fmt.Sprintf("%v", err) == "read from closed connection")
}

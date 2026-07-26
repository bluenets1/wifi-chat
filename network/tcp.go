package network

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/batspeed/wifi-chat/models"
	"github.com/google/uuid"
)

type TCPServer struct {
	username    string
	listener    net.Listener
	port        int
	eventCh     chan models.Event
	peerManager *PeerManager
	stopCh      chan struct{}
	wg          sync.WaitGroup
	mu          sync.Mutex
	running     bool
}

func NewTCPServer(username string, pm *PeerManager, eventCh chan models.Event) *TCPServer {
	return &TCPServer{
		username:    username,
		peerManager: pm,
		eventCh:     eventCh,
		stopCh:      make(chan struct{}),
	}
}

func (s *TCPServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return nil
	}

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return fmt.Errorf("tcp listen: %w", err)
	}
	s.listener = listener
	s.port = listener.Addr().(*net.TCPAddr).Port
	s.running = true

	s.wg.Add(1)
	go s.acceptLoop()

	return nil
}

func (s *TCPServer) Port() int {
	return s.port
}

func (s *TCPServer) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stopCh)
	if s.listener != nil {
		s.listener.Close()
	}
	s.wg.Wait()
}

func (s *TCPServer) acceptLoop() {
	defer s.wg.Done()

	for {
		select {
		case <-s.stopCh:
			return
		default:
			s.listener.(*net.TCPListener).SetDeadline(time.Now().Add(time.Second))
			conn, err := s.listener.Accept()
			if err != nil {
				if isTimeout(err) || isClosed(err) {
					continue
				}
				select {
				case <-s.stopCh:
					return
				default:
					log.Printf("TCP accept error: %v", err)
					continue
				}
			}

			s.wg.Add(1)
			go s.handleConnection(conn)
		}
	}
}

func (s *TCPServer) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	remoteUsername, remoteTCPPort, err := s.readIdentify(conn)
	if err != nil {
		return
	}

	remoteIP := conn.RemoteAddr().(*net.TCPAddr).IP.String()
	peerKey := remoteIP + ":" + itoa(remoteTCPPort)

	peer := s.peerManager.GetPeer(peerKey)
	if peer == nil {
		peer = models.NewPeer(remoteUsername, "", remoteIP, remoteTCPPort)
		s.peerManager.AddPeer(peer)
		s.eventCh <- models.NewPeerJoinedEvent(peer)
	}

	if peer.Connected {
		return
	}

	peer.SetConn(conn)

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 65536), 65536)
	for {
		select {
		case <-s.stopCh:
			return
		default:
			if !scanner.Scan() {
				s.peerManager.HandleDisconnect(peer)
				return
			}

			data := scanner.Bytes()
			msg, err := models.MessageFromJSON(data)
			if err != nil {
				continue
			}

			s.handleIncomingMessage(msg, peer)
		}
	}
}

func (s *TCPServer) readIdentify(conn net.Conn) (string, int, error) {
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 4096), 4096)

	if !scanner.Scan() {
		return "", 0, fmt.Errorf("identify timeout")
	}

	var msg models.Message
	if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
		return "", 0, err
	}

	if msg.Type != models.MsgTypeIdentify {
		return "", 0, fmt.Errorf("expected identify message")
	}

	remoteTCPPort, _ := strconv.Atoi(msg.Content)
	return msg.Username, remoteTCPPort, nil
}

func (s *TCPServer) handleIncomingMessage(msg models.Message, peer *models.Peer) {
	peer.Touch()

	switch msg.Type {
	case models.MsgTypeHeartbeat:
		return
	case models.MsgTypeChat, models.MsgTypePrivate, models.MsgTypeBroadcast, models.MsgTypeMe:
		msg.Username = peer.Username
		s.eventCh <- models.NewMessageEvent(&msg)
	default:
		s.eventCh <- models.NewMessageEvent(&msg)
	}
}

func ConnectToPeer(username string, localPort int, peer *models.Peer, eventCh chan models.Event) error {
	addr := fmt.Sprintf("%s:%d", peer.IP, peer.TCPPort)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("connect to %s: %w", addr, err)
	}

	identMsg := models.NewMessage(username, models.MsgTypeIdentify, itoa(localPort))
	identMsg.ID = uuid.New().String()
	data, _ := identMsg.ToJSON()
	data = append(data, '\n')

	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if _, err := conn.Write(data); err != nil {
		conn.Close()
		return err
	}

	peer.SetConn(conn)
	peer.Touch()

	go handlePeerConnection(conn, peer, eventCh)

	return nil
}

func handlePeerConnection(conn net.Conn, peer *models.Peer, eventCh chan models.Event) {
	defer func() {
		peer.SetConn(nil)
		peer.Connected = false
		conn.Close()
	}()

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 65536), 65536)
	for {
		if !scanner.Scan() {
			eventCh <- models.NewPeerLeftEvent(peer)
			return
		}

		data := scanner.Bytes()
		msg, err := models.MessageFromJSON(data)
		if err != nil {
			continue
		}

		peer.Touch()

		switch msg.Type {
		case models.MsgTypeHeartbeat:
			continue
		case models.MsgTypeChat, models.MsgTypePrivate, models.MsgTypeBroadcast, models.MsgTypeMe:
			msg.Username = peer.Username
		}

		eventCh <- models.NewMessageEvent(&msg)
	}
}

func SendMessage(peer *models.Peer, msg models.Message) error {
	conn := peer.Conn()
	if conn == nil {
		return fmt.Errorf("peer %s not connected", peer.Username)
	}

	msg.ID = uuid.New().String()
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	data, err := msg.ToJSON()
	if err != nil {
		return err
	}
	data = append(data, '\n')

	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err = conn.Write(data)
	if err != nil {
		return err
	}

	return nil
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

package network

import (
	"log"
	"sync"
	"time"

	"github.com/batspeed/wifi-chat/models"
	"github.com/google/uuid"
)

const (
	heartbeatInterval = 5 * time.Second
	peerTimeout       = 15 * time.Second
	cleanupInterval   = 10 * time.Second
)

type PeerManager struct {
	username  string
	localPort int
	peers     map[string]*models.Peer
	mu        sync.RWMutex
	eventCh   chan models.Event
	stopCh    chan struct{}
	wg        sync.WaitGroup
}

func NewPeerManager(username string, eventCh chan models.Event) *PeerManager {
	pm := &PeerManager{
		username: username,
		peers:    make(map[string]*models.Peer),
		eventCh:  eventCh,
		stopCh:   make(chan struct{}),
	}
	pm.wg.Add(1)
	go pm.cleanupLoop()
	return pm
}

func (pm *PeerManager) SetLocalPort(port int) {
	pm.localPort = port
}

func (pm *PeerManager) Stop() {
	close(pm.stopCh)
	pm.wg.Wait()
}

func (pm *PeerManager) AddPeer(peer *models.Peer) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	key := peer.Key()
	if existing, ok := pm.peers[key]; ok {
		existing.Touch()
		return
	}
	pm.peers[key] = peer
}

func (pm *PeerManager) RemovePeer(key string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if peer, ok := pm.peers[key]; ok {
		if conn := peer.Conn(); conn != nil {
			conn.Close()
		}
		delete(pm.peers, key)
	}
}

func (pm *PeerManager) GetPeer(key string) *models.Peer {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.peers[key]
}

func (pm *PeerManager) GetPeers() []*models.Peer {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	result := make([]*models.Peer, 0, len(pm.peers))
	for _, p := range pm.peers {
		result = append(result, p)
	}
	return result
}

func (pm *PeerManager) GetPeerByUsername(username string) *models.Peer {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	for _, p := range pm.peers {
		if p.Username == username {
			return p
		}
	}
	return nil
}

func (pm *PeerManager) ConnectToPeer(peer *models.Peer) {
	if peer.Connected {
		return
	}
	if peer.Conn() != nil {
		peer.SetConn(nil)
	}

	go func() {
		err := ConnectToPeer(pm.username, pm.localPort, peer, pm.eventCh)
		if err != nil {
			log.Printf("connect to peer %s: %v", peer.Username, err)
		}
	}()
}

func (pm *PeerManager) ConnectToAll() {
	peers := pm.GetPeers()
	for _, peer := range peers {
		if !peer.Connected {
			pm.ConnectToPeer(peer)
		}
	}
}

func (pm *PeerManager) HandleDisconnect(peer *models.Peer) {
	peer.SetConn(nil)
	peer.Connected = false
	pm.eventCh <- models.NewPeerLeftEvent(peer)
}

func (pm *PeerManager) BroadcastMessage(msg models.Message) {
	msg.ID = uuid.New().String()
	msg.Username = pm.username
	msg.Type = models.MsgTypeChat
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for _, peer := range pm.peers {
		if peer.Connected {
			if err := SendMessage(peer, msg); err != nil {
				log.Printf("send to %s: %v", peer.Username, err)
				pm.HandleDisconnect(peer)
			}
		}
	}
}

func (pm *PeerManager) SendPrivateMessage(targetUsername, content string) error {
	peer := pm.GetPeerByUsername(targetUsername)
	if peer == nil {
		return nil
	}

	msg := models.NewPrivateMessage(pm.username, targetUsername, content)
	msg.ID = uuid.New().String()
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	if !peer.Connected {
		return nil
	}

	return SendMessage(peer, msg)
}

func (pm *PeerManager) SendHeartbeats() {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	msg := models.NewMessage(pm.username, models.MsgTypeHeartbeat, "")
	msg.ID = uuid.New().String()

	for _, peer := range pm.peers {
		if peer.Connected {
			if err := SendMessage(peer, msg); err != nil {
				pm.HandleDisconnect(peer)
			}
		}
	}
}

func (pm *PeerManager) cleanupLoop() {
	defer pm.wg.Done()

	heartbeatTicker := time.NewTicker(heartbeatInterval)
	cleanupTicker := time.NewTicker(cleanupInterval)

	for {
		select {
		case <-pm.stopCh:
			heartbeatTicker.Stop()
			cleanupTicker.Stop()
			return
		case <-heartbeatTicker.C:
			pm.SendHeartbeats()
		case <-cleanupTicker.C:
			pm.checkTimeouts()
		}
	}
}

func (pm *PeerManager) checkTimeouts() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for key, peer := range pm.peers {
		if peer.IsExpired(peerTimeout) {
			if conn := peer.Conn(); conn != nil {
				conn.Close()
			}
			delete(pm.peers, key)
			pm.eventCh <- models.NewPeerLeftEvent(peer)
		}
	}
}

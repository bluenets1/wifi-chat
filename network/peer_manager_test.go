package network

import (
	"testing"
	"time"

	"github.com/batspeed/wifi-chat/models"
)

func TestNewPeerManager(t *testing.T) {
	eventCh := make(chan models.Event, 100)
	pm := NewPeerManager("testuser", eventCh)
	if pm == nil {
		t.Fatal("expected non-nil PeerManager")
	}
	if pm.username != "testuser" {
		t.Errorf("expected testuser, got %s", pm.username)
	}
	pm.Stop()
}

func TestAddAndGetPeer(t *testing.T) {
	eventCh := make(chan models.Event, 100)
	pm := NewPeerManager("testuser", eventCh)
	defer pm.Stop()

	peer := models.NewPeer("alice", "alice-pc", "192.168.1.5", 34567)
	pm.AddPeer(peer)

	got := pm.GetPeer(peer.Key())
	if got == nil {
		t.Fatal("expected to find peer")
	}
	if got.Username != "alice" {
		t.Errorf("expected alice, got %s", got.Username)
	}
}

func TestAddDuplicatePeer(t *testing.T) {
	eventCh := make(chan models.Event, 100)
	pm := NewPeerManager("testuser", eventCh)
	defer pm.Stop()

	peer1 := models.NewPeer("alice", "alice-pc", "192.168.1.5", 34567)
	peer2 := models.NewPeer("alice", "alice-pc", "192.168.1.5", 34567)
	pm.AddPeer(peer1)
	pm.AddPeer(peer2)

	peers := pm.GetPeers()
	if len(peers) != 1 {
		t.Errorf("expected 1 peer, got %d", len(peers))
	}
}

func TestRemovePeer(t *testing.T) {
	eventCh := make(chan models.Event, 100)
	pm := NewPeerManager("testuser", eventCh)
	defer pm.Stop()

	peer := models.NewPeer("alice", "alice-pc", "192.168.1.5", 34567)
	pm.AddPeer(peer)
	pm.RemovePeer(peer.Key())

	got := pm.GetPeer(peer.Key())
	if got != nil {
		t.Error("expected peer to be removed")
	}
}

func TestGetPeers(t *testing.T) {
	eventCh := make(chan models.Event, 100)
	pm := NewPeerManager("testuser", eventCh)
	defer pm.Stop()

	pm.AddPeer(models.NewPeer("alice", "alice-pc", "192.168.1.5", 34567))
	pm.AddPeer(models.NewPeer("bob", "bob-pc", "192.168.1.6", 34568))

	peers := pm.GetPeers()
	if len(peers) != 2 {
		t.Errorf("expected 2 peers, got %d", len(peers))
	}
}

func TestGetPeerByUsername(t *testing.T) {
	eventCh := make(chan models.Event, 100)
	pm := NewPeerManager("testuser", eventCh)
	defer pm.Stop()

	pm.AddPeer(models.NewPeer("alice", "alice-pc", "192.168.1.5", 34567))

	peer := pm.GetPeerByUsername("alice")
	if peer == nil {
		t.Fatal("expected to find alice")
	}
	if peer.Hostname != "alice-pc" {
		t.Errorf("expected alice-pc, got %s", peer.Hostname)
	}

	missing := pm.GetPeerByUsername("nonexistent")
	if missing != nil {
		t.Error("expected nil for nonexistent user")
	}
}

func TestHandleDisconnect(t *testing.T) {
	eventCh := make(chan models.Event, 100)
	pm := NewPeerManager("testuser", eventCh)
	defer pm.Stop()

	peer := models.NewPeer("alice", "alice-pc", "192.168.1.5", 34567)
	peer.Connected = true
	pm.AddPeer(peer)

	pm.HandleDisconnect(peer)

	if peer.Connected {
		t.Error("expected peer to be disconnected")
	}

	select {
	case event := <-eventCh:
		if event.Type != models.EventPeerLeft {
			t.Errorf("expected EventPeerLeft, got %d", event.Type)
		}
	case <-time.After(time.Second):
		t.Error("expected peer left event")
	}
}

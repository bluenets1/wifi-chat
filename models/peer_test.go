package models

import (
	"testing"
	"time"
)

func TestNewPeer(t *testing.T) {
	p := NewPeer("alice", "alice-pc", "192.168.1.5", 34567)
	if p.Username != "alice" {
		t.Errorf("expected alice, got %s", p.Username)
	}
	if p.Hostname != "alice-pc" {
		t.Errorf("expected alice-pc, got %s", p.Hostname)
	}
	if p.IP != "192.168.1.5" {
		t.Errorf("expected 192.168.1.5, got %s", p.IP)
	}
	if p.TCPPort != 34567 {
		t.Errorf("expected 34567, got %d", p.TCPPort)
	}
	if p.LastSeen.IsZero() {
		t.Error("expected LastSeen to be set")
	}
}

func TestPeerKey(t *testing.T) {
	p := NewPeer("alice", "alice-pc", "192.168.1.5", 34567)
	key := p.Key()
	expected := "192.168.1.5:34567"
	if key != expected {
		t.Errorf("expected %s, got %s", expected, key)
	}
}

func TestPeerIsExpired(t *testing.T) {
	p := NewPeer("alice", "alice-pc", "192.168.1.5", 34567)
	p.LastSeen = time.Now().Add(-20 * time.Second)
	if !p.IsExpired(15 * time.Second) {
		t.Error("expected peer to be expired")
	}
	if p.IsExpired(25 * time.Second) {
		t.Error("expected peer not to be expired")
	}
}

func TestPeerTouch(t *testing.T) {
	p := NewPeer("alice", "alice-pc", "192.168.1.5", 34567)
	old := p.LastSeen
	time.Sleep(time.Millisecond)
	p.Touch()
	if !p.LastSeen.After(old) {
		t.Error("expected LastSeen to be updated")
	}
}

func TestPeerString(t *testing.T) {
	p := NewPeer("alice", "alice-pc", "192.168.1.5", 34567)
	s := p.String()
	expected := "alice@alice-pc"
	if s != expected {
		t.Errorf("expected %s, got %s", expected, s)
	}
}

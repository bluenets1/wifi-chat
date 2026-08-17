package models

import (
	"errors"
	"testing"
)

func TestNewPeerJoinedEvent(t *testing.T) {
	peer := NewPeer("alice", "alice-pc", "192.168.1.5", 34567)
	event := NewPeerJoinedEvent(peer)
	if event.Type != EventPeerJoined {
		t.Errorf("expected EventPeerJoined, got %d", event.Type)
	}
	if event.Peer != peer {
		t.Error("expected peer to match")
	}
}

func TestNewPeerLeftEvent(t *testing.T) {
	peer := NewPeer("alice", "alice-pc", "192.168.1.5", 34567)
	event := NewPeerLeftEvent(peer)
	if event.Type != EventPeerLeft {
		t.Errorf("expected EventPeerLeft, got %d", event.Type)
	}
}

func TestNewMessageEvent(t *testing.T) {
	msg := NewMessage("alice", MsgTypeChat, "hello")
	event := NewMessageEvent(&msg)
	if event.Type != EventNewMessage {
		t.Errorf("expected EventNewMessage, got %d", event.Type)
	}
	if event.Message != &msg {
		t.Error("expected message to match")
	}
}

func TestNewErrorEvent(t *testing.T) {
	err := errors.New("test error")
	event := NewErrorEvent(err)
	if event.Type != EventError {
		t.Errorf("expected EventError, got %d", event.Type)
	}
	if event.Error != err {
		t.Error("expected error to match")
	}
}

func TestEventTimestamps(t *testing.T) {
	peer := NewPeer("alice", "alice-pc", "192.168.1.5", 34567)
	event := NewPeerJoinedEvent(peer)
	if event.Timestamp.IsZero() {
		t.Error("expected timestamp to be set")
	}
}

package models

import (
	"testing"
	"time"
)

func TestNewMessage(t *testing.T) {
	msg := NewMessage("alice", MsgTypeChat, "hello")
	if msg.Username != "alice" {
		t.Errorf("expected alice, got %s", msg.Username)
	}
	if msg.Type != MsgTypeChat {
		t.Errorf("expected chat, got %s", msg.Type)
	}
	if msg.Content != "hello" {
		t.Errorf("expected hello, got %s", msg.Content)
	}
}

func TestNewPrivateMessage(t *testing.T) {
	msg := NewPrivateMessage("alice", "bob", "secret")
	if msg.Type != MsgTypePrivate {
		t.Errorf("expected private, got %s", msg.Type)
	}
	if msg.Target != "bob" {
		t.Errorf("expected bob, got %s", msg.Target)
	}
}

func TestNewSystemMessage(t *testing.T) {
	msg := NewSystemMessage("system message")
	if msg.Username != "SYSTEM" {
		t.Errorf("expected SYSTEM, got %s", msg.Username)
	}
	if msg.Type != MsgTypeSystem {
		t.Errorf("expected system, got %s", msg.Type)
	}
}

func TestMessageJSONRoundTrip(t *testing.T) {
	original := Message{
		ID:        "123",
		Username:  "alice",
		Timestamp: time.Now(),
		Type:      MsgTypeChat,
		Content:   "hello world",
		Target:    "",
	}

	data, err := original.ToJSON()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	decoded, err := MessageFromJSON(data)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID: expected %s, got %s", original.ID, decoded.ID)
	}
	if decoded.Username != original.Username {
		t.Errorf("Username: expected %s, got %s", original.Username, decoded.Username)
	}
	if decoded.Type != original.Type {
		t.Errorf("Type: expected %s, got %s", original.Type, decoded.Type)
	}
	if decoded.Content != original.Content {
		t.Errorf("Content: expected %s, got %s", original.Content, decoded.Content)
	}
}

func TestMessageString(t *testing.T) {
	tests := []struct {
		msg  Message
		want string
	}{
		{NewSystemMessage("hello"), "[System] hello"},
		{NewMessage("alice", MsgTypeMe, "waves"), "* alice waves"},
		{NewPrivateMessage("alice", "bob", "secret"), "[Private] alice -> bob: secret"},
		{NewMessage("alice", MsgTypeChat, "hi"), "alice: hi"},
	}

	for _, tt := range tests {
		got := tt.msg.String()
		if got != tt.want {
			t.Errorf("String() = %q, want %q", got, tt.want)
		}
	}
}

func TestMessageFromJSONError(t *testing.T) {
	_, err := MessageFromJSON([]byte(`{invalid json`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

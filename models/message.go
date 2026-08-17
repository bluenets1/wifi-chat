package models

import (
	"encoding/json"
	"time"
)

const (
	MsgTypeChat     = "chat"
	MsgTypePrivate  = "private"
	MsgTypeBroadcast = "broadcast"
	MsgTypeSystem   = "system"
	MsgTypeHeartbeat = "heartbeat"
	MsgTypeAnnounce = "announce"
	MsgTypeIdentify = "identify"
	MsgTypeError    = "error"
	MsgTypeMe       = "me"
)

type Message struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	Target    string    `json:"target,omitempty"`
	Color     string    `json:"-"`
}

func NewMessage(username, msgType, content string) Message {
	return Message{
		ID:        "",
		Username:  username,
		Timestamp: time.Now(),
		Type:      msgType,
		Content:   content,
	}
}

func NewPrivateMessage(username, target, content string) Message {
	return Message{
		ID:        "",
		Username:  username,
		Timestamp: time.Now(),
		Type:      MsgTypePrivate,
		Content:   content,
		Target:    target,
	}
}

func NewSystemMessage(content string) Message {
	return Message{
		ID:        "",
		Username:  "SYSTEM",
		Timestamp: time.Now(),
		Type:      MsgTypeSystem,
		Content:   content,
	}
}

func (m Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

func MessageFromJSON(data []byte) (Message, error) {
	var m Message
	err := json.Unmarshal(data, &m)
	return m, err
}

func (m Message) String() string {
	if m.Type == MsgTypeSystem {
		return "[System] " + m.Content
	}
	if m.Type == MsgTypeMe {
		return "* " + m.Username + " " + m.Content
	}
	if m.Type == MsgTypePrivate {
		return "[Private] " + m.Username + " -> " + m.Target + ": " + m.Content
	}
	return m.Username + ": " + m.Content
}

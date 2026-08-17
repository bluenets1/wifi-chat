package models

import "time"

type EventType int

const (
	EventPeerJoined EventType = iota
	EventPeerLeft
	EventNewMessage
	EventPeerUpdated
	EventError
)

type Event struct {
	Type      EventType
	Peer      *Peer
	Message   *Message
	Timestamp time.Time
	Error     error
}

func NewPeerJoinedEvent(peer *Peer) Event {
	return Event{
		Type:      EventPeerJoined,
		Peer:      peer,
		Timestamp: time.Now(),
	}
}

func NewPeerLeftEvent(peer *Peer) Event {
	return Event{
		Type:      EventPeerLeft,
		Peer:      peer,
		Timestamp: time.Now(),
	}
}

func NewMessageEvent(msg *Message) Event {
	return Event{
		Type:      EventNewMessage,
		Message:   msg,
		Timestamp: time.Now(),
	}
}

func NewErrorEvent(err error) Event {
	return Event{
		Type:      EventError,
		Error:     err,
		Timestamp: time.Now(),
	}
}

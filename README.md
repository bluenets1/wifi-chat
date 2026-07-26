# WiFi Chat

A terminal-based LAN chat application for users on the same local Wi-Fi/LAN network. No internet connection, no central server, and no manual IP configuration required.

## Features

- **Automatic Peer Discovery** — UDP broadcast-based discovery, no configuration needed
- **Real-Time Messaging** — TCP-based chat with public, private, and broadcast messages
- **Terminal UI** — Bubble Tea TUI with split-panel layout
- **Commands** — `/help`, `/users`, `/msg`, `/broadcast`, `/clear`, `/exit`, `/ping`, `/me`, `/nick`
- **Colorized Usernames** — Each user gets a unique color
- **Heartbeat & Timeout** — Automatic peer health monitoring
- **Cross-Platform** — Linux, macOS, Windows

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                     main.go                          │
├──────────┬──────────┬──────────┬──────────┬─────────┤
│ network/ │   ui/    │ models/  │commands/ │storage/ │
├──────────┤          │          │          │         │
│discovery │terminal  │  peer    │ parser   │ history │
│ tcp      │ colors   │ message  │          │         │
│ peer_mgr │ tui      │ events   │          │         │
└──────────┴──────────┴──────────┴──────────┴─────────┘
```

### Network Flow

```
┌─────────┐   UDP Broadcast (Port 9999)   ┌─────────┐
│ Client  │ ────────────────────────────→ │  Peer   │
│    A    │ ←──────────────────────────── │    B    │
│         │   TCP (random port)           │         │
│         │ ←═══════════════════════════→ │         │
└─────────┘     Messages over TCP         └─────────┘
```

- **UDP**: Peer discovery and heartbeat announcements
- **TCP**: Message delivery (each client is both server and client)

## Quick Start

### Prerequisites

- Go 1.21+

### Install

```bash
git clone <repo-url> wifi-chat
cd wifi-chat
go build -o wifi-chat .
```

### Run

```bash
./wifi-chat
```

Run on multiple machines on the same LAN. Peers are discovered automatically.

## Commands

```
/help                    Show help
/users                   List online users
/msg <user> <message>    Send private message
/broadcast <message>     Broadcast to all
/clear                   Clear chat
/exit                    Exit
/ping                    Check connectivity
/me <message>            Action message
/nick <newname>          Change username
```

## Tests

```bash
go test ./...
```

## Extending

The modular architecture supports adding:

- File sharing
- Voice messages  ·  Image transfer
- End-to-end encryption
- Mesh networking
- Chat rooms
- Persistent chat history
- Plugins

## Keyboard

| Key | Action |
|-----|--------|
| Type | Enter message |
| Enter | Send |
| Ctrl+C | Exit |
| Ctrl+L | Clear screen |
| PgUp/PgDn | Scroll chat |
| ↑/↓ | Scroll line |

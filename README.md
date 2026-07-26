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

### Run on Termux (Android)

1. Install Termux from F-Droid (recommended) or Google Play Store
2. Open Termux and install Go:
   ```bash
   pkg update && pkg upgrade
   pkg install golang git
   ```
3. Clone and build:
   ```bash
   git clone <repo-url> wifi-chat
   cd wifi-chat
   go build -o wifi-chat .
   ```
4. Run:
   ```bash
   ./wifi-chat
   ```
5. Make sure your phone is connected to the **same Wi-Fi network** as your desktop
6. For the best experience, use a **landscape orientation** or a terminal emulator that supports large sizes (like Termux's own fullscreen mode)

> **Note:** On some Android devices, UDP broadcast to `255.255.255.255` may not work due to WiFi firmware restrictions. The app also sends discovery packets to subnet broadcast addresses (e.g., `192.168.1.255`) and standard broadcast — one of these methods will work on most devices.

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

## Termux Tips

- **Keyboard shortcuts**: Termux's extra keys row (Volume+Q) includes Tab, Ctrl, Alt keys useful for shortcuts like Ctrl+C (exit) and Ctrl+L (clear)
- **Fullscreen**: Use Termux's fullscreen mode (swipe down from top) for more chat space
- **Landscape mode**: Turn your phone sideways for a wider layout
- **Background**: To keep the chat running when switching apps, use `tmux` or `screen`:
  ```bash
  pkg install tmux
  tmux
  ./wifi-chat
  ```
- **Exit cleanly**: Type `/exit` or press Ctrl+C to disconnect gracefully

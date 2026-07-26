package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/batspeed/wifi-chat/models"
	"github.com/batspeed/wifi-chat/network"
	"github.com/batspeed/wifi-chat/ui"
)

func main() {
	username := getDefaultUsername()
	fmt.Printf("LAN Chat starting...\n")
	fmt.Printf("Username: %s\n", username)

	eventCh := make(chan models.Event, 256)

	pm := network.NewPeerManager(username, eventCh)

	tcpServer := network.NewTCPServer(username, pm, eventCh)
	if err := tcpServer.Start(); err != nil {
		log.Fatalf("Failed to start TCP server: %v", err)
	}

	disc := network.NewDiscovery(username, tcpServer.Port(), pm, eventCh)
	if err := disc.Start(); err != nil {
		log.Fatalf("Failed to start discovery: %v", err)
	}

	model := ui.NewModel(username, pm, eventCh)
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	disc.Stop()
	tcpServer.Stop()
	pm.Stop()
}

func getDefaultUsername() string {
	envUser := os.Getenv("USER")
	if envUser != "" {
		return envUser
	}
	u, err := user.Current()
	if err == nil && u.Username != "" {
		name := strings.Split(u.Username, " ")[0]
		if name != "" {
			return name
		}
	}
	hostname, err := os.Hostname()
	if err == nil {
		return hostname
	}
	return "anonymous"
}

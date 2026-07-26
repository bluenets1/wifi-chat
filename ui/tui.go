package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"

	"github.com/batspeed/wifi-chat/commands"
	"github.com/batspeed/wifi-chat/models"
	"github.com/batspeed/wifi-chat/network"
)

type eventMsg struct {
	event models.Event
}

type tickMsg struct{}

type Model struct {
	username      string
	messages      []models.Message
	peers         map[string]*models.Peer
	peerList      []string
	chatViewport  viewport.Model
	input         string
	ready         bool
	width         int
	height        int
	err           error
	eventCh       chan models.Event
	peerManager   *network.PeerManager
	showHelp      bool
}

func NewModel(username string, pm *network.PeerManager, eventCh chan models.Event) Model {
	return Model{
		username:    username,
		messages:    make([]models.Message, 0, 1000),
		peers:       make(map[string]*models.Peer),
		peerList:    make([]string, 0),
		eventCh:     eventCh,
		peerManager: pm,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		waitForEvent(m.eventCh),
		tickCmd(),
	)
}

func waitForEvent(eventCh chan models.Event) tea.Cmd {
	return func() tea.Msg {
		event := <-eventCh
		return eventMsg{event: event}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		viewportWidth := msg.Width - 30 - 4
		viewportHeight := msg.Height - 6
		if viewportWidth < 10 {
			viewportWidth = 10
		}
		if viewportHeight < 3 {
			viewportHeight = 3
		}
		if !m.ready {
			m.chatViewport = viewport.New(viewportWidth, viewportHeight)
			m.chatViewport.YPosition = 2
			m.chatViewport.HighPerformanceRendering = false
			m.ready = true
		} else {
			m.chatViewport.Width = viewportWidth
			m.chatViewport.Height = viewportHeight
		}

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEnter:
			if m.input == "" {
				break
			}
			m.handleInput()

		case tea.KeyBackspace:
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}

		case tea.KeySpace:
			m.input += " "

		case tea.KeyTab:
			m.input += "\t"

		case tea.KeyRunes:
			m.input += string(msg.Runes)

		case tea.KeyPgUp:
			m.chatViewport.HalfViewUp()

		case tea.KeyPgDown:
			m.chatViewport.HalfViewDown()

		case tea.KeyUp:
			m.chatViewport.LineUp(1)

		case tea.KeyDown:
			m.chatViewport.LineDown(1)

		case tea.KeyCtrlL:
			m.messages = m.messages[:0]
		}

	case eventMsg:
		m.handleEvent(msg.event)

	case tickMsg:
		cmds = append(cmds, tickCmd())
	}

	if m.ready {
		if m.chatViewport.AtBottom() {
			m.chatViewport.SetContent(m.renderChatContent())
			m.chatViewport.GotoBottom()
		} else {
			m.chatViewport.SetContent(m.renderChatContent())
		}
	}

	cmds = append(cmds, waitForEvent(m.eventCh))
	return m, tea.Batch(cmds...)
}

func (m *Model) handleInput() {
	input := strings.TrimSpace(m.input)
	m.input = ""

	if input == "" {
		return
	}

	cmd := commands.Parse(input)

	switch cmd.Type {
	case commands.CmdExit:
		return

	case commands.CmdClear:
		m.messages = m.messages[:0]
		return

	case commands.CmdUsers:
		users := make([]string, 0, len(m.peerList))
		for _, name := range m.peerList {
			users = append(users, name)
		}
		msg := models.NewSystemMessage(commands.ListUsers(users))
		m.messages = append(m.messages, msg)
		return

	case commands.CmdHelp:
		msg := models.NewSystemMessage(commands.HelpText())
		m.messages = append(m.messages, msg)
		return

	case commands.CmdNick:
		if len(cmd.Args) > 0 {
			m.username = cmd.Args[0]
			msg := models.NewSystemMessage(fmt.Sprintf("You changed your name to %s", m.username))
			m.messages = append(m.messages, msg)
		}
		return

	case commands.CmdPing:
		msg := models.NewSystemMessage(fmt.Sprintf("pong (%d peers online)", len(m.peerList)))
		m.messages = append(m.messages, msg)
		return

	case commands.CmdMsg:
		if len(cmd.Args) < 2 {
			msg := models.NewSystemMessage("usage: /msg <username> <message>")
			m.messages = append(m.messages, msg)
			return
		}
		target := cmd.Args[0]
		content := strings.Join(cmd.Args[1:], " ")
		chatMsg := models.NewPrivateMessage(m.username, target, content)
		chatMsg.ID = uuid.New().String()
		m.messages = append(m.messages, chatMsg)
		go m.peerManager.SendPrivateMessage(target, content)
		return

	case commands.CmdBroadcast:
		content := strings.Join(cmd.Args, " ")
		msg := models.NewMessage(m.username, models.MsgTypeChat, content)
		msg.ID = uuid.New().String()
		m.messages = append(m.messages, msg)
		go func() {
			broadcastMsg := models.NewMessage(m.username, models.MsgTypeChat, content)
			broadcastMsg.ID = uuid.New().String()
			m.peerManager.BroadcastMessage(broadcastMsg)
		}()
		return

	case commands.CmdMe:
		content := strings.Join(cmd.Args, " ")
		msg := models.NewMessage(m.username, models.MsgTypeMe, content)
		msg.ID = uuid.New().String()
		m.messages = append(m.messages, msg)
		go func() {
			broadcastMsg := models.NewMessage(m.username, models.MsgTypeMe, content)
			broadcastMsg.ID = uuid.New().String()
			m.peerManager.BroadcastMessage(broadcastMsg)
		}()
		return

	case commands.CmdUnknown:
		content := input
		msg := models.NewMessage(m.username, models.MsgTypeChat, content)
		msg.ID = uuid.New().String()
		m.messages = append(m.messages, msg)
		go func() {
			broadcastMsg := models.NewMessage(m.username, models.MsgTypeChat, content)
			broadcastMsg.ID = uuid.New().String()
			m.peerManager.BroadcastMessage(broadcastMsg)
		}()
		return
	}
}

func (m *Model) handleEvent(event models.Event) {
	switch event.Type {
	case models.EventPeerJoined:
		if event.Peer != nil {
			m.peers[event.Peer.Key()] = event.Peer
			m.refreshPeerList()
			msg := models.NewSystemMessage(fmt.Sprintf("%s has joined the chat", event.Peer.Username))
			m.messages = append(m.messages, msg)
		}

	case models.EventPeerLeft:
		if event.Peer != nil {
			delete(m.peers, event.Peer.Key())
			m.refreshPeerList()
			msg := models.NewSystemMessage(fmt.Sprintf("%s has left the chat", event.Peer.Username))
			m.messages = append(m.messages, msg)
		}

	case models.EventNewMessage:
		if event.Message != nil {
			if event.Message.Username == m.username {
				return
			}
			m.messages = append(m.messages, *event.Message)
		}

	case models.EventError:
		if event.Error != nil {
			msg := models.NewSystemMessage(fmt.Sprintf("error: %v", event.Error))
			m.messages = append(m.messages, msg)
		}
	}
}

func (m *Model) refreshPeerList() {
	names := make([]string, 0, len(m.peers))
	for _, peer := range m.peers {
		names = append(names, peer.Username)
	}
	m.peerList = names
}

func (m Model) View() string {
	if !m.ready {
		return "\n  Loading..."
	}

	header := HeaderStyle.
		Width(m.width - 4).
		Render(fmt.Sprintf(" LAN CHAT - %s ", m.username))

	usersPanel := m.renderUsersPanel()
	chatPanel := ChatPanelStyle.
		Width(m.width-30-4).
		Height(m.height-4-2).
		Render(m.chatViewport.View())

	body := lipgloss.JoinHorizontal(
		lipgloss.Top,
		usersPanel,
		chatPanel,
	)

	inputBar := InputStyle.
		Width(m.width - 4).
		Render(fmt.Sprintf(" > %s", m.input))

	statusText := fmt.Sprintf(" %d users online | %d msgs | /help", len(m.peerList), len(m.messages))
	statusBar := StatusBarStyle.
		Width(m.width - 4).
		Render(statusText)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		body,
		inputBar,
		statusBar,
	)

	return AppStyle.Render(content)
}

func (m Model) renderUsersPanel() string {
	panelWidth := 28
	var b strings.Builder
	b.WriteString(" Online Users\n")
	b.WriteString(" ─────────────\n")
	if len(m.peerList) == 0 {
		b.WriteString("  No other users\n")
	} else {
		for _, name := range m.peerList {
			b.WriteString(fmt.Sprintf("  ● %s\n", UserOnlineStyle.Render(name)))
		}
	}

	return UsersPanelStyle.
		Width(panelWidth).
		Height(m.height - 4 - 2).
		Render(b.String())
}

func (m Model) renderChatContent() string {
	if len(m.messages) == 0 {
		return "  Welcome to LAN Chat!\n  Type /help for commands."
	}

	var b strings.Builder
	for _, msg := range m.messages {
		b.WriteString(m.formatMessage(msg))
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func (m Model) formatMessage(msg models.Message) string {
	timestamp := TimestampStyle.Render(msg.Timestamp.Format("15:04:05"))

	switch msg.Type {
	case models.MsgTypeSystem:
		return fmt.Sprintf("%s %s", timestamp, SystemMsgStyle.Render(msg.Content))

	case models.MsgTypePrivate:
		if msg.Username == m.username {
			return fmt.Sprintf("%s %s %s",
				timestamp,
				PrivateMsgStyle.Render(fmt.Sprintf("[to %s]", msg.Target)),
				msg.Content)
		}
		return fmt.Sprintf("%s %s %s",
			timestamp,
			PrivateMsgStyle.Render(fmt.Sprintf("[from %s]", msg.Username)),
			msg.Content)

	case models.MsgTypeMe:
		return fmt.Sprintf("%s * %s %s",
			timestamp,
			ColorForUsername(msg.Username).Render(msg.Username),
			MeMsgStyle.Render(msg.Content))

	case models.MsgTypeBroadcast:
		return fmt.Sprintf("%s %s %s",
			timestamp,
			ColorForUsername(msg.Username).Render(msg.Username+": "),
			msg.Content)

	default:
		if msg.Username == m.username {
			return fmt.Sprintf("%s %s %s",
				timestamp,
				YourMsgStyle.Render("You: "),
				msg.Content)
		}
		return fmt.Sprintf("%s %s %s",
			timestamp,
			ColorForUsername(msg.Username).Render(msg.Username+": "),
			msg.Content)
	}
}

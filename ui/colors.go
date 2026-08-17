package ui

import (
	"hash/fnv"
	"sync"

	"github.com/charmbracelet/lipgloss"
)

var (
	AppStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63"))

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("63")).
			Padding(0, 2)

	UsersPanelStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)

	ChatPanelStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)

	InputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")).
			Background(lipgloss.Color("236")).
			Padding(0, 1).Width(80)

	SystemMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")).
			Italic(true)

	PrivateMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))

	MeMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	YourMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("84"))

	UserOnlineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("84"))

	UserOfflineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	TimestampStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))
)

var userColors = []string{
	"33",  // blue
	"45",  // cyan
	"46",  // spring green
	"214", // orange
	"205", // pink
	"99",  // purple
	"82",  // green
	"208", // dark orange
	"39",  // deep sky blue
	"191", // yellow-green
	"129", // magenta
	"87",  // aquamarine
}

var colorCache = make(map[string]lipgloss.Style)
var colorMu sync.Mutex

func ColorForUsername(username string) lipgloss.Style {
	colorMu.Lock()
	defer colorMu.Unlock()

	if style, ok := colorCache[username]; ok {
		return style
	}

	h := fnv.New32a()
	h.Write([]byte(username))
	idx := int(h.Sum32()) % len(userColors)

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(userColors[idx])).
		Bold(true)

	colorCache[username] = style
	return style
}

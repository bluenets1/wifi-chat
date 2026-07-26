package commands

import (
	"fmt"
	"strings"
)

type CommandType int

const (
	CmdHelp     CommandType = iota
	CmdUsers    CommandType = iota
	CmdMsg      CommandType = iota
	CmdBroadcast CommandType = iota
	CmdClear    CommandType = iota
	CmdExit     CommandType = iota
	CmdPing     CommandType = iota
	CmdMe       CommandType = iota
	CmdNick     CommandType = iota
	CmdUnknown  CommandType = iota
)

var commandNames = map[CommandType]string{
	CmdHelp:      "help",
	CmdUsers:     "users",
	CmdMsg:       "msg",
	CmdBroadcast: "broadcast",
	CmdClear:     "clear",
	CmdExit:      "exit",
	CmdPing:      "ping",
	CmdMe:        "me",
	CmdNick:      "nick",
}

type Command struct {
	Type    CommandType
	Raw     string
	Args    []string
	Message string
}

func Parse(input string) Command {
	input = strings.TrimSpace(input)
	if input == "" {
		return Command{Type: CmdUnknown, Raw: input}
	}

	if !strings.HasPrefix(input, "/") {
		return Command{
			Type:    CmdMsg,
			Raw:     input,
			Message: input,
		}
	}

	parts := strings.Fields(input)
	if len(parts) == 0 {
		return Command{Type: CmdUnknown, Raw: input}
	}

	cmdName := strings.TrimPrefix(parts[0], "/")
	args := parts[1:]

	var cmdType CommandType
	found := false
	for t, name := range commandNames {
		if name == cmdName {
			cmdType = t
			found = true
			break
		}
	}
	if !found {
		return Command{Type: CmdUnknown, Raw: input, Args: args}
	}

	return Command{
		Type: cmdType,
		Raw:  input,
		Args: args,
	}
}

func (c Command) Execute(username string) (string, bool) {
	switch c.Type {
	case CmdHelp:
		return HelpText(), false
	case CmdUsers:
		return "", false
	case CmdMsg:
		return c.Message, false
	case CmdBroadcast:
		return strings.Join(c.Args, " "), false
	case CmdClear:
		return "", true
	case CmdExit:
		return "", true
	case CmdPing:
		return "pong", false
	case CmdMe:
		return strings.Join(c.Args, " "), false
	case CmdNick:
		return strings.Join(c.Args, " "), false
	default:
		return fmt.Sprintf("unknown command: %s. Type /help for available commands.", c.Raw), false
	}
}

func HelpText() string {
	return `Available Commands:
  /help                    Show this help message
  /users                   List online users
  /msg <user> <message>    Send a private message
  /broadcast <message>     Send a broadcast message to all users
  /clear                   Clear chat window
  /exit                    Exit the application
  /ping                    Check connection status
  /me <message>            Send an action message
  /nick <newname>          Change your username

Shortcuts:
  Ctrl+C    Exit
  Ctrl+L    Clear screen
  PageUp/Dn Scroll chat`
}

func ListUsers(users []string) string {
	if len(users) == 0 {
		return "No other users online."
	}
	var b strings.Builder
	b.WriteString("Online Users:\n")
	for _, u := range users {
		b.WriteString(fmt.Sprintf("  ● %s\n", u))
	}
	return strings.TrimRight(b.String(), "\n")
}

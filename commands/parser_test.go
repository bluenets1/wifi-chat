package commands

import (
	"strings"
	"testing"
)

func TestParseChatMessage(t *testing.T) {
	cmd := Parse("hello everyone")
	if cmd.Type != CmdMsg {
		t.Errorf("expected CmdMsg, got %v", cmd.Type)
	}
	if cmd.Message != "hello everyone" {
		t.Errorf("expected 'hello everyone', got '%s'", cmd.Message)
	}
}

func TestParseEmpty(t *testing.T) {
	cmd := Parse("")
	if cmd.Type != CmdUnknown {
		t.Errorf("expected CmdUnknown, got %v", cmd.Type)
	}
}

func TestParseSlashCommand(t *testing.T) {
	tests := []struct {
		input    string
		cmdType  CommandType
		argsLen  int
	}{
		{"/help", CmdHelp, 0},
		{"/users", CmdUsers, 0},
		{"/clear", CmdClear, 0},
		{"/exit", CmdExit, 0},
		{"/ping", CmdPing, 0},
		{"/msg bob hello", CmdMsg, 2},
		{"/me waves hello", CmdMe, 2},
		{"/nick newname", CmdNick, 1},
		{"/broadcast hello all", CmdBroadcast, 2},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			cmd := Parse(tt.input)
			if cmd.Type != tt.cmdType {
				t.Errorf("expected type %v, got %v", tt.cmdType, cmd.Type)
			}
			if len(cmd.Args) != tt.argsLen {
				t.Errorf("expected %d args, got %d: %v", tt.argsLen, len(cmd.Args), cmd.Args)
			}
		})
	}
}

func TestParseUnknownCommand(t *testing.T) {
	cmd := Parse("/unknown")
	if cmd.Type != CmdUnknown {
		t.Errorf("expected CmdUnknown, got %v", cmd.Type)
	}
}

func TestHelpText(t *testing.T) {
	text := HelpText()
	if !strings.Contains(text, "/help") {
		t.Error("help should contain /help")
	}
	if !strings.Contains(text, "/exit") {
		t.Error("help should contain /exit")
	}
}

func TestListUsers(t *testing.T) {
	result := ListUsers([]string{"alice", "bob"})
	if !strings.Contains(result, "alice") {
		t.Error("result should contain alice")
	}
	if !strings.Contains(result, "bob") {
		t.Error("result should contain bob")
	}
}

func TestListUsersEmpty(t *testing.T) {
	result := ListUsers(nil)
	if !strings.Contains(result, "No other users") {
		t.Error("empty users should show appropriate message")
	}
}

func TestCommandExecute(t *testing.T) {
	cmd := Parse("/help")
	msg, clear := cmd.Execute("test")
	if clear {
		t.Error("help should not clear")
	}
	if msg == "" {
		t.Error("help should return text")
	}
}

func TestCommandExecuteClear(t *testing.T) {
	cmd := Parse("/clear")
	_, clear := cmd.Execute("test")
	if !clear {
		t.Error("clear should return clear=true")
	}
}

func TestCommandExecuteExit(t *testing.T) {
	cmd := Parse("/exit")
	_, clear := cmd.Execute("test")
	if !clear {
		t.Error("exit should return clear=true")
	}
}

func TestCommandExecuteUnknown(t *testing.T) {
	cmd := Parse("/unknown")
	msg, _ := cmd.Execute("test")
	if msg == "" {
		t.Error("unknown command should return error message")
	}
}

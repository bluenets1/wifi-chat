package storage

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/batspeed/wifi-chat/models"
)

const historyFileName = "wifi-chat-history.jsonl"

type History struct {
	filePath string
	file     *os.File
	mu       sync.Mutex
	enabled  bool
}

func NewHistory() *History {
	h := &History{enabled: false}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return h
	}
	dir := filepath.Join(homeDir, ".wifi-chat")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return h
	}
	path := filepath.Join(dir, historyFileName)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return h
	}
	h.filePath = path
	h.file = f
	h.enabled = true
	return h
}

func (h *History) Save(msg models.Message) error {
	if !h.enabled {
		return nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = h.file.Write(data)
	return err
}

func (h *History) Load() ([]models.Message, error) {
	if !h.enabled {
		return nil, nil
	}

	f, err := os.Open(h.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var messages []models.Message
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var msg models.Message
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			continue
		}
		messages = append(messages, msg)
	}
	return messages, scanner.Err()
}

func (h *History) Close() {
	if h.file != nil {
		h.file.Close()
	}
}

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Session struct {
	AccessToken string    `json:"access_token"`
	UserID      string    `json:"user_id"`
	ServerID    string    `json:"server_id"`
	DeviceID    string    `json:"device_id"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func LoadSession(path string) (*Session, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("session: parse: %w", err)
	}
	return &s, nil
}

func SaveSession(path string, s *Session) error {
	if err := EnsureDir(dirname(path)); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

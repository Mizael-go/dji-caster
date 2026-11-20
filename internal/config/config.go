package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Settings is persisted as JSON in the user config dir.
type Settings struct {
	RtmpPort int    `json:"rtmpPort"`
	ApiPort  int    `json:"apiPort"`
	MediaMtx string `json:"mediaMtxPath"`
	FFmpeg   string `json:"ffmpegPath"`
}

func Defaults() Settings {
	return Settings{
		RtmpPort: 1935,
		ApiPort:  9997,
	}
}

// Dir is the per-user data directory of the app ("~/.config/dji-caster" on
// Linux). Binaries downloaded by the app live in Dir()/bin so a machine never
// depends on a system-wide MediaMTX install.
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		home, herr := os.UserHomeDir()
		if herr != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	dir := filepath.Join(base, "dji-caster")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func BinDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	bin := filepath.Join(dir, "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		return "", err
	}
	return bin, nil
}

func path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "settings.json"), nil
}

func Load() (Settings, error) {
	s := Defaults()
	p, err := path()
	if err != nil {
		return s, err
	}
	raw, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return s, err
	}
	if err := json.Unmarshal(raw, &s); err != nil {
		return Defaults(), err
	}
	if s.RtmpPort <= 0 {
		s.RtmpPort = 1935
	}
	if s.ApiPort <= 0 {
		s.ApiPort = 9997
	}
	return s, nil
}

func Save(s Settings) error {
	p, err := path()
	if err != nil {
		return err
	}
	raw, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, raw, 0o600)
}

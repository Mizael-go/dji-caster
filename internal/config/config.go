package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Network is a Wi-Fi profile the camera may join. It is remembered so the
// same venue does not need the password typed again.
type Network struct {
	SSID     string `json:"ssid"`
	PSK      string `json:"psk"`
	LastUsed int64  `json:"lastUsed"`
}

// Settings is persisted as JSON in the user config dir. It contains Wi-Fi
// passwords, so the file is written with 0600 permissions.
type Settings struct {
	Networks   []Network `json:"networks"`
	RtmpPort   int       `json:"rtmpPort"`
	ApiPort    int       `json:"apiPort"`
	Transcode  bool      `json:"transcode"`
	LastDevice string    `json:"lastDevice"`
	MediaMtx   string    `json:"mediaMtxPath"`
	DjiCtl     string    `json:"djiCtlPath"`
	FFmpeg     string    `json:"ffmpegPath"`
	// Réglages appliqués à distance à l'encodeur de la caméra DJI lors du
	// forçage BLE (options --resolution/--bitrate-kbps/--fps de djictl).
	StreamResolution string `json:"streamResolution"`
	StreamBitrate    int    `json:"streamBitrate"`
	StreamFPS        int    `json:"streamFps"`
}

func Defaults() Settings {
	return Settings{
		Networks:         []Network{},
		RtmpPort:         1935,
		ApiPort:          9997,
		StreamResolution: "1080p",
		StreamBitrate:    6000,
		StreamFPS:        30,
	}
}

// NormalizeStream normalise les réglages d'encodage de la caméra aux valeurs
// acceptées par djictl : 480p/720p/1080p, 25 ou 30 fps, débit 1000–20000 kbps.
func (s *Settings) NormalizeStream() {
	switch s.StreamResolution {
	case "480p", "720p", "1080p":
	default:
		s.StreamResolution = "1080p"
	}
	if s.StreamFPS != 25 && s.StreamFPS != 30 {
		s.StreamFPS = 30
	}
	if s.StreamBitrate < 1000 || s.StreamBitrate > 20000 {
		s.StreamBitrate = 6000
	}
}

// UpsertNetwork stores (or refreshes) a profile and marks it as the most
// recently used one.
func (s *Settings) UpsertNetwork(ssid, psk string) {
	ssid = strings.TrimSpace(ssid)
	if ssid == "" {
		return
	}
	now := time.Now().UnixMilli()
	for i := range s.Networks {
		if s.Networks[i].SSID == ssid {
			s.Networks[i].PSK = psk
			s.Networks[i].LastUsed = now
			s.sortNetworks()
			return
		}
	}
	s.Networks = append(s.Networks, Network{SSID: ssid, PSK: psk, LastUsed: now})
	s.sortNetworks()
}

func (s *Settings) RemoveNetwork(ssid string) bool {
	for i := range s.Networks {
		if s.Networks[i].SSID == ssid {
			s.Networks = append(s.Networks[:i], s.Networks[i+1:]...)
			return true
		}
	}
	return false
}

// PrimaryNetwork is the profile used by default in the connect form.
func (s *Settings) PrimaryNetwork() (Network, bool) {
	if len(s.Networks) == 0 {
		return Network{}, false
	}
	return s.Networks[0], true
}

func (s *Settings) sortNetworks() {
	sort.SliceStable(s.Networks, func(i, j int) bool {
		return s.Networks[i].LastUsed > s.Networks[j].LastUsed
	})
}

// Dir is the per-user data directory of the app ("~/.config/dji-caster" on
// Linux). Binaries downloaded or built by the app live in Dir()/bin so a
// machine never depends on a system-wide MediaMTX install.
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
	if s.Networks == nil {
		s.Networks = []Network{}
	}
	s.NormalizeStream()
	s.sortNetworks()
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

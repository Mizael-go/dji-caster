package media

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Paths published on the local MediaMTX. The DJI always pushes to PathIn; when
// transcoding is enabled ffmpeg republishes an H.264 copy on PathOut so OBS
// never has to decode the camera's HEVC stream.
const (
	PathIn  = "live/dji"
	PathOut = "live/dji-h264"

	// RtspPort is the loopback-only RTSP listener. MediaMTX does not forward
	// H.265 to plain RTMP readers, so ffmpeg reads the camera stream over RTSP
	// (H.265 supported natively) and republishes H.264 over RTMP.
	RtspPort = 8554
)

type ServerConfig struct {
	Binary    string
	Dir       string
	RtmpPort  int
	ApiPort   int
	Logf      func(string)
}

type PathStat struct {
	Name          string   `json:"name"`
	Ready         bool     `json:"ready"`
	Tracks        []string `json:"tracks"`
	BytesReceived int64    `json:"bytesReceived"`
	Readers       int      `json:"readers"`
}

type Server struct {
	mu     sync.Mutex
	cfg    ServerConfig
	cmd    *exec.Cmd
	err    string
	done   chan struct{}
	onExit func(string)
}

func NewServer(cfg ServerConfig) *Server {
	return &Server{cfg: cfg}
}

func (s *Server) Running() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cmd != nil
}

func (s *Server) Err() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.err
}

func (s *Server) SetOnExit(fn func(string)) {
	s.mu.Lock()
	s.onExit = fn
	s.mu.Unlock()
}

// ConfigPath is where the generated MediaMTX config lives.
func ConfigPath(dir string) string { return filepath.Join(dir, "mediamtx.yml") }

func (s *Server) writeConfig() (string, error) {
	cfg := fmt.Sprintf(`# Généré par DJI Caster — ne pas éditer à la main.
logLevel: info
logDestinations: [stdout]
readTimeout: 15s
writeTimeout: 15s

# RTMP est le protocole exposé (le DJI pousse, OBS lit).
# RTSP n'écoute qu'en loopback : c'est par là que ffmpeg lit le H.265.
rtsp: yes
rtspAddress: 127.0.0.1:%d
hls: no
webrtc: no
srt: no
rtmp: yes
rtmpAddress: :%d
rtmpEncryption: "no"

# API locale utilisée par l'app pour savoir si un flux est réellement reçu.
api: yes
apiAddress: 127.0.0.1:%d

paths:
  all_others:
`, RtspPort, s.cfg.RtmpPort, s.cfg.ApiPort)

	path := ConfigPath(s.cfg.Dir)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(cfg), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cmd != nil {
		return nil
	}
	if s.cfg.Binary == "" {
		return fmt.Errorf("MediaMTX n'est pas installé")
	}
	if portFree(s.cfg.RtmpPort) == false {
		return fmt.Errorf("le port RTMP %d est déjà utilisé (autre MediaMTX ou OBS ?)", s.cfg.RtmpPort)
	}

	cfgPath, err := s.writeConfig()
	if err != nil {
		return fmt.Errorf("écriture de la configuration MediaMTX : %w", err)
	}

	cmd := exec.Command(s.cfg.Binary, cfgPath)
	// MediaMTX écrit ses fichiers auto-générés (auto.crt/auto.key) dans son
	// répertoire courant : on le fixe sur les données de l'app.
	cmd.Dir = s.cfg.Dir
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("lancement de MediaMTX : %w", err)
	}

	s.cmd = cmd
	s.err = ""
	s.done = make(chan struct{})

	logf := s.cfg.Logf
	if logf == nil {
		logf = func(string) {}
	}
	go pipeLines(stdout, logf)
	go pipeLines(stderr, logf)

	exited := s.done
	go func() {
		err := cmd.Wait()
		s.mu.Lock()
		if s.cmd == cmd {
			s.cmd = nil
			if err != nil {
				s.err = err.Error()
			}
		}
		cb := s.onExit
		s.mu.Unlock()
		if err != nil {
			logf("MediaMTX arrêté : " + err.Error())
		}
		if cb != nil {
			cb(s.err)
		}
		close(exited)
	}()

	if !waitPort(s.cfg.RtmpPort, 12*time.Second, exited) {
		_ = s.stopLocked()
		return fmt.Errorf("MediaMTX n'a pas ouvert le port %d", s.cfg.RtmpPort)
	}
	logf(fmt.Sprintf("MediaMTX prêt — RTMP sur :%d", s.cfg.RtmpPort))
	return nil
}

func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stopLocked()
}

func (s *Server) stopLocked() error {
	if s.cmd == nil {
		return nil
	}
	cmd := s.cmd
	_ = cmd.Process.Signal(syscall.SIGINT)
	select {
	case <-s.done:
	case <-time.After(5 * time.Second):
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		select {
		case <-s.done:
		case <-time.After(2 * time.Second):
		}
	}
	s.cmd = nil
	return nil
}

// Paths queries the MediaMTX API so the UI can tell "the server is up" apart
// from "the camera is actually streaming to us".
func (s *Server) Paths(ctx context.Context) (map[string]PathStat, error) {
	out := map[string]PathStat{}
	url := fmt.Sprintf("http://127.0.0.1:%d/v3/paths/list", s.cfg.ApiPort)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return out, err
	}
	cl := &http.Client{Timeout: 2 * time.Second}
	resp, err := cl.Do(req)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return out, fmt.Errorf("API MediaMTX : HTTP %d", resp.StatusCode)
	}
	var payload struct {
		Items []struct {
			Name          string   `json:"name"`
			Ready         bool     `json:"ready"`
			Tracks        []string `json:"tracks"`
			BytesReceived int64    `json:"bytesReceived"`
			Readers       []struct {
				Type string `json:"type"`
			} `json:"readers"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return out, err
	}
	for _, it := range payload.Items {
		out[it.Name] = PathStat{
			Name:          it.Name,
			Ready:         it.Ready,
			Tracks:        it.Tracks,
			BytesReceived: it.BytesReceived,
			Readers:       len(it.Readers),
		}
	}
	return out, nil
}

func pipeLines(r interface{ Read([]byte) (int, error) }, logf func(string)) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		logf(line)
	}
}

func portFree(port int) bool {
	ln, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}

func waitPort(port int, timeout time.Duration, exited <-chan struct{}) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-exited:
			return false
		default:
		}
		conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)), 400*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return true
		}
		time.Sleep(150 * time.Millisecond)
	}
	return false
}

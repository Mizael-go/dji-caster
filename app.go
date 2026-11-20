package main

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"djicaster/internal/bins"
	"djicaster/internal/config"
	"djicaster/internal/logs"
	"djicaster/internal/media"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// ---------------------------------------------------------------- view models

type LogLine struct {
	Scope string `json:"scope"`
	Level string `json:"level"`
	Text  string `json:"text"`
	Time  int64  `json:"time"`
}

type EnvView struct {
	OS            string `json:"os"`
	Arch          string `json:"arch"`
	LocalIP       string `json:"localIP"`
	DataDir       string `json:"dataDir"`
	MediaMtxPath  string `json:"mediaMtxPath"`
	FFmpegPath    string `json:"ffmpegPath"`
	MediaMtxReady bool   `json:"mediaMtxReady"`
	FFmpegReady   bool   `json:"ffmpegReady"`
}

type ServerView struct {
	Running  bool   `json:"running"`
	Err      string `json:"err"`
	RtmpPort int    `json:"rtmpPort"`
	ApiPort  int    `json:"apiPort"`
}

type PathView struct {
	Ready   bool     `json:"ready"`
	Codecs  []string `json:"codecs"`
	Bytes   int64    `json:"bytes"`
	Readers int      `json:"readers"`
}

type StreamView struct {
	// RTMP is only the ingest: the camera or the Android relay pushes there.
	Host      string `json:"host"`
	RtmpPort  int    `json:"rtmpPort"`
	DirectURL string `json:"directURL"`

	// WebRTC is the only playback output.
	WebrtcPort     int    `json:"webrtcPort"`
	WebrtcURL      string `json:"webrtcURL"`
	WebrtcLocalURL string `json:"webrtcLocalURL"`
	WebrtcReady    bool   `json:"webrtcReady"`
	WebrtcAudio    bool   `json:"webrtcAudio"`
	EncoderErr     string `json:"encoderErr"`

	Incoming PathView `json:"incoming"`
	Webrtc   PathView `json:"webrtc"`
}

type Snapshot struct {
	Settings    config.Settings `json:"settings"`
	Environment EnvView         `json:"environment"`
	Server      ServerView      `json:"server"`
	Stream      StreamView      `json:"stream"`
}

// ------------------------------------------------------------------------ App

type App struct {
	ctx context.Context

	mu       sync.Mutex
	settings config.Settings
	dataDir  string
	binDir   string

	server  *media.Server
	encoder *media.Transcoder
	logs    *logs.Buffer

	emitMu   sync.Mutex
	lastEmit time.Time
	pending  bool
}

func NewApp() *App {
	return &App{
		settings: config.Defaults(),
		logs:     logs.New(400),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	dir, err := config.Dir()
	if err != nil {
		a.log("app", "répertoire de configuration indisponible : "+err.Error())
	}
	a.dataDir = dir
	binDir, err := config.BinDir()
	if err != nil {
		a.log("app", "répertoire des binaires indisponible : "+err.Error())
	}
	a.binDir = binDir

	settings, err := config.Load()
	if err != nil {
		a.log("app", "configuration illisible, valeurs par défaut : "+err.Error())
	}
	a.mu.Lock()
	a.settings = settings
	a.mu.Unlock()
	a.resolveBinaries()

	a.server = media.NewServer(media.ServerConfig{
		Binary:   a.mediaMtxPath(),
		Dir:      a.dataDir,
		RtmpPort: settings.RtmpPort,
		ApiPort:  settings.ApiPort,
		Logf:     func(s string) { a.log("mediamtx", s) },
	})
	a.server.SetOnExit(func(reason string) {
		if reason != "" {
			a.log("mediamtx", "MediaMTX s'est arrêté : "+reason)
		}
		a.emit()
	})
	a.encoder = media.NewTranscoder(a.ffmpegPath(), func(s string) { a.log("ffmpeg", s) })

	a.syncMedia()
	go a.mediaWatcher()
}

func (a *App) shutdown(ctx context.Context) {
	if a.encoder != nil {
		a.encoder.Stop()
	}
	if a.server != nil {
		_ = a.server.Stop()
	}
}

// ------------------------------------------------------------------- bindings

func (a *App) Snapshot() Snapshot {
	a.mu.Lock()
	settings := a.settings
	a.mu.Unlock()

	paths := map[string]media.PathStat{}
	if a.server != nil && a.server.Running() {
		ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
		if p, err := a.server.Paths(ctx); err == nil {
			paths = p
		}
		cancel()
	}
	in := paths[media.PathIn]
	web := paths[media.PathWebrtc]

	host := media.LocalIP()
	direct := media.RtmpURL(host, settings.RtmpPort, media.PathIn)

	serverErr := ""
	running := false
	if a.server != nil {
		running = a.server.Running()
		serverErr = a.server.Err()
	}
	encoderErr := ""
	if a.encoder != nil {
		encoderErr = a.encoder.Err()
	}

	return Snapshot{
		Settings: settings,
		Environment: EnvView{
			OS:            goos(),
			Arch:          goarch(),
			LocalIP:       host,
			DataDir:       a.dataDir,
			MediaMtxPath:  a.mediaMtxPath(),
			FFmpegPath:    a.ffmpegPath(),
			MediaMtxReady: a.mediaMtxPath() != "",
			FFmpegReady:   a.ffmpegPath() != "",
		},
		Server: ServerView{
			Running:  running,
			Err:      serverErr,
			RtmpPort: settings.RtmpPort,
			ApiPort:  settings.ApiPort,
		},
		Stream: StreamView{
			Host:           host,
			RtmpPort:       settings.RtmpPort,
			DirectURL:      direct,
			WebrtcPort:     media.WebrtcPort,
			WebrtcURL:      media.WebrtcURL(host, media.WebrtcPort, media.PathWebrtc),
			WebrtcLocalURL: media.WebrtcURL("127.0.0.1", media.WebrtcPort, media.PathWebrtc),
			WebrtcReady:    web.Ready,
			WebrtcAudio:    media.HasAudio(web.Tracks),
			EncoderErr:     encoderErr,
			Incoming:       toPathView(in),
			Webrtc:         toPathView(web),
		},
	}
}

func (a *App) Logs() []LogLine {
	out := []LogLine{}
	for _, l := range a.logs.Snapshot() {
		out = append(out, LogLine{Scope: l.Scope, Level: l.Level, Text: l.Text, Time: l.Time})
	}
	return out
}

func (a *App) ClearLogs() {
	a.logs.Clear()
	a.emit()
}

func (a *App) SaveSettings(next config.Settings) error {
	if next.RtmpPort <= 0 || next.RtmpPort > 65535 {
		return fmt.Errorf("port RTMP invalide")
	}
	if next.ApiPort <= 0 || next.ApiPort > 65535 || next.ApiPort == next.RtmpPort {
		return fmt.Errorf("port d'API invalide (et différent du port RTMP)")
	}
	if err := config.Save(next); err != nil {
		return fmt.Errorf("sauvegarde impossible : %w", err)
	}

	a.mu.Lock()
	a.settings = next
	a.mu.Unlock()

	// Rebuild the media stack with the new ports.
	if a.server != nil {
		_ = a.server.Stop()
	}
	a.encoder.Stop()
	a.resolveBinaries()
	a.server = media.NewServer(media.ServerConfig{
		Binary:   a.mediaMtxPath(),
		Dir:      a.dataDir,
		RtmpPort: next.RtmpPort,
		ApiPort:  next.ApiPort,
		Logf:     func(s string) { a.log("mediamtx", s) },
	})
	a.server.SetOnExit(func(reason string) { a.emit() })
	a.encoder.SetBinary(a.ffmpegPath())
	a.syncMedia()
	a.emit()
	return nil
}

func (a *App) StartServer() error {
	if err := a.ensureMediaMtx(context.Background()); err != nil {
		return err
	}
	return a.syncMediaErr()
}

func (a *App) StopServer() error {
	a.encoder.Stop()
	if a.server != nil {
		if err := a.server.Stop(); err != nil {
			return err
		}
	}
	a.log("mediamtx", "serveur RTMP arrêté")
	a.emit()
	return nil
}

// InstallTools downloads MediaMTX into the app data dir.
func (a *App) InstallTools() error {
	a.log("app", "installation des outils…")
	if _, err := bins.EnsureMediaMtx(context.Background(), a.binDir, func(s string) { a.log("app", s) }); err != nil {
		a.log("app", err.Error())
		a.emit()
		return err
	}
	a.resolveBinaries()
	a.encoder.SetBinary(a.ffmpegPath())
	if a.server != nil {
		a.server.Stop()
	}
	a.server = media.NewServer(media.ServerConfig{
		Binary:   a.mediaMtxPath(),
		Dir:      a.dataDir,
		RtmpPort: a.portRtmp(),
		ApiPort:  a.portAPI(),
		Logf:     func(s string) { a.log("mediamtx", s) },
	})
	a.server.SetOnExit(func(reason string) { a.emit() })
	a.log("app", "outils prêts")
	a.emit()
	return nil
}

// OpenLogsFolder reveals the app data dir in the file manager.
func (a *App) OpenLogsFolder() {
	if a.dataDir != "" {
		wailsruntime.BrowserOpenURL(a.ctx, "file://"+a.dataDir)
	}
}

// ------------------------------------------------------------------- internals

func (a *App) resolveBinaries() {
	a.mu.Lock()
	defer a.mu.Unlock()
	// MediaMTX: our own copy only, never the one installed system-wide.
	mtx := a.settings.MediaMtx
	if bins.LookPath(mtx) == "" {
		mtx = filepath.Join(a.binDir, bins.MediaMtxName())
		if bins.LookPath(mtx) == "" {
			mtx = ""
		}
	}
	ff := bins.LookPath(a.settings.FFmpeg)
	if ff == "" {
		ff = bins.LookPath("ffmpeg")
	}
	a.settings.MediaMtx = mtx
	a.settings.FFmpeg = ff
}

func (a *App) mediaMtxPath() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.settings.MediaMtx
}

func (a *App) ffmpegPath() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.settings.FFmpeg
}

func (a *App) portRtmp() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.settings.RtmpPort
}

func (a *App) portAPI() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.settings.ApiPort
}

func (a *App) ensureMediaMtx(ctx context.Context) error {
	if a.mediaMtxPath() != "" {
		return nil
	}
	if a.binDir == "" {
		return fmt.Errorf("répertoire des binaires indisponible")
	}
	if _, err := bins.EnsureMediaMtx(ctx, a.binDir, func(s string) { a.log("app", s) }); err != nil {
		return err
	}
	a.resolveBinaries()
	a.mu.Lock()
	settings := a.settings
	a.mu.Unlock()
	a.server = media.NewServer(media.ServerConfig{
		Binary:   settings.MediaMtx,
		Dir:      a.dataDir,
		RtmpPort: settings.RtmpPort,
		ApiPort:  settings.ApiPort,
		Logf:     func(s string) { a.log("mediamtx", s) },
	})
	a.server.SetOnExit(func(reason string) { a.emit() })
	return nil
}

// syncMedia reconciles the running processes with the desired state.
func (a *App) syncMedia() {
	if err := a.syncMediaErr(); err != nil {
		a.log("app", err.Error())
	}
}

func (a *App) syncMediaErr() error {
	if a.server == nil {
		return nil
	}
	if a.server.Running() {
		return a.syncEncoder()
	}
	// Auto-start: the RTMP target must exist before the camera is asked to push.
	if err := a.server.Start(); err != nil {
		return err
	}
	return a.syncEncoder()
}

// syncEncoder starts the single ffmpeg pipeline that publishes PathWebrtc
// (H.264 + Opus). It only runs while a source is actually arriving, so it
// never spins on a dead input. Called on every state change and by the watcher.
func (a *App) syncEncoder() error {
	if a.encoder == nil {
		return nil
	}
	a.mu.Lock()
	settings := a.settings
	a.mu.Unlock()
	if a.server == nil || !a.server.Running() {
		a.encoder.Stop()
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1200*time.Millisecond)
	paths, err := a.server.Paths(ctx)
	cancel()
	if err != nil || !paths[media.PathIn].Ready {
		a.encoder.Stop()
		return nil
	}

	in := media.RtspURL("127.0.0.1", media.RtspPort, media.PathIn)
	out := media.RtmpURL("127.0.0.1", settings.RtmpPort, media.PathWebrtc)
	return a.encoder.Start(in, out)
}

// mediaWatcher keeps the UI and the encoder in sync with the real stream: it
// notices when the camera starts or stops pushing.
func (a *App) mediaWatcher() {
	t := time.NewTicker(2 * time.Second)
	defer t.Stop()
	wasReady := false
	for {
		select {
		case <-a.ctx.Done():
			return
		case <-t.C:
			if a.server == nil || !a.server.Running() {
				wasReady = false
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), 1200*time.Millisecond)
			paths, err := a.server.Paths(ctx)
			cancel()
			if err != nil {
				continue
			}
			ready := paths[media.PathIn].Ready
			if ready != wasReady {
				wasReady = ready
				a.emit()
			}
			if err := a.syncEncoder(); err != nil {
				a.log("ffmpeg", err.Error())
			}
		}
	}
}

func (a *App) log(scope, text string) {
	if a.logs == nil {
		return
	}
	a.logs.Add(scope, text)
}

// emit notifies the frontend that something changed. Calls are coalesced so a
// burst of log lines does not flood the webview.
func (a *App) emit() {
	if a.ctx == nil {
		return
	}
	a.emitMu.Lock()
	if time.Since(a.lastEmit) < 200*time.Millisecond {
		if a.pending {
			a.emitMu.Unlock()
			return
		}
		a.pending = true
		a.emitMu.Unlock()
		time.AfterFunc(220*time.Millisecond, func() {
			a.emitMu.Lock()
			a.pending = false
			a.lastEmit = time.Now()
			a.emitMu.Unlock()
			wailsruntime.EventsEmit(a.ctx, "state", "change")
		})
		return
	}
	a.lastEmit = time.Now()
	a.pending = false
	a.emitMu.Unlock()
	wailsruntime.EventsEmit(a.ctx, "state", "change")
}

func toPathView(p media.PathStat) PathView {
	// Tracks peut être nil (aucun flux) : on renvoie toujours un tableau pour
	// que le frontend n'ait jamais à gérer un null.
	tracks := p.Tracks
	if tracks == nil {
		tracks = []string{}
	}
	return PathView{Ready: p.Ready, Codecs: tracks, Bytes: p.BytesReceived, Readers: p.Readers}
}

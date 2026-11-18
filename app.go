package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"djicaster/internal/bins"
	"djicaster/internal/ble"
	"djicaster/internal/config"
	"djicaster/internal/djictl"
	"djicaster/internal/logs"
	"djicaster/internal/media"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// ---------------------------------------------------------------- view models

type DeviceView struct {
	Address   string `json:"address"`
	Name      string `json:"name"`
	RSSI      int    `json:"rssi"`
	Connected bool   `json:"connected"`
	Paired    bool   `json:"paired"`
	IsDJI     bool   `json:"isDJI"`
	LastSeen  int64  `json:"lastSeen"`
}

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
	WifiSSID      string `json:"wifiSSID"`
	DataDir       string `json:"dataDir"`
	MediaMtxPath  string `json:"mediaMtxPath"`
	DjiCtlPath    string `json:"djiCtlPath"`
	FFmpegPath    string `json:"ffmpegPath"`
	MediaMtxReady bool   `json:"mediaMtxReady"`
	DjiCtlReady   bool   `json:"djiCtlReady"`
	FFmpegReady   bool   `json:"ffmpegReady"`
}

type ScanView struct {
	Scanning bool   `json:"scanning"`
	Err      string `json:"err"`
}

// RadioView exposes the kernel radio switch so the UI can offer to unblock it.
type RadioView struct {
	Present bool `json:"present"`
	Soft    bool `json:"soft"`
	Hard    bool `json:"hard"`
}

type ServerView struct {
	Running  bool   `json:"running"`
	Err      string `json:"err"`
	RtmpPort int    `json:"rtmpPort"`
	ApiPort  int    `json:"apiPort"`
}

type SessionView struct {
	State      string `json:"state"`
	Err        string `json:"err"`
	Hint       string `json:"hint"`
	DeviceAddr string `json:"deviceAddr"`
	DeviceName string `json:"deviceName"`
	Command    string `json:"command"`
}

type PathView struct {
	Ready   bool     `json:"ready"`
	Codecs  []string `json:"codecs"`
	Bytes   int64    `json:"bytes"`
	Readers int      `json:"readers"`
}

type StreamView struct {
	Host         string   `json:"host"`
	RtmpPort     int      `json:"rtmpPort"`
	DirectURL    string   `json:"directURL"`
	ObsURL       string   `json:"obsURL"`
	LocalURL     string   `json:"localURL"`
	Transcode    bool     `json:"transcode"`
	TranscodeErr string   `json:"transcodeErr"`
	Transcoding  bool     `json:"transcoding"`
	Incoming     PathView `json:"incoming"`
	Outgoing     PathView `json:"outgoing"`
}

type Snapshot struct {
	Settings    config.Settings `json:"settings"`
	Environment EnvView         `json:"environment"`
	Radio       RadioView       `json:"radio"`
	Scan        ScanView        `json:"scan"`
	Devices     []DeviceView    `json:"devices"`
	Server      ServerView      `json:"server"`
	Session     SessionView     `json:"session"`
	Stream      StreamView      `json:"stream"`
}

type ConnectRequest struct {
	Address    string `json:"address"`
	Name       string `json:"name"`
	SSID       string `json:"ssid"`
	PSK        string `json:"psk"`
	Resolution string `json:"resolution"`
	Bitrate    int    `json:"bitrate"`
	FPS        int    `json:"fps"`
}

// ------------------------------------------------------------------------ App

type App struct {
	ctx context.Context

	mu       sync.Mutex
	settings config.Settings
	dataDir  string
	binDir   string

	scanner    *ble.Scanner
	server     *media.Server
	transcoder *media.Transcoder
	logs       *logs.Buffer
	runner     *djictl.Runner

	sessionState  string
	sessionErr    string
	sessionHint   string
	sessionDevice string
	sessionName   string
	sessionCmd    string
	djiLogTail    []string

	emitMu   sync.Mutex
	lastEmit time.Time
	pending  bool
}

func NewApp() *App {
	a := &App{
		settings:     config.Defaults(),
		logs:         logs.New(400),
		runner:       djictl.NewRunner(),
		sessionState: "idle",
	}
	a.scanner = ble.NewScanner(a.emit)
	return a
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
		a.mu.Lock()
		if reason != "" && a.sessionState != "error" {
			a.sessionErr = "MediaMTX s'est arrêté : " + reason
		}
		a.mu.Unlock()
		a.emit()
	})
	a.transcoder = media.NewTranscoder(a.ffmpegPath(), func(s string) { a.log("ffmpeg", s) })

	a.syncMedia()
	go a.bleUpdater()
	go a.mediaWatcher()
}

func (a *App) shutdown(ctx context.Context) {
	_ = a.scanner.Stop()
	a.runner.Stop()
	if a.transcoder != nil {
		a.transcoder.Stop()
	}
	if a.server != nil {
		_ = a.server.Stop()
	}
}

// bleUpdater periodically re-emits the device list so RSSI changes show up even
// when BlueZ is quiet.
func (a *App) bleUpdater() {
	t := time.NewTicker(2 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-a.ctx.Done():
			return
		case <-t.C:
			if a.scanner.Scanning() {
				// Borne la mémoire : un appareil absent depuis 2 min disparaît.
				a.scanner.Prune(2 * time.Minute)
				a.emit()
			}
		}
	}
}

// ------------------------------------------------------------------- bindings

func (a *App) Snapshot() Snapshot {
	a.mu.Lock()
	settings := a.settings
	session := SessionView{
		State:      a.sessionState,
		Err:        a.sessionErr,
		Hint:       a.sessionHint,
		DeviceAddr: a.sessionDevice,
		DeviceName: a.sessionName,
		Command:    a.sessionCmd,
	}
	a.mu.Unlock()

	devices := []DeviceView{}
	for _, d := range a.scanner.DeviceList() {
		devices = append(devices, DeviceView{
			Address:   d.Address,
			Name:      d.Name,
			RSSI:      d.RSSI,
			Connected: d.Connected,
			Paired:    d.Paired,
			IsDJI:     d.IsDJI,
			LastSeen:  d.LastSeen,
		})
	}

	paths := map[string]media.PathStat{}
	if a.server != nil && a.server.Running() {
		ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
		if p, err := a.server.Paths(ctx); err == nil {
			paths = p
		}
		cancel()
	}
	in := paths[media.PathIn]
	out := paths[media.PathOut]

	host := media.LocalIP()
	direct := media.RtmpURL(host, settings.RtmpPort, media.PathIn)
	obs := direct
	if settings.Transcode {
		obs = media.RtmpURL(host, settings.RtmpPort, media.PathOut)
	}

	if in.Ready && session.State != "error" {
		session.State = "live"
	} else if session.State == "live" {
		session.State = "waiting"
	}

	serverErr := ""
	running := false
	if a.server != nil {
		running = a.server.Running()
		serverErr = a.server.Err()
	}
	transcoding := false
	transcodeErr := ""
	if a.transcoder != nil {
		transcoding = a.transcoder.Running()
		transcodeErr = a.transcoder.Err()
	}

	return Snapshot{
		Settings: settings,
		Environment: EnvView{
			OS:            goos(),
			Arch:          goarch(),
			LocalIP:       host,
			WifiSSID:      media.WifiSSID(),
			DataDir:       a.dataDir,
			MediaMtxPath:  a.mediaMtxPath(),
			DjiCtlPath:    a.djiCtlPath(),
			FFmpegPath:    a.ffmpegPath(),
			MediaMtxReady: a.mediaMtxPath() != "",
			DjiCtlReady:   a.djiCtlPath() != "",
			FFmpegReady:   a.ffmpegPath() != "",
		},
		Radio:   radioView(),
		Scan:    ScanView{Scanning: a.scanner.Scanning(), Err: a.scanner.Err()},
		Devices: devices,
		Server: ServerView{
			Running:  running,
			Err:      serverErr,
			RtmpPort: settings.RtmpPort,
			ApiPort:  settings.ApiPort,
		},
		Session: session,
		Stream: StreamView{
			Host:         host,
			RtmpPort:     settings.RtmpPort,
			DirectURL:    direct,
			ObsURL:       obs,
			LocalURL:     media.RtmpURL("127.0.0.1", settings.RtmpPort, obsPath(settings.Transcode)),
			Transcode:    settings.Transcode,
			Transcoding:  transcoding,
			TranscodeErr: transcodeErr,
			Incoming:     toPathView(in),
			Outgoing:     toPathView(out),
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
	if next.Networks == nil {
		// Le frontend ne renvoie pas la liste : on garde celle déjà chargée.
		a.mu.Lock()
		next.Networks = a.settings.Networks
		a.mu.Unlock()
	}
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

	// Rebuild the media stack with the new ports / paths.
	if a.server != nil {
		_ = a.server.Stop()
	}
	a.transcoder.Stop()
	a.resolveBinaries()
	a.server = media.NewServer(media.ServerConfig{
		Binary:   a.mediaMtxPath(),
		Dir:      a.dataDir,
		RtmpPort: next.RtmpPort,
		ApiPort:  next.ApiPort,
		Logf:     func(s string) { a.log("mediamtx", s) },
	})
	a.server.SetOnExit(func(reason string) { a.emit() })
	a.transcoder.SetBinary(a.ffmpegPath())
	a.syncMedia()
	a.emit()
	return nil
}

func (a *App) StartScan() error {
	// rfkill soft-block : on le lève d'abord (no-op si déjà débloqué).
	if err := ble.Unblock(); err != nil {
		a.log("ble", err.Error())
		a.emit()
		return err
	}
	if err := a.scanner.Start(); err != nil {
		a.log("ble", err.Error())
		a.emit()
		return err
	}
	a.log("ble", "scan Bluetooth démarré")
	a.emit()
	return nil
}

func (a *App) StopScan() error {
	_ = a.scanner.Stop()
	a.log("ble", "scan Bluetooth arrêté")
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
	a.transcoder.Stop()
	if a.server != nil {
		if err := a.server.Stop(); err != nil {
			return err
		}
	}
	a.log("mediamtx", "serveur RTMP arrêté")
	a.emit()
	return nil
}

func (a *App) SetTranscode(enabled bool) error {
	a.mu.Lock()
	a.settings.Transcode = enabled
	settings := a.settings
	a.mu.Unlock()

	if err := config.Save(settings); err != nil {
		return fmt.Errorf("sauvegarde impossible : %w", err)
	}
	a.syncMedia()
	a.emit()
	return nil
}

// StartConnect launches djictl in the background. The UI follows the progress
// through the session state and the log stream.
func (a *App) StartConnect(req ConnectRequest) error {
	addr := strings.TrimSpace(req.Address)
	ssid := strings.TrimSpace(req.SSID)
	if addr == "" {
		return fmt.Errorf("aucun appareil sélectionné")
	}
	if ssid == "" {
		return fmt.Errorf("le SSID Wi-Fi est requis")
	}
	if a.runner.Running() {
		return fmt.Errorf("une connexion est déjà en cours")
	}
	// La découverte BlueZ est coupée pendant la connexion : certains
	// contrôleurs la refusent en parallèle d'un appairage.
	_ = a.scanner.Stop()
	bin := a.djiCtlPath()
	if bin == "" {
		return fmt.Errorf("djictl n'est pas installé (bouton « Installer les outils »)")
	}
	if err := a.ensureMediaMtx(context.Background()); err != nil {
		return err
	}
	if err := a.syncMediaErr(); err != nil {
		return err
	}

	a.mu.Lock()
	// Chaque connexion réussie mémorise le réseau (SSID + PSK) et les
	// réglages d'encodage pour la fois suivante : c'est la liste des profils
	// du formulaire.
	a.settings.UpsertNetwork(ssid, req.PSK)
	a.settings.LastDevice = addr
	// Champs absents (ancien client) : on garde les réglages mémorisés.
	if strings.TrimSpace(req.Resolution) != "" {
		a.settings.StreamResolution = strings.TrimSpace(req.Resolution)
	}
	if req.Bitrate > 0 {
		a.settings.StreamBitrate = req.Bitrate
	}
	if req.FPS > 0 {
		a.settings.StreamFPS = req.FPS
	}
	a.settings.NormalizeStream()
	settings := a.settings
	a.sessionState = "starting"
	a.sessionErr = ""
	a.sessionHint = ""
	a.sessionDevice = addr
	a.sessionName = req.Name
	a.sessionCmd = djictl.CommandLine(bin, addr, ssid, req.PSK,
		media.RtmpURL(media.LocalIP(), settings.RtmpPort, media.PathIn),
		settings.StreamResolution, settings.StreamBitrate, settings.StreamFPS)
	a.djiLogTail = nil
	a.mu.Unlock()
	_ = config.Save(settings)

	rtmpURL := media.RtmpURL(media.LocalIP(), settings.RtmpPort, media.PathIn)
	a.log("djictl", fmt.Sprintf("connexion de %s (%s) au Wi-Fi « %s » → %s (%s, %d kbps, %d fps)",
		req.Name, addr, ssid, rtmpURL,
		settings.StreamResolution, settings.StreamBitrate, settings.StreamFPS))
	a.emit()

	go func() {
		err := a.runner.Run(a.ctx, bin, addr, ssid, req.PSK, rtmpURL,
			settings.StreamResolution, settings.StreamBitrate, settings.StreamFPS,
			func(line string) {
				a.mu.Lock()
				a.djiLogTail = append(a.djiLogTail, line)
				if len(a.djiLogTail) > 40 {
					a.djiLogTail = a.djiLogTail[len(a.djiLogTail)-40:]
				}
				a.mu.Unlock()
				a.log("djictl", line)
			})

		a.mu.Lock()
		if err != nil {
			a.sessionState = "error"
			a.sessionErr = err.Error()
			a.sessionHint = djictl.Hint(a.djiLogTail)
		} else {
			a.sessionState = "waiting"
		}
		a.mu.Unlock()
		a.emit()
	}()
	return nil
}

func (a *App) CancelConnect() {
	a.runner.Stop()
	a.mu.Lock()
	if a.sessionState == "starting" {
		a.sessionState = "idle"
	}
	a.mu.Unlock()
	a.emit()
}

func (a *App) ResetSession() {
	a.mu.Lock()
	a.sessionState = "idle"
	a.sessionErr = ""
	a.sessionHint = ""
	a.sessionCmd = ""
	a.mu.Unlock()
	a.emit()
}

// UnblockBluetooth clears the rfkill soft block (Linux) and reports why it
// failed when the user lacks the permission.
func (a *App) UnblockBluetooth() error {
	if err := ble.Unblock(); err != nil {
		a.log("ble", err.Error())
		a.emit()
		return err
	}
	a.log("ble", "radio Bluetooth débloquée")
	a.emit()
	return nil
}

// SaveNetwork stores a Wi-Fi profile without connecting the camera.
func (a *App) SaveNetwork(ssid, psk string) error {
	ssid = strings.TrimSpace(ssid)
	if ssid == "" {
		return fmt.Errorf("le SSID est requis")
	}
	a.mu.Lock()
	a.settings.UpsertNetwork(ssid, psk)
	settings := a.settings
	a.mu.Unlock()
	if err := config.Save(settings); err != nil {
		return fmt.Errorf("sauvegarde impossible : %w", err)
	}
	a.emit()
	return nil
}

func (a *App) DeleteNetwork(ssid string) error {
	a.mu.Lock()
	removed := a.settings.RemoveNetwork(ssid)
	settings := a.settings
	a.mu.Unlock()
	if !removed {
		return fmt.Errorf("réseau inconnu : %s", ssid)
	}
	if err := config.Save(settings); err != nil {
		return fmt.Errorf("sauvegarde impossible : %w", err)
	}
	a.emit()
	return nil
}

func radioView() RadioView {
	st := ble.ReadRfkill()
	return RadioView{Present: st.Present, Soft: st.Soft, Hard: st.Hard}
}

// InstallTools downloads MediaMTX and builds djictl into the app data dir.
func (a *App) InstallTools() error {
	a.log("app", "installation des outils…")
	if _, err := bins.EnsureMediaMtx(context.Background(), a.binDir, func(s string) { a.log("app", s) }); err != nil {
		a.log("app", err.Error())
		a.emit()
		return err
	}
	if _, err := bins.EnsureDjictl(context.Background(), a.binDir, func(s string) { a.log("app", s) }); err != nil {
		a.log("app", err.Error())
		a.emit()
		return err
	}
	a.resolveBinaries()
	a.transcoder.SetBinary(a.ffmpegPath())
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
	dji := a.settings.DjiCtl
	if bins.LookPath(dji) == "" {
		dji = filepath.Join(a.binDir, bins.DjiCtlName())
		if bins.LookPath(dji) == "" {
			dji = ""
		}
	}
	ff := bins.LookPath(a.settings.FFmpeg)
	if ff == "" {
		ff = bins.LookPath("ffmpeg")
	}
	a.settings.MediaMtx = mtx
	a.settings.DjiCtl = dji
	a.settings.FFmpeg = ff
}

func (a *App) mediaMtxPath() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.settings.MediaMtx
}

func (a *App) djiCtlPath() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.settings.DjiCtl
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
		if err := a.syncTranscode(); err != nil {
			return err
		}
		return nil
	}
	// Auto-start: the RTMP target must exist before the camera is asked to push.
	if err := a.server.Start(); err != nil {
		a.mu.Lock()
		a.sessionErr = err.Error()
		a.mu.Unlock()
		return err
	}
	return a.syncTranscode()
}

// syncTranscode starts ffmpeg only when a camera stream is actually arriving,
// so it never spins on a dead input. It is called on every state change and by
// the media watcher.
func (a *App) syncTranscode() error {
	a.mu.Lock()
	settings := a.settings
	a.mu.Unlock()
	if !settings.Transcode || a.server == nil || !a.server.Running() {
		a.transcoder.Stop()
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1200*time.Millisecond)
	paths, err := a.server.Paths(ctx)
	cancel()
	if err != nil || !paths[media.PathIn].Ready {
		a.transcoder.Stop()
		return nil
	}

	in := media.RtspURL("127.0.0.1", media.RtspPort, media.PathIn)
	out := media.RtmpURL("127.0.0.1", settings.RtmpPort, media.PathOut)
	return a.transcoder.Start(in, out)
}

// mediaWatcher keeps the UI and the transcoder in sync with the real stream:
// it notices when the camera starts or stops pushing.
func (a *App) mediaWatcher() {
	t := time.NewTicker(2 * time.Second)
	defer t.Stop()
	wasReady := false
	for {
		select {
		case <-a.ctx.Done():
			return
		case <-t.C:
			a.mu.Lock()
			transcode := a.settings.Transcode
			a.mu.Unlock()

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
			if transcode {
				if err := a.syncTranscode(); err != nil {
					a.log("ffmpeg", err.Error())
				}
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

func obsPath(transcode bool) string {
	if transcode {
		return media.PathOut
	}
	return media.PathIn
}

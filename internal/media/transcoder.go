package media

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Transcoder republishes the camera stream as H.264 so OBS can consume it even
// when the DJI only offers HEVC. ffmpeg is supervised: it is restarted while
// enabled and the input is not ready yet.
type Transcoder struct {
	mu      sync.Mutex
	binary  string
	logf    func(string)
	cancel  context.CancelFunc
	cmd     *exec.Cmd
	running bool
	lastErr string
}

func NewTranscoder(binary string, logf func(string)) *Transcoder {
	return &Transcoder{binary: binary, logf: logf}
}

func (t *Transcoder) Running() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.running
}

func (t *Transcoder) Err() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.lastErr
}

func (t *Transcoder) SetBinary(binary string) {
	t.mu.Lock()
	t.binary = binary
	t.mu.Unlock()
}

func (t *Transcoder) Start(input, output string) error {
	t.mu.Lock()
	if t.running {
		t.mu.Unlock()
		return nil
	}
	if t.binary == "" {
		t.mu.Unlock()
		return fmt.Errorf("ffmpeg introuvable : installez-le (apt install ffmpeg)")
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel
	t.running = true
	t.lastErr = ""
	t.mu.Unlock()

	go t.supervise(ctx, output, input)
	return nil
}

func (t *Transcoder) Stop() {
	t.mu.Lock()
	cancel := t.cancel
	cmd := t.cmd
	t.cancel = nil
	t.running = false
	t.cmd = nil
	t.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	stopCmd(cmd)
}

func (t *Transcoder) supervise(ctx context.Context, output, input string) {
	t.logf("Republication H.264 activée (re-timestamping + transcodage)")
	for {
		if ctx.Err() != nil {
			return
		}
		cmd := exec.CommandContext(ctx, t.binary, transcodeArgs(input, output)...)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			t.fail(err)
			return
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			t.fail(err)
			return
		}
		if err := cmd.Start(); err != nil {
			t.fail(err)
			return
		}
		t.mu.Lock()
		t.cmd = cmd
		t.mu.Unlock()

		go pipeLines(stdout, t.logf)
		go pipeLines(stderr, t.logf)

		err = cmd.Wait()
		t.mu.Lock()
		if t.cmd == cmd {
			t.cmd = nil
		}
		running := t.running
		t.mu.Unlock()

		if ctx.Err() != nil || !running {
			return
		}
		if err != nil {
			t.logf("ffmpeg : " + err.Error() + " — nouvelle tentative dans 3s")
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
		}
	}
}

func (t *Transcoder) fail(err error) {
	t.mu.Lock()
	t.running = false
	t.lastErr = err.Error()
	t.mu.Unlock()
	t.logf("Transcodage impossible : " + err.Error())
}

func stopCmd(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGINT)
	time.Sleep(300 * time.Millisecond)
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}

func transcodeArgs(input, output string) []string {
	return []string{
		"-hide_banner",
		"-nostdin",
		"-loglevel", "warning",
		"-flags", "low_delay",
		// Les timestamps du flux source (caméra DJI ou relais Android) sont
		// cassés : DTS égaux/répétés et audio démarrant avant la vidéo. Sans
		// restampage, OBS gèle sur les DTS non monotones. L'horloge murale
		// donne des timestamps monotones et recale l'A/V à l'arrivée des
		// paquets, quel que soit le décalage initial de la source.
		"-use_wallclock_as_timestamps", "1",
		// +genpts complète les pts manquants ; avec l'horloge murale il ne
		// dérive plus (le flux est déjà entièrement timestampé).
		"-fflags", "+genpts",
		// Sans framerate d'entrée forcé, genpts peut dériver sur une source
		// mal timestampée (framerate apparent ~90 000 fps vu par x264).
		"-r", "30",
		// Sonder vite : moins d'échantillons analysés = premier frame plus tôt.
		"-analyzeduration", "1000000",
		"-probesize", "1000000",
		// L'entrée est en RTSP loopback : c'est le seul moyen de lire le H.265
		// du DJI, MediaMTX ne le forwardant pas aux lecteurs RTMP classiques.
		"-rtsp_transport", "tcp",
		"-i", input,
		"-c:v", "libx264",
		"-preset", "veryfast",
		"-tune", "zerolatency",
		"-profile:v", "high",
		"-pix_fmt", "yuv420p",
		// Latence : keyframe toutes les secondes, aucun frame B, pas de
		// lookahead (zerolatency s'en charge déjà, on le rend explicite).
		"-g", "30",
		"-keyint_min", "30",
		"-sc_threshold", "0",
		"-bf", "0",
		"-x264-params", "rc-lookahead=0:sync-lookahead=0",
		"-b:v", "6M",
		"-maxrate", "6M",
		"-bufsize", "6M",
		"-fps_mode", "cfr",
		"-c:a", "aac",
		"-b:a", "128k",
		"-ar", "44100",
		// async=1 rattrape les trous/chevauchements de l'audio source en
		// insérant du silence ou en supprimant des échantillons. Sans
		// first_pts=0 : avec l'horloge murale, forcer l'audio à 0 pendant que
		// la vidéo part sur l'horloge système crée un décalage A/V énorme.
		"-af", "aresample=async=1",
		// Écriture immédiate des paquets : sans ça, le muxer FLV peut garder
		// une vidéo en tampon, ce qui ajoute de la latence côté OBS.
		"-flush_packets", "1",
		"-f", "flv",
		output,
	}
}

// RtmpURL builds the URL handed to the camera or pasted into OBS.
func RtmpURL(host string, port int, path string) string {
	return "rtmp://" + host + ":" + strconv.Itoa(port) + "/" + strings.TrimPrefix(path, "/")
}

// RtspURL is the loopback RTSP URL used by the transcoder.
func RtspURL(host string, port int, path string) string {
	return "rtsp://" + host + ":" + strconv.Itoa(port) + "/" + strings.TrimPrefix(path, "/")
}

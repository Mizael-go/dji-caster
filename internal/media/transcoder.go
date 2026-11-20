package media

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Transcoder builds the WebRTC copy of the incoming stream.
//
// WebRTC only carries H.264 video and Opus/G722/G711 audio. The camera (and
// the Android relay) push AAC audio, which a browser silently drops, and may
// push H.265. A single ffmpeg process therefore reads the source once and
// re-encodes both tracks directly (no intermediate AAC copy), which keeps the
// audio quality intact and adds only one frame of latency.
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
	t.logf("Encodeur WebRTC démarré (H.264 + Opus)")
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

		go watchProgress(stdout, t.logf)
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

// slowSpeed is the encoding speed under which ffmpeg falls behind real time.
// slowReports consecutive reports (stats_period = 2 s) trigger a warning.
const (
	slowSpeed   = 0.95
	slowReports = 3
)

// watchProgress reads ffmpeg's -progress output and logs only state changes:
// a warning when transcoding stays slower than real time (latency then grows
// without bound), and a notice once it catches up again.
func watchProgress(r io.Reader, logf func(string)) {
	sc := bufio.NewScanner(r)
	slow := 0
	warned := false
	for sc.Scan() {
		key, val, ok := strings.Cut(strings.TrimSpace(sc.Text()), "=")
		if !ok || key != "speed" {
			continue
		}
		speed, err := strconv.ParseFloat(strings.TrimSuffix(strings.TrimSpace(val), "x"), 64)
		if err != nil {
			continue // "N/A" au démarrage
		}
		if speed < slowSpeed {
			slow++
			if slow >= slowReports && !warned {
				warned = true
				logf(fmt.Sprintf("Transcodage plus lent que le temps réel (%.2fx) : la latence va augmenter. Baissez la résolution ou le débit de la caméra.", speed))
			}
			continue
		}
		slow = 0
		if warned {
			warned = false
			logf(fmt.Sprintf("Transcodage revenu au temps réel (%.2fx)", speed))
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
		// Progression machine sur stdout (lignes clé=valeur) : surveillée par
		// watchProgress pour signaler un transcodage plus lent que le temps
		// réel. -nostats évite les lignes \r illisibles sur stderr.
		"-nostats",
		"-progress", "pipe:1",
		"-stats_period", "2",
		"-flags", "low_delay",
		// Les timestamps du flux source (caméra DJI ou relais Android) sont
		// cassés : DTS égaux/répétés et audio démarrant avant la vidéo. Sans
		// restampage, le lecteur gèle sur les DTS non monotones. L'horloge
		// murale donne des timestamps monotones et recale l'A/V à l'arrivée des
		// paquets, quel que soit le décalage initial de la source.
		"-use_wallclock_as_timestamps", "1",
		// +genpts complète les pts manquants ; avec l'horloge murale il ne
		// dérive plus (le flux est déjà entièrement timestampé).
		//
		// Pas de -r en entrée : il remplace les timestamps (y compris l'horloge
		// murale) par un compteur d'images à 1/30 s. Après une coupure Wi-Fi,
		// la rafale d'images rattrapée était alors rejouée en temps réel et le
		// retard s'ajoutait définitivement (latence croissante, ~4 s observées).
		// La cadence est imposée en sortie (-r + -fps_mode cfr).
		"-fflags", "+genpts",
		// Sonder vite : moins d'échantillons analysés = premier frame plus tôt.
		"-analyzeduration", "1000000",
		"-probesize", "1000000",
		// L'entrée est en RTSP loopback : c'est le seul moyen de lire le H.265
		// du DJI, MediaMTX ne le forwardant pas aux lecteurs RTMP classiques.
		"-rtsp_transport", "tcp",
		"-i", input,
		"-c:v", "libx264",
		// ultrafast : le décodage HEVC logiciel + x264 1080p doit rester
		// au-dessus du temps réel, sinon le retard s'accumule sans fin.
		"-preset", "ultrafast",
		"-tune", "zerolatency",
		"-profile:v", "high",
		"-pix_fmt", "yuv420p",
		// Latence : keyframe toutes les secondes, aucun frame B, pas de
		// lookahead (zerolatency s'en charge déjà, on le rend explicite).
		"-g", "30",
		"-keyint_min", "30",
		"-sc_threshold", "0",
		"-bf", "0",
		// repeat-headers : SPS/PPS dans le flux, indispensables à un lecteur
		// WebRTC qui rejoint en cours de diffusion (sinon il n'a pas la
		// séquence de décodage et reste noir).
		"-x264-params", "rc-lookahead=0:sync-lookahead=0:repeat-headers=1:keyint=30:min-keyint=30:scenecut=0",
		// Keyframe forcée toutes les secondes même si la cadence d'entrée
		// dérive : un lecteur WebRTC qui rejoint attend une IDR.
		"-force_key_frames", "expr:gte(t,n_forced*1)",
		"-b:v", "6M",
		"-maxrate", "6M",
		"-bufsize", "6M",
		// Cadence fixe en sortie, calculée sur l'horloge murale : une rafale
		// après coupure est écrêtée (images en trop jetées) et un trou est
		// comblé par duplication, ce qui garde la latence constante. -r en
		// sortie fixe aussi le framerate vu par x264 (sinon ~90 000 fps).
		"-r", "30",
		"-fps_mode", "cfr",
		// Audio : Opus est le seul codec transmis par WebRTC. On l'encode
		// directement depuis la source (et non depuis une copie AAC
		// intermédiaire) : l'AAC→Opus en deux passes était la cause du son
		// dégradé. 48 kHz stéréo, 160 kb/s, profil « audio » (meilleure
		// qualité que « lowdelay » à ce débit), trames de 20 ms (défaut WebRTC).
		"-c:a", "libopus",
		"-b:a", "160k",
		"-ar", "48000",
		"-ac", "2",
		"-application", "audio",
		"-frame_duration", "20",
		// async=1 rattrape les trous/chevauchements de l'audio source en
		// insérant du silence ou en supprimant des échantillons. Sans
		// first_pts=0 : avec l'horloge murale, forcer l'audio à 0 pendant que
		// la vidéo part sur l'horloge système crée un décalage A/V énorme.
		"-af", "aresample=async=1",
		// Écriture immédiate des paquets : le muxer FLV ne doit rien tamponner,
		// sinon la latence WebRTC augmente.
		"-flush_packets", "1",
		"-muxdelay", "0",
		"-muxpreload", "0",
		"-f", "flv",
		output,
	}
}

// WebrtcURL is MediaMTX's built-in player page: it negotiates WHEP in the
// browser and needs no extra client.
func WebrtcURL(host string, port int, path string) string {
	return "http://" + host + ":" + strconv.Itoa(port) + "/" + strings.TrimPrefix(path, "/")
}

// RtmpURL builds the ingest URL handed to the camera or the Android relay.
func RtmpURL(host string, port int, path string) string {
	return "rtmp://" + host + ":" + strconv.Itoa(port) + "/" + strings.TrimPrefix(path, "/")
}

// RtspURL is the loopback RTSP URL used by the transcoder.
func RtspURL(host string, port int, path string) string {
	return "rtsp://" + host + ":" + strconv.Itoa(port) + "/" + strings.TrimPrefix(path, "/")
}

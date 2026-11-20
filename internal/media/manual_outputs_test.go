//go:build manual

package media

import (
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Vérification manuelle de la sortie WebRTC : démarre le vrai MediaMTX avec la
// configuration générée par l'app, pousse un flux de test H.264 + AAC en RTMP,
// puis encode la copie WebRTC (H.264 + Opus) et interroge MediaMTX.
//
//	go test -tags manual -run TestWebrtcOutput -v ./internal/media/
func TestWebrtcOutput(t *testing.T) {
	home, _ := os.UserHomeDir()
	bin := filepath.Join(home, ".config", "dji-caster", "bin", "mediamtx")
	if _, err := os.Stat(bin); err != nil {
		t.Skip("MediaMTX absent de ~/.config/dji-caster/bin")
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg absent")
	}

	srv := NewServer(ServerConfig{
		Binary:   bin,
		Dir:      t.TempDir(),
		RtmpPort: 11935,
		ApiPort:  19997,
		Logf:     func(s string) { t.Log("mediamtx: " + s) },
	})
	if err := srv.Start(); err != nil {
		t.Fatalf("démarrage : %v", err)
	}
	defer srv.Stop()

	push := exec.Command("ffmpeg",
		"-hide_banner", "-loglevel", "error", "-re",
		"-f", "lavfi", "-i", "testsrc=size=640x360:rate=30",
		"-f", "lavfi", "-i", "sine=frequency=440", "-t", "40",
		"-c:v", "libx264", "-preset", "ultrafast", "-tune", "zerolatency",
		"-g", "30", "-pix_fmt", "yuv420p",
		"-c:a", "aac",
		"-f", "flv", "rtmp://127.0.0.1:11935/live/dji")
	var pushErr strings.Builder
	push.Stderr = &pushErr
	if err := push.Start(); err != nil {
		t.Fatalf("ffmpeg : %v", err)
	}
	defer func() {
		_ = push.Process.Kill()
		_ = push.Wait()
		if s := pushErr.String(); s != "" {
			t.Logf("ffmpeg stderr : %s", s)
		}
	}()

	// La source doit être prête avant de lancer l'encodeur.
	sourceDeadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(sourceDeadline) {
		body, code := get(t, "http://127.0.0.1:19997/v3/paths/get/live/dji")
		if code == http.StatusOK && strings.Contains(body, `"ready":true`) {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Encodeur réel : les arguments utilisés par l'app doivent produire un
	// flux H.264 + Opus lisible en WebRTC.
	transcoder := NewTranscoder("ffmpeg", func(s string) { t.Log("ffmpeg: " + s) })
	if err := transcoder.Start(
		"rtsp://127.0.0.1:8554/live/dji",
		"rtmp://127.0.0.1:11935/live/dji-webrtc",
	); err != nil {
		t.Fatalf("encodage : %v", err)
	}
	defer transcoder.Stop()

	transDeadline := time.Now().Add(25 * time.Second)
	ready := false
	for time.Now().Before(transDeadline) {
		body, code := get(t, "http://127.0.0.1:19997/v3/paths/get/live/dji-webrtc")
		if code == http.StatusOK && strings.Contains(body, `"ready":true`) {
			if !strings.Contains(body, "H264") {
				t.Fatalf("sortie WebRTC sans H.264 : %s", firstLine(body))
			}
			if !strings.Contains(body, "Opus") {
				t.Fatalf("sortie WebRTC sans Opus : %s", firstLine(body))
			}
			t.Logf("Encodeur : live/dji-webrtc prêt avec H.264 + Opus")
			ready = true
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if !ready {
		t.Fatalf("live/dji-webrtc jamais prêt (err=%q)", transcoder.Err())
	}

	page, code := get(t, "http://127.0.0.1:8889/live/dji-webrtc")
	if code != http.StatusOK {
		t.Fatalf("page WebRTC : HTTP %d", code)
	}
	if !strings.Contains(strings.ToLower(page), "whep") {
		t.Errorf("page WebRTC inattendue : %q", firstLine(page))
	}
	t.Logf("WebRTC : page de lecture servie (%d octets)", len(page))

	// HLS ne doit plus écouter.
	if _, code := get(t, "http://127.0.0.1:8888/live/dji/index.m3u8"); code != 0 {
		t.Errorf("HLS répond encore sur :8888 (HTTP %d)", code)
	}
}

func get(t *testing.T, url string) (string, int) {
	t.Helper()
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", 0
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	return string(body), resp.StatusCode
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	if len(s) > 160 {
		s = s[:160]
	}
	return s
}

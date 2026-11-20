package media

import (
	"slices"
	"strings"
	"testing"
)

func TestWatchProgressWarnsOnlyOnSustainedSlowness(t *testing.T) {
	in := strings.Join([]string{
		"speed=N/A", "progress=continue",
		"speed=0.80x", "speed=0.90x", // 2 rapports lents : pas encore d'alerte
		"speed=1.01x",
		"speed=0.70x", "speed=0.70x", "speed=0.70x", "speed=0.60x",
		"speed=1.00x",
	}, "\n")
	var logs []string
	watchProgress(strings.NewReader(in), func(s string) { logs = append(logs, s) })
	if len(logs) != 2 {
		t.Fatalf("attendu 2 messages (alerte + rétablissement), obtenu %q", logs)
	}
	if !strings.Contains(logs[0], "0.70x") || !strings.Contains(logs[1], "1.00x") {
		t.Fatalf("messages inattendus : %q", logs)
	}
}

func TestTranscodeArgsNoInputFramerate(t *testing.T) {
	args := transcodeArgs("rtsp://in", "rtmp://out")
	i := slices.Index(args, "-i")
	if slices.Contains(args[:i], "-r") {
		t.Fatal("-r ne doit pas être une option d'entrée : il écrase l'horloge murale")
	}
	if !slices.Contains(args[i:], "-r") {
		t.Fatal("-r attendu en sortie")
	}
}

func TestTranscodeArgsEncodeOpusForWebrtc(t *testing.T) {
	args := strings.Join(transcodeArgs("rtsp://in", "rtmp://out"), " ")
	// Une seule passe : la vidéo et l'audio sont réencodés ensemble, l'audio
	// Opus part directement de la source (pas de copie AAC intermédiaire).
	for _, want := range []string{
		"-c:v libx264",
		"-c:a libopus",
		"-b:a 160k",
		"-ar 48000",
		"-application audio",
		"-frame_duration 20",
	} {
		if !strings.Contains(args, want) {
			t.Errorf("argument manquant pour WebRTC : %q dans %q", want, args)
		}
	}
}

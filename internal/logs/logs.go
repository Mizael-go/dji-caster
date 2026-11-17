package logs

import (
	"strings"
	"sync"
	"time"
)

type Line struct {
	Scope string `json:"scope"`
	Level string `json:"level"`
	Text  string `json:"text"`
	Time  int64  `json:"time"`
}

type Buffer struct {
	mu    sync.Mutex
	lines []Line
	max   int
}

func New(max int) *Buffer {
	if max <= 0 {
		max = 300
	}
	return &Buffer{max: max}
}

func (b *Buffer) Add(scope, text string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.lines = append(b.lines, Line{
		Scope: scope,
		Level: classify(text),
		Text:  text,
		Time:  time.Now().UnixMilli(),
	})
	if len(b.lines) > b.max {
		b.lines = b.lines[len(b.lines)-b.max:]
	}
}

func (b *Buffer) Snapshot() []Line {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]Line, len(b.lines))
	copy(out, b.lines)
	return out
}

func (b *Buffer) Clear() {
	b.mu.Lock()
	b.lines = nil
	b.mu.Unlock()
}

var warnMarkers = []string{
	"error", "erreur", "failed", "fail", "warn", "impossible", "refus", "denied",
}

// classify tags a line as a warning so the console can highlight it.
func classify(text string) string {
	low := strings.ToLower(text)
	for _, m := range warnMarkers {
		if strings.Contains(low, m) {
			return "warn"
		}
	}
	return "info"
}

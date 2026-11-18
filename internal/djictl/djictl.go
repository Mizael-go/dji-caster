// Package djictl drives the upstream djictl CLI that talks to a DJI Osmo over
// BLE and orders it to join the local Wi-Fi then push RTMP to us.
package djictl

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Runner struct {
	mu      sync.Mutex
	cancel  context.CancelFunc
	running bool
	lastErr string
}

func NewRunner() *Runner { return &Runner{} }

func (r *Runner) Running() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.running
}

func (r *Runner) Err() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.lastErr
}

// Args mirrors:
//
//	djictl --filter-device-addr <MAC> ble connect-wifi-and-start-streaming \
//	  --wifi-ssid <SSID> --wifi-psk <PSK> --rtmp-url <URL> \
//	  --resolution <480p|720p|1080p> --bitrate-kbps <n> --fps <25|30>
//
// resolution, bitrateKbps et fps sont appliqués à distance à l'encodeur de la
// caméra : rien à régler dans le menu du DJI.
func Args(addr, ssid, psk, rtmpURL, resolution string, bitrateKbps, fps int) []string {
	return []string{
		"--log-level", "debug",
		"--filter-device-addr", addr,
		"ble", "connect-wifi-and-start-streaming",
		"--wifi-ssid", ssid,
		"--wifi-psk", psk,
		"--rtmp-url", rtmpURL,
		"--resolution", resolution,
		"--bitrate-kbps", strconv.Itoa(bitrateKbps),
		"--fps", strconv.Itoa(fps),
	}
}

// CommandLine is shown in the UI so the user can reproduce the call manually.
func CommandLine(bin, addr, ssid, psk, rtmpURL, resolution string, bitrateKbps, fps int) string {
	parts := []string{bin}
	for _, a := range Args(addr, ssid, psk, rtmpURL, resolution, bitrateKbps, fps) {
		if strings.ContainsAny(a, " \"'") {
			parts = append(parts, "'"+a+"'")
		} else {
			parts = append(parts, a)
		}
	}
	return strings.Join(parts, " ")
}

// Run executes djictl and streams its output line by line. It blocks until the
// command exits or ctx is cancelled.
func (r *Runner) Run(ctx context.Context, bin, addr, ssid, psk, rtmpURL, resolution string, bitrateKbps, fps int, logf func(string)) error {
	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return fmt.Errorf("une connexion DJI est déjà en cours")
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	r.cancel = cancel
	r.running = true
	r.lastErr = ""
	r.mu.Unlock()

	defer func() {
		cancel()
		r.mu.Lock()
		r.running = false
		r.cancel = nil
		r.mu.Unlock()
	}()

	logf("$ " + CommandLine(bin, addr, ssid, psk, rtmpURL, resolution, bitrateKbps, fps))
	cmd := exec.CommandContext(ctx, bin, Args(addr, ssid, psk, rtmpURL, resolution, bitrateKbps, fps)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		r.setErr(err)
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		r.setErr(err)
		return err
	}
	if err := cmd.Start(); err != nil {
		r.setErr(err)
		return fmt.Errorf("lancement de djictl : %w", err)
	}

	var wg sync.WaitGroup
	scan := func(rd interface{ Read([]byte) (int, error) }) {
		defer wg.Done()
		sc := bufio.NewScanner(rd)
		sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for sc.Scan() {
			logf(strings.TrimRight(sc.Text(), "\r"))
		}
	}
	wg.Add(2)
	go scan(stdout)
	go scan(stderr)

	err = cmd.Wait()
	wg.Wait()

	if ctx.Err() == context.DeadlineExceeded {
		r.setErr(fmt.Errorf("djictl expiré après 3 minutes"))
		return fmt.Errorf("djictl expiré après 3 minutes")
	}
	if err != nil {
		if ctx.Err() == context.Canceled {
			return fmt.Errorf("connexion annulée")
		}
		r.setErr(err)
		return fmt.Errorf("djictl a échoué : %w", err)
	}
	return nil
}

// Stop cancels a running invocation.
func (r *Runner) Stop() {
	r.mu.Lock()
	cancel := r.cancel
	r.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (r *Runner) setErr(err error) {
	r.mu.Lock()
	r.lastErr = err.Error()
	r.mu.Unlock()
}

// Hint turns the usual djictl failures into actionable French messages.
func Hint(logs []string) string {
	joined := strings.ToLower(strings.Join(logs, "\n"))
	switch {
	case strings.Contains(joined, "permission denied"), strings.Contains(joined, "operation not permitted"):
		return "djictl n'a pas accès au Bluetooth : accordez les droits (groupe bluetooth / policykit) ou lancez-le avec les privilèges adéquats."
	case strings.Contains(joined, "no such device"), strings.Contains(joined, "not found"):
		return "Appareil introuvable : rapprochez-vous du DJI, réveillez-le (écran allumé) puis relancez le scan."
	case strings.Contains(joined, "connection refused"), strings.Contains(joined, "timeout"):
		return "La caméra n'a pas répondu. Vérifiez qu'elle est appairée et que le Wi-Fi ciblé est bien en 2,4 GHz."
	}
	return ""
}

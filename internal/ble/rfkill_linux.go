//go:build linux

package ble

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// RfkillState reflects the kernel radio switch state for Bluetooth.
type RfkillState struct {
	Present bool
	Soft    bool
	Hard    bool
}

func ReadRfkill() RfkillState {
	st := RfkillState{}
	paths, err := filepath.Glob("/sys/class/rfkill/rfkill*")
	if err != nil {
		return st
	}
	for _, p := range paths {
		raw, err := os.ReadFile(filepath.Join(p, "type"))
		if err != nil || strings.TrimSpace(string(raw)) != "bluetooth" {
			continue
		}
		st.Present = true
		if readSwitch(filepath.Join(p, "soft")) {
			st.Soft = true
		}
		if readSwitch(filepath.Join(p, "hard")) {
			st.Hard = true
		}
	}
	return st
}

func readSwitch(path string) bool {
	raw, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	v, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	return err == nil && v == 1
}

// Unblock clears the rfkill soft block. Writing to the kernel switch needs
// root (or a polkit rule); the app never asks for a password itself and simply
// reports the exact command to run once when it is refused.
func Unblock() error {
	st := ReadRfkill()
	if !st.Present {
		return nil
	}
	if st.Hard {
		return fmt.Errorf("bluetooth coupé par un interrupteur physique : rallumez-le")
	}
	if !st.Soft {
		return nil
	}
	bin, err := exec.LookPath("rfkill")
	if err != nil {
		return fmt.Errorf("outil rfkill introuvable (paquet util-linux) — exécutez « sudo rfkill unblock bluetooth »")
	}
	out, err := exec.Command(bin, "unblock", "bluetooth").CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("déblocage refusé (%s) — exécutez une fois « sudo rfkill unblock bluetooth »", msg)
	}
	// Le noyau applique le changement de façon asynchrone.
	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)
		if s := ReadRfkill(); !s.Soft && !s.Hard {
			return nil
		}
	}
	return fmt.Errorf("rfkill signale toujours un blocage après « rfkill unblock bluetooth »")
}

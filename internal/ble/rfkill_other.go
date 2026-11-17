//go:build !linux

package ble

// RfkillState is Linux-only; other platforms report "no radio switch".
type RfkillState struct {
	Present bool
	Soft    bool
	Hard    bool
}

func ReadRfkill() RfkillState { return RfkillState{} }

func Unblock() error { return nil }

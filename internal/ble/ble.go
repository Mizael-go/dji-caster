// Package ble exposes a minimal BlueZ client over D-Bus: adapter discovery,
// continuous device discovery and a snapshot of every device seen so far.
// Talking to org.bluez directly avoids parsing bluetoothctl's ANSI-coloured
// interactive output.
package ble

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/godbus/dbus/v5"
)

const (
	bluezService = "org.bluez"
	adapterIface = "org.bluez.Adapter1"
	deviceIface  = "org.bluez.Device1"
	omIface      = "org.freedesktop.DBus.ObjectManager"
	propsIface   = "org.freedesktop.DBus.Properties"
)

type Device struct {
	Address   string `json:"address"`
	Name      string `json:"name"`
	RSSI      int    `json:"rssi"`
	Connected bool   `json:"connected"`
	Paired    bool   `json:"paired"`
	IsDJI     bool   `json:"isDJI"`
	LastSeen  int64  `json:"lastSeen"`
}

type Scanner struct {
	mu       sync.Mutex
	conn     *dbus.Conn
	adapter  dbus.ObjectPath
	devices  map[string]*Device
	scanning bool
	err      string
	onChange func()
	stop     chan struct{}
}

func NewScanner(onChange func()) *Scanner {
	return &Scanner{devices: map[string]*Device{}, onChange: onChange}
}

func (s *Scanner) DeviceList() []Device {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Device, 0, len(s.devices))
	for _, d := range s.devices {
		out = append(out, *d)
	}
	// DJI hardware first, then by signal strength, then by name.
	sort.Slice(out, func(i, j int) bool {
		if out[i].IsDJI != out[j].IsDJI {
			return out[i].IsDJI
		}
		if out[i].RSSI != out[j].RSSI {
			return out[i].RSSI > out[j].RSSI
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func (s *Scanner) Scanning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.scanning
}

func (s *Scanner) Err() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.err
}

// Start powers the default adapter, enables discovery and subscribes to BlueZ
// signals. It is safe to call again while already scanning.
func (s *Scanner) Start() error {
	s.mu.Lock()
	if s.scanning {
		s.mu.Unlock()
		return nil
	}
	s.mu.Unlock()

	// Un blocage rfkill est la cause n°1 d'un scan vide : on tente de le lever
	// avant même d'ouvrir le bus.
	if st := ReadRfkill(); st.Present && (st.Soft || st.Hard) {
		if err := Unblock(); err != nil {
			s.mu.Lock()
			s.err = err.Error()
			s.mu.Unlock()
			return err
		}
	}

	conn, err := dbus.SystemBusPrivate()
	if err != nil {
		return fmt.Errorf("bus système D-Bus inaccessible : %w", err)
	}
	if err := conn.Auth(nil); err != nil {
		_ = conn.Close()
		return fmt.Errorf("authentification D-Bus refusée : %w", err)
	}
	if err := conn.Hello(); err != nil {
		_ = conn.Close()
		return fmt.Errorf("bus système D-Bus indisponible : %w", err)
	}

	adapter, err := findAdapter(conn)
	if err != nil {
		return err
	}
	// L'allumage peut échouer (rfkill, polkit) : la découverte reste possible
	// si l'adaptateur est déjà alimenté, sinon l'erreur est remontée plus bas.
	// Adapter1.SetPowered a été retiré des BlueZ récents : la propriétéPowered
	// s'écrit via org.freedesktop.DBus.Properties.Set.
	powerErr := conn.Object(bluezService, adapter).Call(
		propsIface+".Set", 0, adapterIface, "Powered", dbus.MakeVariant(true),
	).Err

	signals := make(chan *dbus.Signal, 128)
	conn.Signal(signals)
	for _, m := range []dbus.MatchOption{
		dbus.WithMatchObjectPath("/"),
		dbus.WithMatchInterface(omIface),
	} {
		if err := conn.AddMatchSignal(m); err != nil {
			return fmt.Errorf("abonnement D-Bus impossible : %w", err)
		}
	}
	// Les PropertiesChanged des appareils sont émis sur leur propre chemin
	// (/org/bluez/hci0/dev_XX) : on filtre sur l'interface de l'argument 0.
	if err := conn.AddMatchSignal(
		dbus.WithMatchInterface(propsIface),
		dbus.WithMatchArg(0, deviceIface),
	); err != nil {
		return fmt.Errorf("abonnement D-Bus impossible : %w", err)
	}

	s.mu.Lock()
	s.conn = conn
	s.adapter = adapter
	s.scanning = true
	s.err = ""
	s.stop = make(chan struct{})
	stop := s.stop
	s.mu.Unlock()

	s.loadManagedObjects(conn)

	if err := conn.Object(bluezService, adapter).Call(adapterIface+".StartDiscovery", 0).Err; err != nil {
		reason := friendlyErr(err)
		if powerErr != nil {
			reason = "adaptateur Bluetooth éteint ou bloqué (rfkill / bluetooth.service) — " + friendlyErr(powerErr)
		}
		s.mu.Lock()
		s.scanning = false
		s.err = reason
		s.mu.Unlock()
		conn.RemoveSignal(signals)
		_ = conn.Close()
		return fmt.Errorf("%s", reason)
	}

	go s.loop(conn, signals, stop)
	s.notify()
	return nil
}

// prune drops devices that have not been seen for a while so the map cannot
// grow without bound on long sessions.
func (s *Scanner) Prune(maxAge time.Duration) {
	cutoff := time.Now().Add(-maxAge).UnixMilli()
	s.mu.Lock()
	removed := 0
	for addr, d := range s.devices {
		if d.LastSeen < cutoff {
			delete(s.devices, addr)
			removed++
		}
	}
	s.mu.Unlock()
	if removed > 0 {
		s.notify()
	}
}

func (s *Scanner) Stop() error {
	s.mu.Lock()
	if !s.scanning {
		s.mu.Unlock()
		return nil
	}
	conn, adapter, stop := s.conn, s.adapter, s.stop
	s.scanning = false
	s.mu.Unlock()

	close(stop)
	if conn != nil {
		_ = conn.Object(bluezService, adapter).Call(adapterIface+".StopDiscovery", 0).Err
		_ = conn.Close()
	}
	s.notify()
	return nil
}

func (s *Scanner) loop(conn *dbus.Conn, signals chan *dbus.Signal, stop chan struct{}) {
	for {
		select {
		case <-stop:
			return
		case sig, ok := <-signals:
			if !ok {
				return
			}
			if s.handleSignal(sig) {
				s.notify()
			}
		}
	}
}

func (s *Scanner) handleSignal(sig *dbus.Signal) bool {
	switch sig.Name {
	case omIface + ".InterfacesAdded":
		if len(sig.Body) < 2 {
			return false
		}
		path, ok := sig.Body[0].(dbus.ObjectPath)
		if !ok {
			return false
		}
		ifaces, ok := sig.Body[1].(map[string]map[string]dbus.Variant)
		if !ok {
			return false
		}
		props, ok := ifaces[deviceIface]
		if !ok {
			return false
		}
		return s.upsert(string(path), props)
	case propsIface + ".PropertiesChanged":
		if len(sig.Body) < 3 {
			return false
		}
		iface, _ := sig.Body[0].(string)
		if iface != deviceIface {
			return false
		}
		changed, ok := sig.Body[1].(map[string]dbus.Variant)
		if !ok {
			return false
		}
		return s.upsert(string(sig.Path), changed)
	}
	return false
}

func (s *Scanner) upsert(path string, props map[string]dbus.Variant) bool {
	addr := variantString(props, "Address")
	if addr == "" {
		addr = strings.TrimPrefix(path, "/org/bluez/hci0/dev_")
		addr = strings.ReplaceAll(addr, "_", ":")
	}
	if addr == "" {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.devices[addr]
	if !ok {
		d = &Device{Address: addr}
		s.devices[addr] = d
	}

	if v := variantString(props, "Alias"); v != "" {
		d.Name = v
	} else if v := variantString(props, "Name"); v != "" {
		d.Name = v
	}
	if v, ok := props["RSSI"]; ok {
		if rssi, ok := v.Value().(int16); ok {
			d.RSSI = int(rssi)
		}
	}
	if v, ok := props["Connected"]; ok {
		if b, ok := v.Value().(bool); ok {
			d.Connected = b
		}
	}
	if v, ok := props["Paired"]; ok {
		if b, ok := v.Value().(bool); ok {
			d.Paired = b
		}
	}
	d.IsDJI = looksLikeDJI(d.Name)
	d.LastSeen = time.Now().UnixMilli()
	return true
}

func (s *Scanner) loadManagedObjects(conn *dbus.Conn) {
	var objects map[dbus.ObjectPath]map[string]map[string]dbus.Variant
	if err := conn.Object(bluezService, "/").Call(omIface+".GetManagedObjects", 0).Store(&objects); err != nil {
		return
	}
	for path, ifaces := range objects {
		if props, ok := ifaces[deviceIface]; ok {
			s.upsert(string(path), props)
		}
	}
}

func (s *Scanner) notify() {
	if s.onChange != nil {
		s.onChange()
	}
}

func findAdapter(conn *dbus.Conn) (dbus.ObjectPath, error) {
	var objects map[dbus.ObjectPath]map[string]map[string]dbus.Variant
	err := conn.Object(bluezService, "/").Call(omIface+".GetManagedObjects", 0).Store(&objects)
	if err != nil {
		return "", fmt.Errorf("BlueZ injoignable : %s", friendlyErr(err))
	}
	var found []dbus.ObjectPath
	for path, ifaces := range objects {
		if _, ok := ifaces[adapterIface]; ok {
			found = append(found, path)
		}
	}
	if len(found) == 0 {
		return "", fmt.Errorf("aucun adaptateur Bluetooth détecté (bluetoothd est-il démarré ?)")
	}
	sort.Slice(found, func(i, j int) bool { return found[i] < found[j] })
	return found[0], nil
}

func variantString(props map[string]dbus.Variant, key string) string {
	if v, ok := props[key]; ok {
		if s, ok := v.Value().(string); ok {
			return s
		}
	}
	return ""
}

func friendlyErr(err error) string {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "NotReady"), strings.Contains(msg, "Resource Not Ready"):
		return "adaptateur Bluetooth éteint (rfkill / bluetooth.service)"
	case strings.Contains(msg, "Blocked"), strings.Contains(msg, "rfkill"):
		return "adaptateur Bluetooth bloqué par rfkill (rfkill unblock bluetooth)"
	case strings.Contains(msg, "PermissionDenied"), strings.Contains(msg, "not authorized"):
		return "accès Bluetooth refusé par polkit (ajoutez l'utilisateur au groupe bluetooth)"
	case strings.Contains(msg, "org.bluez.Error.InProgress"):
		return "découverte déjà en cours"
	}
	return msg
}

var djiMarkers = []string{"dji", "osmo", "mavic", "avata", "goggles", "ronin", "pocket 2", "pocket 3", "action 2", "action 3", "action 4", "action 5", "neo "}

func looksLikeDJI(name string) bool {
	n := strings.ToLower(strings.TrimSpace(name))
	if n == "" {
		return false
	}
	for _, m := range djiMarkers {
		if strings.Contains(n, m) {
			return true
		}
	}
	return false
}

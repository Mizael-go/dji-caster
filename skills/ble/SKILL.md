---
name: ble
description: >-
  Découverte Bluetooth Low Energy sous Linux pour DJI Caster : BlueZ via D-Bus,
  appairage, pièges RSSI/scan. À lire avant de modifier internal/ble ou le flux
  de connexion DJI. Réf. : https://github.com/dotintent/awesome-ble
---

# BLE — BlueZ via D-Bus

## Choix d'implémentation

`bluetoothctl` en mode non interactif écrit des séquences ANSI et des prompts,
ce qui rend le parsing fragile. `internal/ble` parle donc directement à
`org.bluez` sur le bus système avec `github.com/godbus/dbus/v5` :

1. `GetManagedObjects` sur `/` → trouver le premier adaptateur `org.bluez.Adapter1`.
2. `SetPowered(true)` (échec toléré si polkit refuse).
3. `StartDiscovery`, puis écoute de :
   - `org.freedesktop.DBus.ObjectManager.InterfacesAdded` (chemin `/`)
   - `org.freedesktop.DBus.Properties.PropertiesChanged` filtré sur
     l'argument 0 = `org.bluez.Device1` (les signaux partent du chemin de
     l'appareil, pas de `/` — un match sur `ObjectPath("/")` ne reçoit rien).
4. Lecture des propriétés `Address`, `Alias`, `Name`, `RSSI` (`int16`),
   `Connected`, `Paired`.

## Pièges

- `dbus.SystemBus()` renvoie une connexion partagée : la fermer casse le bus
  pour tout le processus. Utiliser `SystemBusPrivate()` + `Auth(nil)` + `Hello()`.
- RSSI = 0 signifie « inconnu » (appareil vu avant le scan), pas « très loin ».
- `StopDiscovery` avant de lancer djictl : certains contrôleurs refusent une
  connexion pendant la découverte.
- Adapter bloqué par rfkill → `StartDiscovery` renvoie `org.bluez.Error.NotReady` ;
  le message est traduit par `friendlyErr`.
- Un appareil DJI est reconnu par son nom (`dji`, `osmo`, `mavic`, `avata`,
  `goggles`, `action 3/4/5`, `pocket 3`…), pas par son adresse MAC.

## Connexion DJI

La connexion BLE n'est pas faite par l'app : `djictl` s'en charge avec
`--filter-device-addr <MAC> ble connect-wifi-and-start-streaming …`. L'app
transmet l'adresse choisie dans la liste et suit la sortie du processus.

---
id: "DJC-TASK-012"
type: "TASK"
status: "review"
priority: "high"
labels: ["desktop", "dji", "ble", "encodeur"]
components: ["internal/config", "internal/djictl", "app", "frontend"]
story_points: 5
epic: null
sprint: null
parent: null
children: []
depends_on: []
blocks: []
related_to: ["DJC-TASK-010", "DJC-BUG-008"]
tokens_estimated: 18k
tokens_used: 16k
created: "2026-11-18"
updated: "2026-11-18"
---
## Summary

Le forçage BLE n'envoyait que le SSID, le PSK et l'URL RTMP. La caméra
restait sur son encodeur par défaut (24 fps) et l'utilisateur devait
chercher les réglages vidéo dans le menu du DJI.

## Description

`djictl ble connect-wifi-and-start-streaming` accepte
`--resolution` (480p/720p/1080p), `--bitrate-kbps` et `--fps` (25/30) :
ces options sont appliquées à distance par Bluetooth en même temps que
l'URL RTMP. L'app doit les exposer et les persister.

## Fix

- `internal/config` : champs `streamResolution` (défaut `1080p`),
  `streamBitrate` (6000 kbps) et `streamFps` (30) ; `NormalizeStream()`
  borne les valeurs aux choix de djictl et est appelée au chargement.
- `internal/djictl` : `Args`/`CommandLine`/`Run` transportent les trois
  paramètres et construisent la ligne de commande affichée dans l'UI.
- `app.go` : `ConnectRequest` étendu (`resolution`, `bitrate`, `fps`) ;
  les valeurs sont mémorisées à chaque connexion et transmises à djictl ;
  les champs absents (ancien client) conservent les réglages enregistrés.
- Frontend : trois sélecteurs « Encodeur de la caméra » dans le dialogue
  de forçage (résolution, débit 4–20 Mb/s, 25/30 fps), pré-remplis avec
  les réglages persistés, avec la mention « Appliqués à distance par
  Bluetooth ».
- Les types Go (tags JSON) et `frontend/src/lib/types.ts` restent miroirs.

## Acceptance Criteria

- [x] 1080p / 6 Mb/s / 30 fps par défaut, y compris sur une config neuve.
- [x] Les valeurs choisies sont persistées et réutilisées au forçage suivant.
- [x] La ligne de commande djictl affichée contient les trois options.
- [x] `go build ./... && go vet ./...` et `npm run build` OK.
- [ ] Validation terrain : la caméra diffuse bien en 1080p30 après forçage.

## Notes

Vérification : `go build ./... && go vet ./...` OK ; `npm run build` OK
(1750 modules). `djictl ble connect-wifi-and-start-streaming --help`
confirme les valeurs acceptées (480p/720p/1080p, 25/30).

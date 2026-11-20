---
id: "DJC-TASK-015"
type: "TASK"
status: "done"
priority: "high"
labels: ["audio", "webrtc", "opus", "ffmpeg"]
components: ["backend", "frontend"]
story_points: 3
epic: null
sprint: null
parent: null
children: []
depends_on: ["DJC-TASK-014"]
blocks: []
related_to: ["DJC-TASK-014"]
tokens_estimated: 12000
tokens_used: 11000
created: "2026-11-19"
updated: "2026-11-19"
---
## Summary

Rendre le son de la caméra audible dans la page WebRTC, sans dégrader OBS ni HLS.

## Description

WebRTC ne transporte que l'Opus, le G722 et le G711. L'AAC du DJI (et l'AAC de la
copie ffmpeg) circule très bien en RTMP et en HLS mais est silencieusement ignoré
par un lecteur navigateur : la page WebRTC n'avait donc que l'image.

MediaMTX n'effectue aucune conversion de codec, d'où un relais ffmpeg dédié
(`internal/media/relay.go`) :

- il lit le chemin déjà lu par OBS (`live/dji`, ou `live/dji-h264` si la
  conversion est active) en RTSP loopback ;
- il **copie la vidéo** (`-c:v copy`, aucun réencodage) et réencode uniquement
  l'audio en Opus (`libopus`, 96 kb/s, 48 kHz, `application lowdelay`, trames de
  20 ms) ;
- il publie `live/dji-webrtc`, lu par la page WebRTC uniquement.

L'AAC reste donc intact sur `live/dji` et `live/dji-h264` : OBS et HLS (et
Safari, qui ne lit pas l'Opus en HLS) gardent leur compatibilité.

`-map 0:a:0?` laisse le relais fonctionner avec une caméra muette ; l'UI
distingue alors « pas de piste audio en entrée » d'un relais en erreur.

## Acceptance Criteria

- [x] Le chemin publié pour WebRTC contient une piste Opus à côté du H.264.
- [x] Le relais ne réencode pas la vidéo (copie de flux).
- [x] OBS et HLS continuent de recevoir l'AAC.
- [x] L'UI indique si le son est présent et remonte l'erreur du relais.
- [x] `go build ./...`, `go vet ./...`, `go test ./...` et `npm run build` passent.
- [x] Le binaire Wails est reconstruit.

## Notes

Vérifié par `internal/media/manual_outputs_test.go`
(`go test -tags manual -run TestLowLatencyOutputs`) : MediaMTX journalise
`[path live/dji-webrtc] … 2 tracks (H264, Opus)`, le chemin H.264 conserve
`MPEG-4 Audio`.

Coût : une seule piste audio réencodée, la vidéo est copiée ; l'ajout de latence
du relais (lecture RTSP + publication RTMP locales) est de l'ordre de quelques
dizaines de millisecondes.

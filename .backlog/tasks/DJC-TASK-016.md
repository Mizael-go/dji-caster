---
id: "DJC-TASK-016"
type: "TASK"
status: "done"
priority: "high"
labels: ["webrtc", "audio", "opus", "suppression", "ffmpeg"]
components: ["backend", "frontend", "packaging"]
story_points: 5
epic: null
sprint: null
parent: null
children: []
depends_on: ["DJC-TASK-014", "DJC-TASK-015"]
blocks: []
related_to: ["DJC-TASK-012", "DJC-TASK-013"]
tokens_estimated: 25000
tokens_used: 26000
created: "2026-11-20"
updated: "2026-11-20"
---
## Summary

Simplifier l'app autour d'une seule sortie WebRTC : retirer le Bluetooth et la
sortie RTMP/HLS, et encoder l'audio Opus directement avec la vidéo.

## Description

- **Bluetooth supprimé** : page « Appareils », scan BLE, rfkill, dialogue de
  forçage et `djictl` (ainsi que son téléchargement) disparaissent. L'app ne
  gère plus la connexion de la caméra : le relais Android (ou une caméra déjà
  sur le Wi-Fi) pousse sur l'entrée RTMP.
- **Sorties RTMP/HLS retirées** : MediaMTX n'expose plus que WebRTC
  (HTTP `:8889`, UDP `:8189`). Le RTMP reste uniquement le protocole d'entrée.
- **Audio corrigé** : le relais Opus séparé (`relay.go`) est supprimé. Un seul
  ffmpeg lit la source en RTSP loopback et réencode la vidéo H.264 **et**
  l'audio Opus en une passe, puis publie `live/dji-webrtc`. L'ancienne chaîne
  AAC → copie AAC → Opus dégradait le son.
- Les réglages Wi-Fi (`networks`), les réglages d'encodeur DJI et le toggle
  « Convertir en H.264 » sont retirés : l'encodage est désormais obligatoire
  pour WebRTC.

## Acceptance Criteria

- [x] Plus aucune référence à BlueZ, `djictl`, rfkill ou au scan BLE.
- [x] MediaMTX ne sert plus HLS et la sortie RTMP pour OBS a disparu de l'UI.
- [x] L'URL affichée est celle de la page WHEP `live/dji-webrtc`.
- [x] `live/dji-webrtc` contient H.264 + Opus, produits par une seule passe.
- [x] `go build ./...`, `go vet ./...`, `go test ./...` et `npm run build` passent.
- [x] `go test -tags manual -run TestWebrtcOutput` valide la sortie réelle.
- [x] Le paquet `.deb` est reconstruit en 0.2.0.

## Notes

Vérifié avec le vrai MediaMTX : `[path live/dji-webrtc] … 2 tracks (H264,
Opus)` et la page WebRTC répond en HTTP 200. FFmpeg accepte bien l'Opus dans
le conteneur FLV utilisé pour publier sur MediaMTX.

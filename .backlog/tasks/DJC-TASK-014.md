---
id: "DJC-TASK-014"
type: "TASK"
status: "done"
priority: "high"
labels: ["latence", "hls", "webrtc", "mediamtx", "ffmpeg"]
components: ["backend", "frontend"]
story_points: 5
epic: null
sprint: null
parent: null
children: []
depends_on: []
blocks: []
related_to: ["DJC-TASK-010", "DJC-BUG-009"]
tokens_estimated: 20000
tokens_used: 24000
created: "2026-11-19"
updated: "2026-11-19"
---
## Summary

Exposer le flux en HLS basse latence et en WebRTC pour comparer les latences,
et rendre ces liens utilisables depuis l'interface.

## Description

MediaMTX produit déjà les trois sorties depuis le même chemin publié : RTMP
(`:1935`, OBS), HLS (`:8888`) et WebRTC (`:8889`, page WHEP intégrée, médias en
UDP `:8189`). La configuration générée par l'app les active avec des réglages
orientés temps réel :

- `hlsVariant: lowLatency`, parts de 200 ms, segments d'1 s, `hlsAlwaysRemux`
  (pas de délai au premier lecteur) ; MediaMTX exige `hlsSegmentCount: 7`.
- `readTimeout: 4s`, `writeTimeout: 1s`, `writeQueueSize: 128`, `moq: no`.
- ffmpeg : `repeat-headers=1` (SPS/PPS dans le flux pour un lecteur WebRTC qui
  rejoint en cours), keyint fixe + `-force_key_frames` à 1 s, `-muxdelay 0`
  `-muxpreload 0`.

Côté UI, une carte « Lecture basse latence » affiche les deux URL (réseau ou
127.0.0.1), permet de copier et d'ouvrir dans le navigateur, et prévient que
WebRTC n'accepte que le H.264 (conversion nécessaire si la caméra est en HEVC).

## Acceptance Criteria

- [x] Le `mediamtx.yml` généré active HLS basse latence et WebRTC.
- [x] La playlist maître et sa variante contiennent des `EXT-X-PART`.
- [x] La page WebRTC du chemin répond en HTTP 200.
- [x] Le transcodage ffmpeg publie `live/dji-h264` avec les nouveaux réglages.
- [x] Les URL HLS/WebRTC suivent le chemin réellement lu (H.264 si conversion).
- [x] `go build ./...`, `go vet ./...`, `go test ./...` et `npm run build` passent.

## Notes

Mesures avec `internal/media/manual_outputs_test.go`
(`go test -tags manual -run TestLowLatencyOutputs`) : ffmpeg pousse un flux de
test, MediaMTX sert HLS + WebRTC, puis le transcoder republie en H.264.

`readTimeout` : 1 s et 2 s coupent le publieur RTMP pendant sa négociation
(`i/o timeout`, aucun publish enregistré) ; 3 s passe juste, 4 s est retenu.

`protocolMaxDelay` demandé **n'existe pas** dans MediaMTX 1.21.4 : les leviers
réels sont `readTimeout`/`writeTimeout`, `writeQueueSize`, `hlsPartDuration`/
`hlsSegmentDuration` et `webrtcHandshakeTimeout`.

Reste à mesurer la latence réelle bout en bout des trois sorties : le test
vérifie la disponibilité des flux, pas les millisecondes.

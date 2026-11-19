---
id: "DJC-BUG-009"
type: "BUG"
status: "review"
priority: "high"
labels: ["desktop", "ffmpeg", "obs", "latence"]
components: ["internal/media"]
story_points: 2
epic: null
sprint: null
parent: null
children: []
depends_on: []
blocks: []
related_to: ["DJC-TASK-010", "DJC-BUG-008"]
tokens_estimated: 10000
tokens_used: 12000
created: "2026-11-19"
updated: "2026-11-19"
---
## Summary

~4 s de latence entre le DJI Osmo Pocket et OBS avec la conversion H.264.

## Description

- `-r 30` en option d'entrée écrasait `-use_wallclock_as_timestamps` pour la
  vidéo (timestamps = compteur d'images). Après chaque micro-coupure Wi-Fi,
  la rafale d'images était étalée à 1/30 s et rejouée en temps réel : le
  retard s'accumulait définitivement, et la vidéo dérivait de l'audio
  (resté sur l'horloge murale).
- `-preset veryfast` (décodage HEVC + x264 1080p logiciels) pouvait passer
  sous 1x sur une machine modeste, sans aucune trace dans les logs.

## Fix

- `-r 30` déplacé en sortie avec `-fps_mode cfr` : écrêtage des rafales,
  duplication sur les trous, latence constante.
- `-preset ultrafast`.
- `-progress pipe:1 -stats_period 2` lu par `watchProgress` : alerte dans les
  logs si `speed` < 0.95x sur 3 rapports, puis message de rétablissement.

## Acceptance Criteria

- [x] Aucun `-r` avant `-i` (test `TestTranscodeArgsNoInputFramerate`).
- [x] Alerte de lenteur testée (`TestWatchProgressWarnsOnlyOnSustainedSlowness`).
- [x] `go build`/`go vet`/`go test ./internal/...` OK.
- [ ] Mesure terrain DJI → OBS (tampon 0 Mo) : latence < 1,5 s et stable sur 10 min.
- [ ] Non-régression relais Android (DJC-BUG-008) : pas de gel ni de saccade.

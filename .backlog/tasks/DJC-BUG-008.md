---
id: "DJC-BUG-008"
type: "BUG"
status: "review"
priority: "high"
labels: ["desktop", "ffmpeg", "obs", "rtmp"]
components: ["internal/media", "app"]
story_points: 3
epic: null
sprint: null
parent: null
children: []
depends_on: []
blocks: []
related_to: ["DJC-BUG-007", "DJC-TASK-008"]
tokens_estimated: 12k
tokens_used: 18k
created: "2026-11-17"
updated: "2026-11-17"
---
## Summary

Le flux du relais Android fait figer OBS après ~20-30 s : image bloquée,
socket muette, MediaMTX finit par fermer la session lecteur qui ne se
reconnecte pas. Le rendu OBS est en plus saccadé et l'audio décalé.

## Description

Le flux publié par le relais Android (RootEncoder) a des timestamps H.264/AAC
invalides :

- DTS vidéo égaux/répétés (non monotones) ;
- audio démarrant à -2,7 s par rapport à la vidéo ;
- images de référence H.264 occasionnellement manquantes.

Constaté avec ffprobe sur `rtsp://127.0.0.1:8554/live/dji` : pts audio
réguliers (23 ms) mais départ négatif, dts vidéo répétés. OBS media source
s'arrête net sur un DTS non monotone sans jamais se reconnecter.

Le transcodage de l'app (`internal/media/transcoder.go`) aggravait le
problème :

- `-fflags +genpts` seul dérivait sur les timestamps cassés (framerate
  apparent ~90 000 fps, `MB rate (324000000) > level limit`) → il fallait
  forcer le framerate d'entrée ;
- `-af aresample=async=1:first_pts=0` remettait l'audio à 0 pendant que la
  vidéo partait sur l'horloge murale → décalage A/V de ~70 s, que OBS
  « attend » indéfiniment (cause principale de la latence perçue).

## Fix

`transcodeArgs` (republication « Convertir en H.264 ») :

- `-use_wallclock_as_timestamps 1` : timestamps monotones, A/V recalé à
  l'arrivée des paquets (le décalage de la source disparaît) ;
- `-r 30` en entrée : empêche la dérivation de genpts ;
- suppression de `first_pts=0` : l'audio suit l'horloge murale comme la
  vidéo ;
- le CFR 30 fps + zerolatency lisse les rafales du Wi-Fi du téléphone
  (les timestamps horloge-murale bruts sont groupés en rafales : ~88 % des
  paquets consécutifs partagent le même timestamp — impassables en copie
  directe, d'où le choix du transcodage plutôt que `-c copy`).

## Acceptance Criteria

- [x] Lecture de 30 s de `live/dji-h264` sans aucun warning DTS.
- [x] Vidéo 30 fps exacte, audio régulier (23,2 ms/paquet), 0 timestamp
      non monotone (validé sur flux de test MediaMTX).
- [x] Écart A/V stable (< 0,5 s au connect, ~30 ms en régime permanent).
- [x] `go build ./... && go vet ./...` OK ; `wails build -tags webkit2_41` OK.
- [ ] Validation en conditions réelles avec le relais Android + OBS
      (nécessite la reconnexion auto du relais, cf. DJC-TASK-008).

## Notes

Une republication en `-c copy` (restampage seul) a été écartée : l'extradata
H.264 n'arrive pas correctement dans l'en-tête FLV côté lecteurs (vidéo
« unspecified size ») et les timestamps horloge-murale restent groupés en
rafales. Les frais CPU du transcodage veryfast 720p restent raisonnables.

Piège de test : simuler la caméra avec `ffmpeg -re -f lavfi` — sans `-re`,
la source publie plus vite que le temps réel et fausse toutes les mesures.

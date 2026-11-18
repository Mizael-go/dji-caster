---
id: "DJC-TASK-010"
type: "TASK"
status: "review"
priority: "high"
labels: ["desktop", "ffmpeg", "obs", "latence"]
components: ["internal/media", "frontend", "landing"]
story_points: 2
epic: null
sprint: null
parent: null
children: []
depends_on: ["DJC-BUG-008"]
blocks: []
related_to: ["DJC-BUG-008", "DJC-TASK-009"]
tokens_estimated: 8k
tokens_used: 9k
created: "2026-11-18"
updated: "2026-11-18"
---
## Summary

Après le correctif des timestamps, il reste 1 à 2 s de latence entre le
relais Android et OBS. Objectif : s'en approcher au maximum du temps réel.

## Description

La latence résiduelle se répartit entre l'encodeur du téléphone, le
transcodage ffmpeg et surtout le tampon réseau de la source média OBS
(1 Mo par défaut, soit plusieurs secondes de flux).

## Fix

- `-flush_packets 1` sur la sortie FLV : ffmpeg n'attend plus de remplir un
  tampon avant d'envoyer les paquets à MediaMTX.
- Documentation : l'interface desktop affiche le réglage OBS à faire
  (tampon réseau 0 Mo) sous l'URL de source ; la landing détaille les
  quatre gisements de latence et la commande ffmpeg complète.
- Les réglages déjà en place (zerolatency, `-bf 0`, GOP 1 s,
  `-use_wallclock_as_timestamps 1`) sont conservés.

## Acceptance Criteria

- [x] `-flush_packets 1` appliqué dans `transcodeArgs`.
- [x] Astuce OBS (tampon 0 Mo) visible dans l'onglet Diffusion.
- [x] Landing mise à jour (checklist latence + commande ffmpeg).
- [x] `go build ./... && go vet ./...` et `npm run build` OK.
- [ ] Mesure terrain avec OBS tampon 0 Mo : latence perçue < 1 s.

## Notes

Le facteur restant côté relais est l'encodage RootEncoder lui-même ; sa
correction (DJC-TASK-009) supprimerait le transcodage et sa part de latence.

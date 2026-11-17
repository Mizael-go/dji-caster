---
id: "DJC-TASK-005"
type: "TASK"
status: "todo"
priority: "medium"
labels: ["tests", "pipeline"]
components: ["backend"]
story_points: 5
epic: null
sprint: null
parent: null
children: []
depends_on: []
blocks: []
related_to: []
tokens_estimated: 25000
tokens_used: 0
created: "2026-11-17"
updated: "2026-11-17"
---
## Summary

Test d'intégration du pipeline MediaMTX + ffmpeg.

## Description

Le transcodage HEVC → H.264 n'a été validé que manuellement (71 frames en 4 s).
Un test Go (`-tags integration`) doit démarrer MediaMTX sur des ports libres,
publier un fichier HEVC de quelques secondes en RTMP, laisser ffmpeg republier
et vérifier les codecs du chemin de sortie via l'API locale.

## Acceptance Criteria

- [ ] Le test choisit des ports libres et n'utilise que des binaires locaux.
- [ ] Il vérifie `codecs` contient H264 sur `live/dji-h264` et échoue sinon.
- [ ] Il nettoie ses processus et ses fichiers temporaires.
- [ ] `go test -tags integration ./internal/media/...` passe hors ligne avec les
      binaires de `~/.config/dji-caster/bin`.

## Notes

Le relais Android (`DJC-TASK-001`) fournit une source H.264 réelle : le même
test peut couvrir le passage direct sans transcodage.

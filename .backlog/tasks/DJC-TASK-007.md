---
id: "DJC-TASK-007"
type: "TASK"
status: "todo"
priority: "low"
labels: ["obs", "configuration"]
components: ["frontend", "backend"]
story_points: 1
epic: null
sprint: null
parent: null
children: []
depends_on: []
blocks: []
related_to: []
tokens_estimated: 4000
tokens_used: 0
created: "2026-11-17"
updated: "2026-11-17"
---
## Summary

L'URL affichée pour OBS ne suit pas toujours le port configuré.

## Description

`StreamView.ObsURL` est construite avec `settings.RtmpPort`, mais le bandeau de
session et certains messages rappellent `rtmp://IP:1935/...` en dur. Si
l'utilisateur change le port dans les Réglages, l'URL copiée et l'URL affichée
divergent.

## Acceptance Criteria

- [ ] Toutes les URL affichées passent par `media.RtmpURL(host, port, path)`.
- [ ] Le port affiché dans l'en-tête est celui du serveur en cours d'exécution.
- [ ] `cd frontend && npm run build` passe et les types Go/TS restent en miroir.

## Notes

Vérifier aussi `connect-dialog` et `session-banner` : la cible RTMP y est
affichée en bandeau.

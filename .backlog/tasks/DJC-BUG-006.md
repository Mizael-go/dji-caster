---
id: "DJC-BUG-006"
type: "BUG"
status: "todo"
priority: "medium"
labels: ["multi-camera", "sessions"]
components: ["backend", "frontend"]
story_points: 8
epic: null
sprint: null
parent: null
children: []
depends_on: ["DJC-BUG-004"]
blocks: []
related_to: ["DJC-TASK-001"]
tokens_estimated: 35000
tokens_used: 0
created: "2026-11-17"
updated: "2026-11-17"
---
## Summary

Une seule session caméra : l'état est global, deux appareils se marchent dessus.

## Description

`sessionState`, `sessionCmd` et le transcoder sont uniques dans `App`. Connecter
une deuxième caméra (ou le relais Android pendant qu'une DJI émet) écrase la
session en cours, et les deux publications se disputent le même chemin
`live/dji`.

## Acceptance Criteria

- [ ] Chemins par session (`live/cam1`, `live/cam2`, …) et sélection dans
      l'onglet Diffusion.
- [ ] Chaque session a son état, son erreur et son entrée de journal.
- [ ] Le transcoder est instancié par session et arrêté avec elle.
- [ ] Une session existante n'est jamais écrasée par une nouvelle connexion.

## Notes

Le relais Android est le meilleur moyen de tester deux sources sans matériel.

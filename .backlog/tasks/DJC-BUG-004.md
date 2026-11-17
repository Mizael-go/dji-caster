---
id: "DJC-BUG-004"
type: "BUG"
status: "todo"
priority: "high"
labels: ["mediamtx", "robustesse"]
components: ["backend"]
story_points: 3
epic: null
sprint: null
parent: null
children: []
depends_on: []
blocks: []
related_to: []
tokens_estimated: 9000
tokens_used: 0
created: "2026-11-17"
updated: "2026-11-17"
---
## Summary

Un MediaMTX orphelin garde le port 1935 et bloque le démarrage suivant.

## Description

Quand l'application est interrompue (redémarrage de `wails dev`, crash), le
processus `mediamtx` de `~/.config/dji-caster/bin` survit et écoute encore sur
1935. Au lancement suivant, le serveur échoue avec « port déjà utilisé » et
l'utilisateur doit tuer le processus à la main.

## Acceptance Criteria

- [ ] Le PID de MediaMTX est écrit dans `~/.config/dji-caster/mediamtx.pid` et
      supprimé à l'arrêt.
- [ ] Au démarrage, si le port est occupé : vérifier que c'est bien notre PID
      (ou notre API `:9997`) avant de le terminer, sinon proposer un autre port
      dans l'UI.
- [ ] Aucun `pkill -f` par nom : cibler le PID enregistré ou l'API locale.
- [ ] `go build ./... && go vet ./...` passe, test manuel : tuer la fenêtre
      brutalement puis relancer.

## Notes

Le watcher `mediaWatcher` surveille déjà l'API `:9997` : la même sonde peut
servir à identifier un serveur survivant.

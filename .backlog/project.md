---
project: "DJI Caster"
prefix: "DJC"
version: "0.1.0"
created: "2026-11-17"
updated: "2026-11-17"
external_projects: []
components:
  - backend        # app.go + internal/*
  - frontend       # Wails / React
  - mobile         # relais Android de test (Capacitor)
  - landing        # page de téléchargement
  - packaging      # .deb
token_budget:
  per_task_warn: 40000
  sprint: 200000
---

# DJI Caster

Application desktop Linux (Wails v2 + Go + React) qui force une caméra DJI Osmo
à rejoindre le Wi-Fi local puis expose son flux RTMP (MediaMTX + ffmpeg) pour OBS.

## Où vit le travail

| Sujet | Emplacement |
| --- | --- |
| Backend Go, bindings | `app.go`, `internal/` |
| Interface | `frontend/src` |
| Relais Android de test | `mobile/` |
| Paquet `.deb` | `scripts/build-deb.sh`, `dist/` |
| Site | `landing/` |
| Notes humaines | `~/Documents/Obsidian Vault/dji-caster/` |

## Conventions

- Préfixe des tickets : `DJC-TASK-`, `DJC-BUG-`, `DJC-SPIKE-`, `DJC-CHORE-`, `DJC-STORY-`.
- Un ticket = un fichier dans `tasks/`, `backlog.md` reste la source de vérité du statut.
- `tokens_estimated` / `tokens_used` en convention locale (voir `skills/backlog/SKILL.md`).
- Vérification minimale avant `done` : `go build ./... && go vet ./...` et
  `cd frontend && npm run build` (ou `gradle assembleDebug` dans le conteneur pour `mobile/`).

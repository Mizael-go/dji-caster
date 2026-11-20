# AGENTS.md — DJI Caster

Application desktop (Wails v2 + Go + React) qui reçoit un flux RTMP (caméra DJI
ou relais Android) et le publie en WebRTC, seule sortie de lecture.

## Stack

- Backend : Go 1.26, Wails v2.12.
- Frontend : React 19, TypeScript, Vite 6, Tailwind v4, shadcn (style
  `radix-nova`, composants copiés dans `frontend/src/components/ui`),
  react-query, react-hook-form.
- Processus externes : MediaMTX (copie propre à l'app), ffmpeg (système).

## Commandes

```bash
wails dev   -tags webkit2_41     # développement (hot reload)
wails build -tags webkit2_41     # binaire : build/bin/dji-caster
cd frontend && npm run build     # typecheck + build (tsc && vite build)
go build ./... && go vet ./...   # backend
wails generate module            # après ajout/modif d'une méthode bindée
```

Le tag `webkit2_41` est requis dès que la distribution n'a que
`webkit2gtk-4.1` (Ubuntu 24.04+). Sans lui, la compilation échoue sur
`Package 'webkit2gtk-4.0' was not found`.

## Architecture

| Chemin             | Rôle                                                       |
| ------------------ | ---------------------------------------------------------- |
| `main.go`          | options de fenêtre (420×860, colonne unique), embed des assets |
| `app.go`           | `*App` : bindings RPC + état + orchestration des processus |
| `internal/config`  | `settings.json` (0600)                                      |
| `internal/media`   | serveur MediaMTX, encodage ffmpeg H.264 + Opus, IP locale   |
| `internal/bins`    | téléchargement MediaMTX                                     |
| `frontend/src/lib/api.ts` | unique point d'entrée vers les bindings Wails       |

## Règles à respecter

- **Aucun appel HTTP direct depuis le frontend** : pas d'axios, pas de fetch.
  Tout passe par les bindings générés (`wailsjs/go/main/App`).
- **MediaMTX toujours local à l'app** (`~/.config/dji-caster/bin`) : ne jamais
  utiliser un MediaMTX installé sur le système, l'app doit rester portable.
- **RTMP n'est que l'entrée.** ffmpeg lit le **RTSP loopback**
  (`127.0.0.1:8554/live/dji`) et publie `live/dji-webrtc` en H.264 + Opus en
  une seule passe. Ne jamais réintroduire une copie AAC intermédiaire : c'est
  ce qui dégradait le son. Détails : `skills/ffmpeg/SKILL.md`.
- **WebRTC est la seule sortie** (HTTP `:8889`, médias UDP `:8189`). HLS et la
  sortie RTMP ont été retirées volontairement.
- Tout processus lancé (`mediamtx`, `ffmpeg`) doit être arrêté dans
  `OnShutdown`, et ffmpeg tourne dans son propre groupe de processus.
- Les types de `app.go` (tags JSON) et `frontend/src/lib/types.ts` doivent
  rester identiques ; mettre à jour les deux ensemble.
- `tsconfig.json` n'inclut que `src` : ne pas y ajouter `wailsjs` (les `.js`
  générés ne passent pas en mode strict).
- Toute modification visuelle suit `skills/design-systems/SKILL.md` (tokens,
  pas de couleurs en dur).

## Skills du projet

- `skills/backlog/SKILL.md` — backlog `.backlog/` + budget de tokens.
- `skills/design-systems/SKILL.md` — tokens et règles UI.
- `skills/wails/SKILL.md` — conventions Wails v2 / bindings / events.
- `skills/ffmpeg/SKILL.md` — pipeline RTMP/RTSP/WebRTC, commande d'encodage.

## Vérification minimale

`go build ./... && go vet ./...` puis `cd frontend && npm run build` (tsc inclus).
Pour un changement touchant la fenêtre ou les options Wails, lancer aussi
`wails build -tags webkit2_41`.

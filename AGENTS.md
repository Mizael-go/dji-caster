# AGENTS.md — DJI Caster

Application desktop (Wails v2 + Go + React) qui force une caméra DJI Osmo à
rejoindre le Wi-Fi de la machine puis expose son flux RTMP pour OBS.

## Stack

- Backend : Go 1.26, Wails v2.12 (`github.com/godbus/dbus/v5` pour BlueZ).
- Frontend : React 19, TypeScript, Vite 6, Tailwind v4, shadcn (style
  `radix-nova`, composants copiés dans `frontend/src/components/ui`),
  react-query, react-hook-form.
- Processus externes : MediaMTX (copie propre à l'app), djictl (compilé
  localement), ffmpeg (système).

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
| `internal/config`  | `settings.json` (0600, contient le PSK Wi-Fi)              |
| `internal/ble`     | découverte BlueZ via D-Bus (pas de parsing de bluetoothctl) |
| `internal/djictl`  | lance `djictl ble connect-wifi-and-start-streaming`        |
| `internal/media`   | serveur MediaMTX, transcodage ffmpeg, IP/SSID locaux       |
| `internal/bins`    | téléchargement MediaMTX, compilation djictl                |
| `frontend/src/lib/api.ts` | unique point d'entrée vers les bindings Wails       |

## Règles à respecter

- **Aucun appel HTTP direct depuis le frontend** : pas d'axios, pas de fetch.
  Tout passe par les bindings générés (`wailsjs/go/main/App`).
- **MediaMTX toujours local à l'app** (`~/.config/dji-caster/bin`) : ne jamais
  utiliser un MediaMTX installé sur le système, l'app doit rester portable.
- Le transcodage lit le **RTSP loopback** (`127.0.0.1:8554`) : MediaMTX ne
  forwarde pas le H.265 aux lecteurs RTMP classiques. Détails :
  `skills/ffmpeg/SKILL.md`.
- Tout processus lancé (`mediamtx`, `ffmpeg`, `djictl`) doit être arrêté dans
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
- `skills/ble/SKILL.md` — BlueZ via D-Bus, pièges du scan.
- `skills/ffmpeg/SKILL.md` — pipeline RTMP/RTSP, commande de transcodage.

## Vérification minimale

`go build ./... && go vet ./...` puis `cd frontend && npm run build` (tsc inclus).
Pour un changement touchant la fenêtre ou les options Wails, lancer aussi
`wails build -tags webkit2_41`.

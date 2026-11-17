---
name: wails
description: >-
  Conventions Wails v2 pour DJI Caster : options de fenêtre, bindings Go↔React,
  événements, build Linux (webkit2gtk). À lire avant de toucher main.go, app.go
  ou les appels du frontend. Réf. : https://github.com/wailsapp/awesome-wails
---

# Wails v2 — règles du projet

## Commandes

```bash
wails dev                 -tags webkit2_41   # hot reload
wails build               -tags webkit2_41   # binaire dans build/bin/
wails generate module                        # régénère frontend/wailsjs
```

Le tag `webkit2_41` est obligatoire sur les distributions récentes
(Ubuntu 24.04+ n'a plus `webkit2gtk-4.0`). Sans lui : `Package
'webkit2gtk-4.0' was not found`.

## Structure

```
main.go              options.App (fenêtre 420×860), //go:embed all:frontend/dist
app.go               *App = seul objet bindé ; toutes les vues JSON y sont
internal/config      settings.json (0600, contient le PSK)
internal/ble         client BlueZ via D-Bus
internal/djictl      exécution du CLI djictl
internal/media       MediaMTX (serveur + config générée) et ffmpeg
internal/bins        téléchargement MediaMTX / compilation djictl
frontend/src         React 19 + TS + Tailwind v4 + shadcn
```

## Règles

- **Pas d'axios, pas de fetch** : tout passe par les bindings générés
  (`frontend/wailsjs/go/main/App`) enveloppés dans `src/lib/api.ts`.
- Une méthode Go exportée sur `*App` devient un appel RPC : la garder
  synchrone et courte. Les opérations longues (djictl, installation) tournent
  dans une goroutine et se lisent via `Snapshot()` / les logs.
- Le backend pousse `EventsEmit(ctx, "state", "change")` (throttlé 200 ms) ; le
  frontend écoute avec `EventsOn("state", …)` et invalide react-query.
- Les types renvoyés au frontend vivent dans `app.go` (avec tags JSON) et sont
  dupliqués à la main dans `src/lib/types.ts` — garder les deux alignés.
- `OnShutdown` doit arrêter BlueZ, djictl, ffmpeg et MediaMTX (pas de processus
  orphelin).
- Les fichiers générés (`frontend/wailsjs`, `frontend/dist`) ne s'éditent pas.
- `tsconfig.json` n'inclut que `src` (sinon tsc type-check les .js générés).

## Débogage

`wails dev` ouvre aussi un serveur web sur `http://localhost:34115` : pratique
pour inspecter l'UI sans la fenêtre native. Les logs Go des processus externes
sont visibles dans l'onglet « Journaux ».

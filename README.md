# DJI Caster

Petite app desktop (fenêtre au format « cast mobile ») qui récupère le flux
d'une caméra DJI Osmo sans DJI MIMO et l'expose en RTMP pour OBS.

```
DJI ──BLE(djictl)──▶ Wi-Fi local ──RTMP──▶ MediaMTX ──▶ OBS
                                            └─(option) ffmpeg H.265→H.264 ──▶ OBS
```

## Ce que fait l'app

1. **Scan Bluetooth** des appareils à portée (BlueZ/D-Bus), caméras DJI en tête.
2. **Connexion forcée** : `djictl` demande à la caméra de rejoindre votre Wi-Fi
   puis de pousser vers `rtmp://<ip-locale>:1935/live/dji`.
3. **Sortie RTMP** copiable pour OBS, avec conversion H.264 optionnelle si OBS
   refuse le HEVC de la caméra.

## Prérequis

- Go ≥ 1.23, Node ≥ 20, Wails v2 (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
- GTK3 + WebKitGTK (`webkit2gtk-4.1` sur Ubuntu 24.04+, `4.0` ailleurs)
- `bluetoothd` actif et un adaptateur BLE
- `ffmpeg` dans le `PATH` (transcodage optionnel)
- Recommandé : NetworkManager (détection du SSID actif)

MediaMTX et djictl n'ont pas à être installés : l'app les télécharge/compile
dans `~/.config/dji-caster/bin` (bouton « Installer les outils » ou au premier
démarrage). Aucun MediaMTX système n'est utilisé.

## Lancer

```bash
wails dev   -tags webkit2_41     # développement
wails build -tags webkit2_41     # build/bin/dji-caster
```

Le tag `webkit2_41` est nécessaire quand seul `webkit2gtk-4.1` est disponible.

## Ports et fichiers

| Élément            | Valeur                                                        |
| ------------------ | ------------------------------------------------------------- |
| RTMP (public)      | `:1935` — `rtmp://<ip>:1935/live/dji`                         |
| RTSP (loopback)    | `127.0.0.1:8554` — lecture du H.265 par ffmpeg                |
| API MediaMTX       | `127.0.0.1:9997` — état des flux (lecture seule)              |
| Flux transcodé     | `live/dji-h264`                                               |
| Config de l'app    | `~/.config/dji-caster/settings.json` (0600)                   |
| Config MediaMTX    | `~/.config/dji-caster/mediamtx.yml` (générée)                 |
| Binaires           | `~/.config/dji-caster/bin/`                                   |

## Notes

- Le mot de passe Wi-Fi est stocké en clair dans un fichier 0600 : c'est un
  outil local, pas un coffre-fort.
- `djictl` doit pouvoir parler au Bluetooth ; si l'accès est refusé, l'app
  affiche la cause (polkit / groupe `bluetooth`). L'app ne lance jamais `sudo`.
- Pour le relais Android, activez « Convertir en H.264 » dans l'onglet
  Diffusion puis utilisez l'URL `live/dji-h264` affichée pour OBS : le flux
  RootEncoder direct (`live/dji`) peut avoir des timestamps incorrects et
  apparaître en retard ou se figer. Pour les autres caméras, activez la
  conversion si OBS ne prend pas en charge le codec reçu.
- Le Wi-Fi ciblé doit être en 2,4 GHz (contrainte des caméras DJI).

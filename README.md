# DJI Caster

Passerelle locale qui reçoit un flux RTMP (caméra DJI ou relais Android) et le
publie en **WebRTC**, la sortie de lecture à la plus basse latence.

```
Caméra / relais ──RTMP──▶ MediaMTX ──RTSP loopback──▶ ffmpeg ──▶ MediaMTX ──WebRTC──▶ navigateur
```

## Ce que fait l'app

1. **Serveur RTMP** local (`:1935`) : la caméra ou l'app mobile y pousse son
   flux sur `live/dji`.
2. **Encodage WebRTC** : un seul ffmpeg lit la source en RTSP loopback et
   réencode la vidéo en H.264 et l'audio en **Opus** (une seule passe, pas de
   copie AAC intermédiaire). Le résultat est publié sur `live/dji-webrtc`.
3. **Sortie WebRTC** : MediaMTX sert une page de lecture WHEP sur `:8889`,
   copiable et ouvrable dans le navigateur.

RTMP ne sert plus que d'entrée ; HLS et la sortie RTMP pour OBS ont été
retirés. Le Bluetooth (scan, forçage DJI par `djictl`) a également été retiré :
la caméra se connecte via le relais Android ou est déjà sur le Wi-Fi.

## Prérequis

- Go ≥ 1.23, Node ≥ 20, Wails v2 (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
- GTK3 + WebKitGTK (`webkit2gtk-4.1` sur Ubuntu 24.04+, `4.0` ailleurs)
- `ffmpeg` dans le `PATH` (obligatoire : vidéo H.264 + audio Opus)

MediaMTX n'a pas à être installé : l'app le télécharge dans
`~/.config/dji-caster/bin` (bouton « Installer MediaMTX » ou au premier
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
| RTMP (entrée)      | `:1935` — `rtmp://<ip>:1935/live/dji`                         |
| RTSP (loopback)    | `127.0.0.1:8554` — lecture de la source par ffmpeg            |
| WebRTC (lecture)   | `:8889` — page WHEP, médias UDP `:8189`                       |
| API MediaMTX       | `127.0.0.1:9997` — état des flux (lecture seule)              |
| Flux WebRTC        | `live/dji-webrtc` (H.264 + Opus)                              |
| Config de l'app    | `~/.config/dji-caster/settings.json` (0600)                   |
| Config MediaMTX    | `~/.config/dji-caster/mediamtx.yml` (générée)                 |
| Binaires           | `~/.config/dji-caster/bin/`                                   |

## Notes

- ffmpeg réencode toujours la vidéo en H.264 et l'audio en Opus : c'est ce qui
  rend la sortie WebRTC lisible partout, y compris quand la caméra émet en
  HEVC ou avec des timestamps irréguliers.
- Si le transcodage passe sous le temps réel (`speed < 0.95x`), l'app le
  signale dans les journaux : réduisez la résolution ou le débit de la source.
- Le relais Android pousse directement sur l'URL d'entrée affichée dans
  l'onglet Diffusion.

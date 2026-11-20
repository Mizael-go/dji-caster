---
name: ffmpeg
description: >-
  Pipeline vidéo de DJI Caster : MediaMTX en entrée RTMP, lecture de la source
  en RTSP loopback, ré-encodage H.264 + Opus en une passe, puis lecture WebRTC.
  À lire avant de modifier les commandes ffmpeg, la config MediaMTX ou l'export.
  Réf. : https://github.com/u2takey/ffmpeg-go
---

# Pipeline vidéo

```
Source ──RTMP──▶ MediaMTX :1935 /live/dji
                      │
                      └──RTSP 127.0.0.1:8554──▶ ffmpeg (H.264 + Opus)
                                                    │
                                                    ▼
                              MediaMTX /live/dji-webrtc ──WebRTC :8889──▶ navigateur
```

RTMP n'est **que l'entrée**. La seule sortie de lecture est WebRTC : HLS et la
sortie RTMP pour OBS ont été retirées. Le Bluetooth et `djictl` aussi.

## Pourquoi Rtsp pour l'encodage

MediaMTX ne forwarde pas le H.265 aux lecteurs RTMP « classiques » (il faut un
client enhanced-RTMP) : un `ffmpeg -i rtmp://…` ne voit que la piste audio.
En RTSP, le H.265 passe sans problème. RTSP n'écoute qu'en loopback, jamais
exposé au réseau.

## Commande d'encodage (référence)

```bash
ffmpeg -hide_banner -nostdin -loglevel warning -nostats -progress pipe:1 -stats_period 2 \
  -flags low_delay -use_wallclock_as_timestamps 1 -fflags +genpts \
  -analyzeduration 1000000 -probesize 1000000 \
  -rtsp_transport tcp -i rtsp://127.0.0.1:8554/live/dji \
  -c:v libx264 -preset ultrafast -tune zerolatency -profile:v high -pix_fmt yuv420p \
  -g 30 -keyint_min 30 -sc_threshold 0 -bf 0 \
  -x264-params rc-lookahead=0:sync-lookahead=0:repeat-headers=1:keyint=30:min-keyint=30:scenecut=0 \
  -force_key_frames 'expr:gte(t,n_forced*1)' \
  -b:v 6M -maxrate 6M -bufsize 6M -r 30 -fps_mode cfr \
  -c:a libopus -b:a 160k -ar 48000 -ac 2 -application audio -frame_duration 20 \
  -af aresample=async=1 \
  -flush_packets 1 -muxdelay 0 -muxpreload 0 \
  -f flv rtmp://127.0.0.1:1935/live/dji-webrtc
```

## Audio : une seule passe

WebRTC ne transporte que l'Opus, le G722 et le G711. L'AAC de la caméra est
silencieusement ignoré par le navigateur, et MediaMTX ne convertit aucun codec.

- La vidéo **et** l'audio sont réencodés dans le même ffmpeg, directement
  depuis la source. Ne jamais réintroduire de copie AAC intermédiaire : c'est
  l'enchaînement AAC → copie AAC → Opus qui dégradait le son (DJC-TASK-016).
- L'Opus est encodé en 48 kHz stéréo, 160 kb/s, `-application audio` (meilleure
  qualité que `lowdelay` à ce débit), trames de 20 ms (défaut WebRTC).
- FFmpeg accepte l'Opus dans le conteneur FLV utilisé pour publier sur
  MediaMTX (vérifié : `2 tracks (H264, Opus)`).

## Pièges

- Ne pas ajouter `-fflags nobuffer` : ffmpeg rate alors la piste vidéo.
- **Jamais de `-r` en entrée** : il remplace l'horloge murale par un compteur
  d'images ; après chaque coupure Wi-Fi la rafale est rejouée en temps réel et
  la latence s'accumule (DJC-BUG-009). La cadence se fixe en sortie
  (`-r 30 -fps_mode cfr`), ce qui jette les images en trop et duplique sur un trou.
- `-preset ultrafast` : décodage HEVC logiciel + x264 1080p doit tenir ≥ 1x.
  `-progress pipe:1` est lu par `watchProgress`, qui alerte dans les logs si
  `speed` reste < 0.95x (latence croissante).
- ffmpeg n'est lancé que si le chemin `/live/dji` est `ready` (API MediaMTX) et
  est relancé toutes les 3 s en cas de coupure (`internal/media/transcoder.go`).
- `repeat-headers=1` : SPS/PPS dans le flux, sinon un lecteur WebRTC qui rejoint
  en cours reste noir.
- `-g 30` + `-force_key_frames` = keyframe toutes les secondes : un lecteur
  qui rejoint accroche vite.

## Vérifier un flux

```bash
curl -s http://127.0.0.1:9997/v3/paths/list | jq '.items[] | {name, ready, tracks, readers}'
ffprobe -rtsp_transport tcp rtsp://127.0.0.1:8554/live/dji
curl -sI http://127.0.0.1:8889/live/dji-webrtc
```

Générer une fausse source HEVC pour tester :

```bash
ffmpeg -re -f lavfi -i testsrc2=size=1920x1080:rate=30 -f lavfi -i sine=frequency=440 \
  -c:v libx265 -preset ultrafast -tune zerolatency -pix_fmt yuv420p -c:a aac -t 30 \
  -f flv rtmp://127.0.0.1:1935/live/dji
```

## ffmpeg-go

`u2takey/ffmpeg-go` n'est pas utilisé : les commandes sont déjà explicites et
l'app doit garder le contrôle du cycle de vie (signaux, groupe de processus,
supervision). Si un besoin de composition apparaît, préférer `exec.Command`
avec des arguments typés plutôt qu'une couche de DSL.

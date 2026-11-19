---
name: ffmpeg
description: >-
  Pipeline vidéo de DJI Caster : MediaMTX en entrée RTMP, lecture H.265 en RTSP
  loopback, ré-encodage H.264 pour OBS. À lire avant de modifier les commandes
  ffmpeg, la config MediaMTX ou l'export. Réf. :
  https://github.com/u2takey/ffmpeg-go
---

# Pipeline vidéo

```
DJI ──RTMP──▶ MediaMTX :1935 /live/dji ──RTSP 127.0.0.1:8554──▶ ffmpeg ──RTMP──▶ /live/dji-h264 ──▶ OBS
                                      └───────────────────────────── RTMP direct ──▶ OBS
```

## Pourquoi RTSP pour le transcodage

MediaMTX ne forwarde pas le H.265 aux lecteurs RTMP « classiques » (il faut un
client enhanced-RTMP) : un `ffmpeg -i rtmp://…` ne voit que la piste audio.
En RTSP, le H.265 passe sans problème (vérifié : `tracks: [H264, MPEG-4 Audio]`
sur la sortie). RTSP n'écoute qu'en loopback, jamais exposé au réseau.

## Commande de transcodage (référence)

```bash
ffmpeg -hide_banner -nostdin -loglevel warning -nostats -progress pipe:1 -stats_period 2 \
  -flags low_delay -use_wallclock_as_timestamps 1 -fflags +genpts \
  -analyzeduration 1000000 -probesize 1000000 \
  -rtsp_transport tcp -i rtsp://127.0.0.1:8554/live/dji \
  -c:v libx264 -preset ultrafast -tune zerolatency -profile:v high -pix_fmt yuv420p \
  -g 30 -keyint_min 30 -sc_threshold 0 -bf 0 -x264-params rc-lookahead=0:sync-lookahead=0 \
  -b:v 6M -maxrate 6M -bufsize 6M -r 30 -fps_mode cfr \
  -c:a aac -b:a 128k -ar 44100 -af aresample=async=1 \
  -flush_packets 1 -f flv rtmp://127.0.0.1:1935/live/dji-h264
```

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
- `-g 30` = keyframe toutes les secondes à 30 fps : OBS démarre vite.

## Vérifier un flux

```bash
curl -s http://127.0.0.1:9997/v3/paths/list | jq '.items[] | {name, ready, tracks, readers}'
ffprobe -rtsp_transport tcp rtsp://127.0.0.1:8554/live/dji
```

Générer un faux DJI HEVC pour tester :

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

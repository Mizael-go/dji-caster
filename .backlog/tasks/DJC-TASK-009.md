---
id: "DJC-TASK-009"
type: "TASK"
status: "todo"
priority: "low"
labels: ["mobile", "android", "rtmp", "qualite"]
components: ["mobile"]
story_points: 5
epic: null
sprint: null
parent: null
children: []
depends_on: []
blocks: []
related_to: ["DJC-BUG-008", "DJC-TASK-008"]
tokens_estimated: 20k
tokens_used: 0
created: "2026-11-17"
updated: "2026-11-17"
---
## Summary

Produire des timestamps H.264/AAC corrects directement dans le relais
Android, pour s'affranchir du transcodage de secours sur le PC.

## Description

RootEncoder 2.8.2 publie des timestamps défectueux (DTS égaux/répétés,
audio démarrant avant la vidéo de ~2,7 s, cf. DJC-BUG-008). À ce jour le PC
compense via la republication ffmpeg (« Convertir en H.264 »), ce qui coûte
du CPU et ~0,5 s de latence.

Pistes à évaluer sur RootEncoder 2.8.2 :

- mode de timestamping du client RTMP (`getStreamClient()`, options de
  timestamp disponibles dans la lib generic/com.pedro.common) ;
- correction du démarrage audio par rapport à la vidéo (ordre
  prepareAudio/prepareVideo, première frame AAC) ;
- à défaut, montée de version RootEncoder (2.8.x+ / 3.x) si le défaut est
  corrigé en amont.

## Acceptance Criteria

- [ ] ffprobe sur `live/dji` : pts monotones sur les deux pistes, écart
      A/V au départ < 100 ms.
- [ ] Lecture OBS directe de `live/dji` sans basculer « Convertir en
      H.264 », stable sur plusieurs minutes.
- [ ] APK rebuild + redéploiement Vercel si le relais change.

# Backlog — DJI Caster

Mis à jour : 2026-11-20 · prochaine revue : à la prochaine session
Budget sprint : 200k tokens (convention `skills/backlog/SKILL.md`)

## in-progress

_(vide)_

## todo

| ID | Titre | Priorité | Points | Estimation | Dépend de |
| --- | --- | --- | --- | --- | --- |
| [DJC-BUG-004](tasks/DJC-BUG-004.md) | MediaMTX orphelin bloque le port 1935 | high | 3 | 9k | — |
| [DJC-BUG-006](tasks/DJC-BUG-006.md) | Une seule session caméra (multi-caméra) | medium | 8 | 35k | BUG-004 |
| [DJC-TASK-005](tasks/DJC-TASK-005.md) | Test d'intégration du pipeline HEVC→H.264 | medium | 5 | 25k | — |
| [DJC-TASK-007](tasks/DJC-TASK-007.md) | URL OBS incohérente avec le port configuré | low | 1 | 4k | — |
| [DJC-TASK-009](tasks/DJC-TASK-009.md) | Timestamps propres à la source (RootEncoder) | low | 5 | 20k | — |

## review

| ID | Titre | Priorité | Points | Estimation | Dépend de |
| --- | --- | --- | --- | --- | --- |
| [DJC-BUG-009](tasks/DJC-BUG-009.md) | Latence ~4 s DJI → OBS (-r en entrée, preset) | high | 2 | 10k | — |
| [DJC-BUG-007](tasks/DJC-BUG-007.md) | Relais Android : getSurfaceTexture null au Démarrer | high | 3 | 12k | — |
| [DJC-BUG-008](tasks/DJC-BUG-008.md) | OBS figé / saccadé : timestamps cassés du relais | high | 3 | 12k | TASK-008 |
| [DJC-TASK-008](tasks/DJC-TASK-008.md) | Relais : reconnexion auto + adaptation orientation | high | 3 | 14k | BUG-007 |
| [DJC-TASK-010](tasks/DJC-TASK-010.md) | Latence résiduelle de 1 à 2 s vers OBS | high | 2 | 8k | BUG-008 |
| [DJC-TASK-011](tasks/DJC-TASK-011.md) | Relais : aperçu caméra plein écran | high | 2 | 8k | TASK-008 |
| [DJC-TASK-012](tasks/DJC-TASK-012.md) | Forçage DJI : URL RTMP + encodeur à distance | high | 5 | 18k | — |
| [DJC-TASK-013](tasks/DJC-TASK-013.md) | Livraison .deb 0.1.1, paquet macOS et icône | high | 3 | 15k | TASK-011, TASK-012 |

## blocked

_(vide)_

## done

| ID | Titre | Terminé le | Tokens |
| --- | --- | --- | --- |
| [DJC-TASK-001](tasks/DJC-TASK-001.md) | Relais Android de test (Capacitor + RootEncoder) | 2026-11-17 | 34k |
| [DJC-SPIKE-002](tasks/DJC-SPIKE-002.md) | Faisabilité passerelle Android RTMP vs WHIP | 2026-11-17 | 7.5k |
| [DJC-TASK-003](tasks/DJC-TASK-003.md) | APK du relais sur la landing + doc | 2026-11-17 | 6k |
| [DJC-TASK-014](tasks/DJC-TASK-014.md) | Sorties HLS basse latence et WebRTC | 2026-11-19 | 24k |
| [DJC-TASK-015](tasks/DJC-TASK-015.md) | Son de la caméra en WebRTC (relais Opus) | 2026-11-19 | 11k |
| [DJC-TASK-016](tasks/DJC-TASK-016.md) | WebRTC seule sortie, sans Bluetooth, audio Opus direct | 2026-11-20 | 26k |

## Dettes hors backlog

- Aucun test automatisé sur le frontend (couvert en partie par `tsc`).
- Pas de `.gitignore` pour `mobile/android/` : à trancher (projet natif généré
  mais patché à la main, donc versionné).

## Historique

| Date | Événement |
| --- | --- |
| 2026-11-17 | Création du backlog, v0.1.0 livrée, relais Android compilé. |
| 2026-11-19 | HLS basse latence + WebRTC activés dans MediaMTX, liens dans Diffusion. |
| 2026-11-19 | Audio WebRTC : relais Opus sur `live/dji-webrtc`, AAC conservé pour OBS/HLS. |
| 2026-11-20 | v0.2.0 : Bluetooth et sorties RTMP/HLS retirés, WebRTC seule sortie, audio Opus encodé avec la vidéo. |
| 2026-11-18 | v0.1.1 bureau + APK 0.1.3 (plein écran, rotation), encodeur DJI à distance, icône unifiée, CI macOS. |

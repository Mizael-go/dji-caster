# Backlog — DJI Caster

Mis à jour : 2026-11-17 · prochaine revue : à la prochaine session
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

## review

_(vide)_

## blocked

_(vide)_

## done

| ID | Titre | Terminé le | Tokens |
| --- | --- | --- | --- |
| [DJC-TASK-001](tasks/DJC-TASK-001.md) | Relais Android de test (Capacitor + RootEncoder) | 2026-11-17 | 34k |
| [DJC-SPIKE-002](tasks/DJC-SPIKE-002.md) | Faisabilité passerelle Android RTMP vs WHIP | 2026-11-17 | 7.5k |
| [DJC-TASK-003](tasks/DJC-TASK-003.md) | APK du relais sur la landing + doc | 2026-11-17 | 6k |

## Dettes hors backlog

- Aucun test automatisé sur le frontend (couvert en partie par `tsc`).
- Pas de `.gitignore` pour `mobile/android/` : à trancher (projet natif généré
  mais patché à la main, donc versionné).

## Historique

| Date | Événement |
| --- | --- |
| 2026-11-17 | Création du backlog, v0.1.0 livrée, relais Android compilé. |

---
id: "DJC-TASK-003"
type: "TASK"
status: "done"
priority: "medium"
labels: ["landing", "mobile", "documentation"]
components: ["landing", "mobile"]
story_points: 2
epic: null
sprint: null
parent: null
children: []
depends_on: ["DJC-TASK-001"]
blocks: []
related_to: []
tokens_estimated: 6000
tokens_used: 6000
created: "2026-11-17"
updated: "2026-11-17"
---
## Summary

Rendre le relais Android téléchargeable depuis la landing et documenter son
usage.

## Description

Le script de build copie déjà l'APK dans `landing/downloads/dji-relay.apk`.
Il reste à proposer le lien à côté du `.deb` et à expliquer la procédure de
test : IP du PC, réseau commun, chemin `live/dji`, conversion à désactiver.

## Acceptance Criteria

- [x] `landing/index.html` propose « APK de test (Android) » vers
      `./downloads/dji-relay.apk`.
- [x] La procédure de test est dans `mobile/README.md` et rappelle que le
      téléphone émet déjà du H.264.
- [x] `mobile/README.md` décrit `npm install`, `npm run android:sync`,
      `./.apk-build.sh`.

## Notes

L'APK est signé avec la clé de debug de Gradle : installable partout, non
publiable sur le Play Store, et une réinstallation après régénération de la clé
impose une désinstallation.

---
id: "DJC-TASK-011"
type: "TASK"
status: "review"
priority: "high"
labels: ["mobile", "android", "ui", "rtmp"]
components: ["mobile"]
story_points: 2
epic: null
sprint: null
parent: null
children: []
depends_on: ["DJC-TASK-008"]
blocks: []
related_to: ["DJC-TASK-008", "DJC-BUG-007"]
tokens_estimated: 8k
tokens_used: 7k
created: "2026-11-18"
updated: "2026-11-18"
---
## Summary

Au démarrage de la retransmission, le téléphone n'affiche qu'un aperçu de
42 % de la hauteur : l'écran ne sert pas la visée. L'aperçu caméra doit
occuper tout l'écran.

## Description

Le relais Android montre la caméra dans une bande en haut de la WebView.
Pour filmer correctement, l'utilisateur veut que l'app passe en plein écran
caméra dès que la diffusion démarre, tout en pouvant l'arrêter sans
repasser par l'interface web.

## Fix (`RtmpRelayPlugin.java`)

- L'aperçu `OpenGlView` est ajouté en `MATCH_PARENT`/`MATCH_PARENT`
  (`Gravity.FILL`) dans la vue contenu, au-dessus de la WebView.
- Un bouton natif « Arrêter » (pill `#CC0A0B0D`, coins arrondis, texte
  blanc) est superposé en bas au centre. Il coupe la diffusion, libère les
  vues et émet l'état `idle` : sinon la WebView, recouverte, serait
  inaccessible.
- `release()` retire l'aperçu **et** le bouton, y compris lors d'une
  rotation ou d'une reconnexion (aucune vue orpheline).
- Helper `dp()` pour les marges/paddings indépendants de la densité.

## Acceptance Criteria

- [x] L'aperçu caméra occupe tout l'écran pendant la diffusion.
- [x] Un bouton « Arrêter » natif reste accessible et arrête la diffusion.
- [x] Rotation et reconnexion ne laissent aucune vue résiduelle.
- [x] APK 0.1.3 produit (`./.apk-build.sh`).
- [ ] Validation sur téléphone réel (bouton Arrêter, retour à l'UI).

## Notes

Vérification : `.apk-build.sh` → `APK : 9.4M — sha256 f79c5130…`.

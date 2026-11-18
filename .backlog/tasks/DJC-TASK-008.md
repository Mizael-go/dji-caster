---
id: "DJC-TASK-008"
type: "TASK"
status: "review"
priority: "high"
labels: ["mobile", "android", "rtmp", "fiabilite"]
components: ["mobile"]
story_points: 3
epic: null
sprint: null
parent: null
children: []
depends_on: ["DJC-BUG-007"]
blocks: []
related_to: ["DJC-BUG-007", "DJC-BUG-008"]
tokens_estimated: 14k
tokens_used: 16k
created: "2026-11-17"
updated: "2026-11-17"
---
## Summary

Le relais Android ne survit pas aux déconnexions RTMP et ne s'adapte pas à
la rotation de l'écran.

## Description

Deux défauts constatés en test réel :

1. Quand MediaMTX ferme la connexion (silence > readTimeout 15 s, Wi-Fi
   instable), RootEncoder lève « Error send packet, Broken pipe » en boucle.
   L'app reste dans un état fantôme : aucune reconnexion, l'aperçu tourne
   mais plus rien n'est publié. Il faut arrêter/relancer à la main.
2. En paysage, le flux reste encodé dans les dimensions portrait : l'image
   est écrasée ou reste verticale, la scène ne suit pas l'orientation.

## Fix (`RtmpRelayPlugin.java`)

1. Reconnexion automatique : sur `onConnectionFailed`/`onDisconnect`, l'état
   « reconnecting » est émis puis `startStream` est retenté avec backoff
   (2 s, 4 s, 8 s, 16 s, 30 s). Au bout de 5 tentatives, abandon propre
   (aperçu fermé, état « error »). Nouvel état `reconnecting` géré par
   `mobile/www/app.js` (point orange `dot.warn`).
2. Rotation : `handleOnConfigurationChanged` redémarre la diffusion
   (`internalStart`) après `ROTATION_SETTLE_MS` (400 ms), le temps qu'Android
   applique la nouvelle orientation lue par `getCameraOrientation`. Le
   manifeste déclare déjà `configChanges` (l'activité n'est pas recréée) ;
   RootEncoder permute largeur/hauteur (90/270°) → le flux s'adapte au
   paysage.
3. Les échecs de redémarrage sans `PluginCall` (cas rotation) sont émis
   comme erreurs au lieu d'être avalés.

## Acceptance Criteria

- [x] Une coupure serveur (> 15 s) déclenche une reconnexion automatique
      sans intervention utilisateur.
- [x] L'UI reflète l'état « reconnexion » (statut + pastille).
- [x] Le flux suit l'orientation : portrait 720x1280, paysage 1280x720.
- [x] APK 0.1.2 rebuild (`./.apk-build.sh`) et copié dans
      `landing/downloads/dji-relay.apk`.
- [ ] Validation manuelle sur téléphone réel : coupure Wi-Fi + rotation
      pendant la diffusion.

## Notes

Version bump 0.1.2. Piège `npx cap sync` : artefacts Gradle root dans
`capacitor-cordova-android-plugins/` — à supprimer via Docker avant sync
(voir DJC-BUG-007).

Rebuild 0.1.3 (avec DJC-TASK-011) : rotation différée de 400 ms pour que
la nouvelle orientation soit effective avant de relancer l'encodeur.

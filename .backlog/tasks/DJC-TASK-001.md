---
id: "DJC-TASK-001"
type: "TASK"
status: "done"
priority: "high"
labels: ["mobile", "test", "rtmp"]
components: ["mobile"]
story_points: 5
epic: null
sprint: null
parent: null
children: []
depends_on: ["DJC-SPIKE-002"]
blocks: ["DJC-TASK-003"]
related_to: []
tokens_estimated: 30000
tokens_used: 34000
created: "2026-11-17"
updated: "2026-11-17"
---
## Summary

Application Android « DJI Relay » qui publie la caméra du téléphone en RTMP sur
le MediaMTX de DJI Caster, pour tester le logiciel sans caméra DJI.

## Description

Le poste de travail ne reçoit que du RTMP sur `live/dji` : si un téléphone y
publie un flux H.264/AAC au bon chemin, le logiciel se comporte exactement
comme avec une caméra DJI (même watcher de chemins, même bannière de session,
même diffusion vers OBS). L'application est construite avec la même méthode que
`ticket-park` : Capacitor (web dans une WebView) + plugin natif Java, APK compilé
par Gradle 8.11 dans l'image Docker JDK 21 / SDK 35.

- UI web : `mobile/www` (saisie IP/port/chemin, résolution, débit, caméra).
- Plugin natif : `RtmpRelayPlugin` (RootEncoder 2.8.2) — `prepareVideo`,
  `prepareAudio`, `startPreview`, `startStream`.
- Aperçu : `OpenGlView` natif ajouté par-dessus la WebView dans la bande haute
  (42 % de la hauteur d'écran).
- Permissions : `CAMERA`, `RECORD_AUDIO`, `INTERNET` (demandées à l'exécution).

## Acceptance Criteria

- [x] `npx cap sync android` copie le web dans le projet natif.
- [x] `./.apk-build.sh` produit `android/app/build/outputs/apk/debug/app-debug.apk`.
- [x] L'APK est copié dans `landing/downloads/dji-relay.apk`.
- [ ] Diffusion vérifiée sur un téléphone réel (à faire avec l'appareil en main).

## Notes

Build reproductible : image `dji-caster-android:latest` (couches JDK 21, SDK 35,
Gradle 8.11.1 réutilisées depuis `ticket-park-android`). `cap sync` tourne sur
l'hôte, seul Gradle tourne dans le conteneur.

Pièges rencontrés :

1. `:capacitor-android:extractDebugAnnotations` se bloque après
   `:app:packageDebug` — l'APK est complet mais Gradle n'exit jamais. Le script
   surveille le fichier, arrête le conteneur et vérifie la fraîcheur de l'APK.
   Exclure la tâche (`-x`) casse `syncDebugLibJars` : ne pas le faire.
2. Sans `buildToolsVersion "35.0.0"` forcé sur tous les modules, Gradle
   installe build-tools 34 à chaque build (conteneur jetable).

APK vérifié : `mg.djicaster.relay`, minSdk 23, target 35, permissions
INTERNET / CAMERA / RECORD_AUDIO, signature debug valide, 9,4 Mo,
sha256 `eeec8b25…`, copié dans `landing/downloads/dji-relay.apk`.
Dex contrôlé : `com/getcapacitor/Bridge`, `RtmpCamera2`, `RtmpRelayPlugin`.

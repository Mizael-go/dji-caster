---
id: "DJC-BUG-007"
type: "BUG"
status: "review"
priority: "high"
labels: ["mobile", "android", "rtmp"]
components: ["mobile"]
story_points: 3
epic: null
sprint: null
parent: null
children: []
depends_on: []
blocks: []
related_to: ["DJC-TASK-001"]
tokens_estimated: 12000
tokens_used: 16000
created: "2026-11-17"
updated: "2026-11-17"
---
## Summary

Le relais Android échoue au « Démarrer » avec
« Diffusion impossible : getSurfaceTexture(...) must not be null ».

## Description

Deux défauts dans `RtmpRelayPlugin.startRelay` :

1. `camera.startPreview(...)` / `camera.startStream(url)` sont appelés de
   façon synchrone juste après `content.addView(preview)`. À cet instant la
   `Surface` de l'`OpenGlView` n'existe pas encore (elle est créée après le
   premier layout/dessin). L'init GL lancée par `glInterface.start()` échoue
   sur `holder.surface` null et l'exception est avalée par `secureSubmit` :
   `mainRender.surfaceTexture` reste null et le null-check Kotlin
   (`OpenGlView.getSurfaceTexture()`) lève le NPE affiché dans l'UI.
2. `prepareVideo` est appelé avec les arguments dans le mauvais ordre :
   `prepareVideo(width, height, bitrate, fps, 90, metrics.densityDpi)` est
   interprété comme `(width, height, fps, bitrate, iFrameInterval, rotation)`
   → fps = débit (2 500 000), débit = 30 bps, iFrameInterval = 90,
   rotation = DPI (~420). Même sans le bug 1, la configuration vidéo était
   invalide. Signature attendue (RootEncoder 2.8.2, `Camera2Base`) :
   `prepareVideo(width, height, fps, bitrate, rotation)` avec
   `rotation = CameraHelper.getCameraOrientation(context)`.

## Acceptance Criteria

- [x] `startRelay` attend que la surface GL soit valide (avec timeout)
      avant d'appeler `startPreview`/`startStream`.
- [x] `prepareVideo` appelé dans l'ordre correct avec la rotation capteur.
- [x] `mobile/package.json` passe en 0.1.1.
- [x] APK rebuild (`./.apk-build.sh`) et copié dans
      `landing/downloads/dji-relay.apk`.
- [x] Landing redéployée sur Vercel (`vercel --prod`).
- [x] Diffusion vérifiée sur un téléphone réel contre DJI Caster (manuel,
      fait avec l'appareil en main : le relais se connecte et publie ; les
      défauts restants — saccadé, figeage OBS, reconnexion — sont tracés
      dans DJC-BUG-008 et DJC-TASK-008).

## Notes

Diagnostic fait en lisant les sources tag 2.8.2 de RootEncoder :
`OpenGlView.start()` (init GL asynchrone, erreur avalée par `secureSubmit`),
`Camera2Base.prepareGlView` (`cameraManager.prepareCamera(getSurfaceTexture())`),
`MainRender.getSurfaceTexture` / `CameraRender.surfaceTexture` (null avant
`initGl`).

## Preuves

- `gradle assembleDebug` → `BUILD SUCCESSFUL in 25s` (journal
  `/tmp/dji-relay-apk.log`).
- APK : 9,4 Mo, sha256 `8771b0ab…ebd834e7`, copié dans
  `landing/downloads/dji-relay.apk`.
- Le dex contient `waitForSurface` (nouvelle logique d'attente de surface),
  contrôlé par extraction de `classes.dex`.
- Déploiement Vercel : `dji-caster-bjhc60c4k…` (Ready, Production), alias
  `https://dji-caster.vercel.app` ; l'APK téléchargé à l'URL publique est
  identique (sha256 égal).

Piège : `npx cap sync android` échoue si `capacitor-cordova-android-plugins/`
contient des artefacts Gradle appartenant à root (Docker) — le supprimer via
`docker run … rm -rf` avant `cap sync`.

# DJI Relay — relais Android de test

Application Android qui publie la caméra du téléphone en **RTMP H.264/AAC** sur
le MediaMTX de DJI Caster. Du point de vue du poste de travail, le flux est
identique à celui d'une caméra DJI : même protocole, même chemin (`live/dji`),
même traitement aval (watcher, bannière de session, OBS).

Elle sert à tester le logiciel sans matériel DJI.

## Comment ça marche

- **Enveloppe** : Capacitor 7 (même méthode que `ticket-park`) — la page
  `www/` est embarquée dans une WebView, l'APK ne dépend d'aucun serveur.
- **Plugin natif** : `RtmpRelayPlugin` (`android/app/src/main/java/mg/djicaster/relay`)
  s'appuie sur RootEncoder 2.8.2 : `MediaCodec` encode le H.264/AAC et le client
  RTMP publie sur l'URL saisie.
- **Aperçu** : un `OpenGlView` natif est ajouté par-dessus la WebView dans la
  bande haute (42 % de la hauteur) que la page laisse libre.
- **Permissions** : `CAMERA` et `RECORD_AUDIO` demandées à l'exécution.

## Utilisation

1. Sur le PC : lancer DJI Caster, démarrer le serveur MediaMTX (port 1935).
2. Sur le téléphone : même Wi-Fi que le PC, ouvrir DJI Relay, saisir l'IP du PC
   et garder le chemin source `live/dji`.
3. « Démarrer » → le chemin `live/dji` passe à *prêt* dans l'onglet Diffusion.
4. Dans DJI Caster, l'URL de source affichée est celle à saisir sur le téléphone.
   `live/dji-h264` est une sortie réservée à FFmpeg lorsque la conversion est
   activée ; ne configurez pas le relais mobile sur cette sortie.
5. Pour le relais Android, activez « Convertir en H.264 » dans DJI Caster afin
   de régénérer les timestamps du flux, puis utilisez dans OBS l'URL affichée
   par DJI Caster, qui se termine par `live/dji-h264`. La lecture directe de
   `live/dji` peut retarder ou figer l'image avec les timestamps RootEncoder.

## Build

```bash
npm install
npm run android:sync     # copie www/ dans le projet natif
./.apk-build.sh          # APK via l'image Docker (JDK 21 + SDK 35 + Gradle 8.11)
```

L'APK sort sur `android/app/build/outputs/apk/debug/app-debug.apk` et une copie
est déposée dans `../landing/downloads/dji-relay.apk`.

L'image est construite une seule fois :

```bash
docker build -f .apk-build.Dockerfile -t dji-caster-android:latest .
```

`cap sync` tourne sur l'hôte (Node), seul Gradle tourne dans le conteneur —
comme dans `ticket-park`. L'APK est signé avec la clé de debug : installable
partout, non publiable sur le Play Store.

## Limites

- Le relais ne simule pas le Bluetooth : le scan et le forçage Wi-Fi d'une
  vraie DJI restent à tester séparément.
- Une seule source par chemin : pour deux téléphones, voir DJC-BUG-006.
- Pas de HTTPS ni d'authentification RTMP : réseau de confiance uniquement.

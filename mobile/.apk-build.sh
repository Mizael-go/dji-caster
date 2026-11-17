#!/usr/bin/env bash
# Construit l'APK du relais Android dans le conteneur Docker (JDK 21 + SDK 35 + Gradle 8.11).
# Même méthode que ticket-park.
#
# Usage :
#   1. image (une seule fois) : docker build -f .apk-build.Dockerfile -t dji-caster-android:latest .
#   2. dépendances + projet natif : npm install && npm run android:sync
#   3. APK : ./.apk-build.sh
#
# L'APK sort sur : android/app/build/outputs/apk/debug/app-debug.apk
#
# Particularité : dans ce conteneur, Gradle n'a jamais rendu la main après
# :app:packageDebug (tâche :capacitor-android:extractDebugAnnotations bloquée).
# L'APK, lui, est complet à ce moment-là. Le script surveille donc le fichier,
# attend que sa taille soit stable, puis arrête le conteneur.
set -uo pipefail
cd "$(dirname "$0")"

IMAGE=dji-caster-android:latest
APK=android/app/build/outputs/apk/debug/app-debug.apk
DEST=../landing/downloads/dji-relay.apk
LOG=${TMPDIR:-/tmp}/dji-relay-apk.log
NAME=dji-caster-apk-$$

if ! docker image inspect "$IMAGE" > /dev/null 2>&1; then
  docker build -f .apk-build.Dockerfile -t "$IMAGE" . || exit 1
fi

START=$(date +%s)
# --console=plain : sans terminal, la sortie riche de Gradle peut rester bloquée
# sur une redirection vers un fichier.
docker run --rm --name "$NAME" \
  -v "$PWD:/app" \
  -v dji-caster-gradle-cache:/root/.gradle \
  -w /app/android \
  "$IMAGE" \
  bash -c 'gradle assembleDebug --no-daemon --console=plain' > "$LOG" 2>&1 &
CLIENT=$!

FRESH=0
for _ in $(seq 1 200); do
  if [ -f "$APK" ] && [ "$(stat -c %Y "$APK")" -ge "$START" ]; then
    SIZE1=$(stat -c %s "$APK")
    sleep 3
    if [ "$SIZE1" = "$(stat -c %s "$APK")" ]; then
      FRESH=1
      break
    fi
  fi
  kill -0 "$CLIENT" 2>/dev/null || break
  sleep 5
done

if [ "$FRESH" -eq 1 ]; then
  docker rm -f "$NAME" > /dev/null 2>&1
  wait "$CLIENT" 2>/dev/null
else
  wait "$CLIENT" 2>/dev/null
  if ! grep -q "BUILD SUCCESSFUL" "$LOG"; then
    echo "Échec du build, journal : $LOG" >&2
    tail -20 "$LOG" >&2
    exit 1
  fi
fi

mkdir -p ../landing/downloads
cp "$APK" "$DEST"
echo "APK : $(du -h "$APK" | cut -f1) — sha256 $(sha256sum "$APK" | cut -d' ' -f1)"
echo "Copie : mobile/$DEST (journal : $LOG)"

#!/usr/bin/env bash
# Construit le binaire Wails puis l'empaquette en .deb installable.
#
#   ./scripts/build-deb.sh            # version 0.2.0
#   VERSION=0.3.0 ./scripts/build-deb.sh
#
# Produit dist/dji-caster_<version>_<arch>.deb et en copie une dans
# landing/downloads/ pour le site.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

VERSION="${VERSION:-0.2.0}"
PKG="dji-caster"
ARCH="$(dpkg --print-architecture 2>/dev/null || echo amd64)"
STAGE="$ROOT/dist/deb"
OUT="$ROOT/dist/${PKG}_${VERSION}_${ARCH}.deb"

command -v dpkg-deb >/dev/null || { echo "dpkg-deb est requis (paquet dpkg-dev)"; exit 1; }
command -v wails >/dev/null || { echo "wails est requis (go install github.com/wailsapp/wails/v2/cmd/wails@latest)"; exit 1; }

echo "▸ Compilation du binaire (tags webkit2_41)"
wails build -tags webkit2_41

echo "▸ Préparation de l'arborescence"
rm -rf "$STAGE"
mkdir -p "$STAGE/DEBIAN" \
         "$STAGE/usr/bin" \
         "$STAGE/usr/share/applications" \
         "$STAGE/usr/share/icons/hicolor/256x256/apps" \
         "$STAGE/usr/share/doc/${PKG}"

install -Dm755 build/bin/dji-caster "$STAGE/usr/bin/${PKG}"
install -Dm644 build/appicon.png \
  "$STAGE/usr/share/icons/hicolor/256x256/apps/${PKG}.png"
install -Dm644 README.md "$STAGE/usr/share/doc/${PKG}/README.md"

cat > "$STAGE/usr/share/applications/${PKG}.desktop" <<'DESKTOP'
[Desktop Entry]
Type=Application
Name=DJI Caster
GenericName=RTMP to WebRTC bridge for DJI cameras
Comment=Reçoit le flux RTMP d'une caméra DJI et le publie en WebRTC
Exec=dji-caster
Icon=dji-caster
Terminal=false
Categories=AudioVideo;Video;Network;Utility;
Keywords=dji;rtmp;webrtc;stream;
StartupNotify=true
DESKTOP

INSTALLED_KB="$(du -sk "$STAGE/usr" | cut -f1)"

cat > "$STAGE/DEBIAN/control" <<CONTROL
Package: ${PKG}
Version: ${VERSION}
Section: video
Priority: optional
Architecture: ${ARCH}
Installed-Size: ${INSTALLED_KB}
Depends: libgtk-3-0t64 | libgtk-3-0, libwebkit2gtk-4.1-0
Recommends: ffmpeg
Maintainer: Mizael-go <pagielamizael@gmail.com>
Homepage: https://github.com/Mizael-go/dji-caster
Description: Passerelle RTMP vers WebRTC pour caméras DJI
 Application desktop qui reçoit le flux RTMP d'une caméra DJI (ou du
 relais Android) et le publie en WebRTC, la sortie de lecture à la plus
 basse latence. MediaMTX est téléchargé dans le dossier de l'utilisateur
 au premier lancement ; ffmpeg encode la vidéo H.264 et l'audio Opus.
CONTROL

echo "▸ Construction du paquet"
mkdir -p "$ROOT/dist"
dpkg-deb --build --root-owner-group "$STAGE" "$OUT" >/dev/null
rm -rf "$STAGE"

mkdir -p "$ROOT/landing/downloads"
cp "$OUT" "$ROOT/landing/downloads/"

echo "✔ $(basename "$OUT") ($(du -h "$OUT" | cut -f1))"
echo "  installation : sudo apt install $OUT"
echo "  copie site   : landing/downloads/$(basename "$OUT")"

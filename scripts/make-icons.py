#!/usr/bin/env python3
"""Génère les icônes de l'app desktop à partir du logo du relais Android.

Le logo source vit dans l'APK (`mobile/android/app/src/main/res/mipmap-*`).
On le reconstruit à haute résolution (plaque blanche arrondie + marque bleue)
pour éviter un agrandissement flou du PNG 192 px.

Produit :
  build/appicon.png          icône Wails / .deb (1024x1024)
  build/windows/icon.ico     icône Windows multi-tailles

Usage : python3 scripts/make-icons.py
"""

from pathlib import Path

from PIL import Image, ImageDraw

ROOT = Path(__file__).resolve().parent.parent
SRC_ICON = ROOT / "mobile/android/app/src/main/res/mipmap-xxxhdpi/ic_launcher.png"
SRC_FG = ROOT / "mobile/android/app/src/main/res/mipmap-xxxhdpi/ic_launcher_foreground.png"
OUT_PNG = ROOT / "build/appicon.png"
OUT_ICO = ROOT / "build/windows/icon.ico"

# Proportions mesurées sur l'icône Android 192 px : plaque 152/192 (rayon
# 22/152), marque 84/192 centrée, bleu #6AC0FD.
PLATE_RATIO = 152 / 192
RADIUS_RATIO = 22 / 152
MARK_RATIO = 84 / 192
MARK_COLOR = (106, 192, 253, 255)

SS = 4  # supersampling : on dessine en 4096 puis on réduit à 1024


def build_icon(size: int) -> Image.Image:
    canvas = size * SS
    img = Image.new("RGBA", (canvas, canvas), (0, 0, 0, 0))
    draw = ImageDraw.Draw(img)

    plate = round(canvas * PLATE_RATIO)
    offset = (canvas - plate) // 2
    radius = round(plate * RADIUS_RATIO)
    draw.rounded_rectangle(
        (offset, offset, offset + plate, offset + plate),
        radius=radius,
        fill=(255, 255, 255, 255),
    )

    # Masque de la marque : dans le foreground Android, l'alpha représente la
    # plaque blanche ; la marque est la zone bleue. On la découpe par couleur.
    import numpy as np

    arr = np.array(Image.open(SRC_FG).convert("RGBA")).astype(int)
    blue_mask = (arr[:, :, 2] - arr[:, :, 0] > 40) & (arr[:, :, 3] > 128)
    mask = Image.fromarray((blue_mask * 255).astype("uint8"))
    box = mask.getbbox()
    mark = mask.crop(box)
    target = round(canvas * MARK_RATIO)
    mark = mark.resize((target, target), Image.LANCZOS).point(
        lambda v: 255 if v > 128 else 0
    )
    mark_img = Image.new("RGBA", mark.size, MARK_COLOR)
    mark_img.putalpha(mark)
    pos = ((canvas - target) // 2, (canvas - target) // 2)
    img.alpha_composite(mark_img, pos)

    return img.resize((size, size), Image.LANCZOS)


def main() -> None:
    icon = build_icon(1024)
    OUT_PNG.parent.mkdir(parents=True, exist_ok=True)
    icon.save(OUT_PNG)
    OUT_ICO.parent.mkdir(parents=True, exist_ok=True)
    icon.save(
        OUT_ICO,
        sizes=[(16, 16), (24, 24), (32, 32), (48, 48), (64, 64), (128, 128), (256, 256)],
    )
    print(f"✔ {OUT_PNG.relative_to(ROOT)} (1024x1024)")
    print(f"✔ {OUT_ICO.relative_to(ROOT)} (multi-tailles)")


if __name__ == "__main__":
    main()

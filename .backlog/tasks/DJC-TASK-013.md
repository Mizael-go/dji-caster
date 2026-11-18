---
id: "DJC-TASK-013"
type: "TASK"
status: "review"
priority: "high"
labels: ["distribution", "packaging", "macos", "icone", "landing"]
components: ["landing", "build", "scripts", ".github"]
story_points: 3
epic: null
sprint: null
parent: null
children: []
depends_on: ["DJC-TASK-011", "DJC-TASK-012"]
blocks: []
related_to: ["DJC-TASK-003", "DJC-TASK-010"]
tokens_estimated: 15k
tokens_used: 13k
created: "2026-11-18"
updated: "2026-11-18"
---
## Summary

Les paquets exposés sur la landing étaient figés en 0.1.0 (Linux) sans
version macOS, et l'icône du logiciel ne correspondait pas au logo de
l'app mobile.

## Description

Livrer la version 0.1.1 du bureau avec : un `.deb` à jour, une archive
macOS universelle, et l'icône du relais Android comme icône de
l'application desktop (Linux, Windows, macOS).

## Fix

- `scripts/make-icons.py` : reconstruit `build/appicon.png` (1024²) et
  `build/windows/icon.ico` depuis
  `mobile/android/app/src/main/res/mipmap-xxxhdpi/ic_launcher.png` —
  plaque blanche arrondie (152/192, rayon 22/152), marque bleue
  `#6AC0FD` (84/192), supersampling ×4 pour éviter l'agrandissement flou.
- `scripts/build-deb.sh` : `VERSION=0.1.1`, le paquet est copié dans
  `landing/downloads/` et l'ancien 0.1.0 est supprimé.
- `.github/workflows/macos.yml` : Wails ne cross-compile pas depuis Linux.
  Le workflow construit `darwin/universal` sur `macos-latest`, zippe
  `dji-caster.app` et le publie comme asset de la release
  `macos-latest` (téléchargeable sans authentification, dépôt public).
- Landing : liens `.deb` en 0.1.1, bouton macOS, note d'installation
  (app non signée → clic droit → Ouvrir), `LISEZ-MOI.txt` réécrit.
- APK 0.1.3 (plein écran + rotation) remplace l'ancien sur la landing.

## Acceptance Criteria

- [x] Icône desktop identique au logo mobile (`build/appicon.png`,
      `build/windows/icon.ico`).
- [x] Binaire desktop reconstruit et `.deb` 0.1.1 produit.
- [x] Landing pointée sur `dji-caster_0.1.1_amd64.deb` et APK 0.1.3.
- [x] Workflow macOS présent et déclenché au push sur `main`.
- [ ] Archive `dji-caster-macos-universal.zip` publiée par la CI puis
      copiée dans `landing/downloads/`.
- [ ] Déploiement Vercel `--prod` depuis la racine du dépôt, les trois
      fichiers servis et vérifiés (sha256 de l'APK identique).

## Notes

Vérification locale : `wails build -tags webkit2_41`,
`VERSION=0.1.1 ./scripts/build-deb.sh`, `python3 scripts/make-icons.py`,
`./.apk-build.sh` (sha256 `f79c5130…`). Le paquet macOS dépend du run
GitHub Actions (pas de Mac disponible ici).

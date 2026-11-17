---
id: "DJC-SPIKE-002"
type: "SPIKE"
status: "done"
priority: "high"
labels: ["mobile", "faisabilite", "rtmp", "webrtc"]
components: ["mobile", "backend"]
story_points: 2
epic: null
sprint: null
parent: null
children: []
depends_on: []
blocks: ["DJC-TASK-001"]
related_to: []
tokens_estimated: 8000
tokens_used: 7500
created: "2026-11-17"
updated: "2026-11-17"
---
## Summary

Une passerelle Android peut-elle produire un flux que DJI Caster reçoit comme
celui d'une caméra DJI ?

## Description

Deux familles de solutions ont été comparées.

| Méthode | Ce qui est envoyé | Verdict |
| --- | --- | --- |
| **RTMP natif** (RootEncoder 2.8.2 via JitPack) | H.264 + AAC dans un RTMP sur `:1935/live/dji` | **Retenu** : c'est le protocole et le chemin exacts d'une DJI, aucun changement côté desktop |
| **WHIP / WebRTC** (support natif de MediaMTX sur `:8889`) | WebRTC depuis la WebView | Écarté comme méthode principale : la WebView encode en VP8 par défaut, et MediaMTX ne remuxe pas VP8 en RTMP pour OBS |

Contraintes vérifiées :

- Android 7+ (`minSdk 23` chez Capacitor) suffit : RootEncoder utilise
  `MediaCodec` (encodage matériel) et l'API Camera2 (21+).
- La WebView Capacitor sert la page depuis `https://localhost` : contexte sûr,
  donc aucune contrainte HTTPS supplémentaire pour la caméra.
- Le plugin doit demander `CAMERA` et `RECORD_AUDIO` à l'exécution ; un refus
  laisse l'app utilisable mais sans flux.
- L'aperçu impose une `SurfaceView` (`OpenGlView`) : elle est ajoutée
  par-dessus la WebView, la page réservant 42 % de la hauteur.

## Acceptance Criteria

- [x] Méthode retenue compatible avec le pipeline existant (aucun changement Go).
- [x] Deux méthodes documentées, avec la raison de l'abandon du WHIP.
- [x] Coût de build identifié : image Docker existante, Gradle 8.11, JitPack.

## Notes

Limite connue : le téléphone émet du H.264, la caméra DJI peut émettre du H.265.
La conversion ffmpeg reste inutile pour ce relais (à décocher côté desktop).
Le relais ne simule pas le BLE : le scan et le forçage Wi-Fi restent à tester
séparément.

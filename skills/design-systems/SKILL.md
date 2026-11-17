---
name: design-systems
description: >-
  Design system de DJI Caster : tokens de couleur, typographie et règles de
  composition pour l'UI React + shadcn. À utiliser pour toute modification
  visuelle. Références : https://github.com/klaufel/awesome-design-systems
  et https://ui.shadcn.com
---

# Design system — console de diffusion

## Principe

Fenêtre étroite (≈420px) : une seule colonne, aucune décoration gratuite.
L'information critique (état du flux, URL RTMP) doit être lisible en une
glance, sans scroll.

## Tokens (source : `frontend/src/style.css`)

| Token                    | Valeur      | Usage                              |
| ------------------------ | ----------- | ---------------------------------- |
| `--background`           | `#0a0b0d`   | fond                               |
| `--card`                 | `#111318`   | surfaces                           |
| `--foreground`           | `#edeff3`   | texte principal                    |
| `--muted-foreground`     | `#8b919d`   | texte secondaire, mono des adresses |
| `--primary`              | `#c8f751`   | actions, « signal », flux actif    |
| `--color-live`           | `#ff5f57`   | indicateur EN DIRECT               |
| `--color-warning`        | `#fbbf24`   | attente, HEVC                      |
| `--destructive`          | `#ff6259`   | erreurs                            |

Ne jamais coder une couleur en dur dans un composant : utiliser les classes
générées (`bg-primary`, `text-muted-foreground`, `border-border`…) ou
`bg-[color:var(--color-live)]` pour les accents hors thème shadcn.

## Règles

- `font-mono` pour tout ce qui est technique : adresses MAC, URLs, ports, logs.
- `tabular-nums` pour les compteurs (octets, lecteurs).
- Rayons : `rounded-lg` pour les blocs interactifs, `rounded-xl` pour les cartes.
- États d'un flux : point coloré + libellé, jamais une couleur seule.
- Un seul bouton primaire par écran (l'action attendue).
- Pas d'ombres : séparation par `border-border` et `bg-muted/40`.
- Les composants shadcn viennent de `frontend/src/components/ui` (style
  `radix-nova`, import du paquet unifié `radix-ui`). Ajouter un composant =
  copier son source, pas installer un wrapper.

## Vérification

Une modification visuelle n'est validée que si `npm run build` passe ; le rendu
lui-même doit être regardé dans `wails dev` (ou signalé comme non vérifié).

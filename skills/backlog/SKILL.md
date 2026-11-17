---
name: backlog
description: >-
  Gestion du backlog et du budget de tokens de DJI Caster. À utiliser dès que
  .backlog/ existe ou qu'il est question de tâches, tickets, dettes techniques
  ou sprints. Source : https://github.com/hantsaniala/backlog-skills
---

# Backlog & budget de tokens

## Quand l'utiliser

- Un dossier `.backlog/` existe à la racine (le créer initialise le système).
- L'utilisateur parle de tâches, tickets, bugs, sprints, dette technique.

## Table de décision

| Objectif                | Action                                                |
| ----------------------- | ----------------------------------------------------- |
| Vue d'ensemble          | Lire `.backlog/backlog.md`                            |
| Détail d'une tâche      | Lire `.backlog/tasks/<ID>.md`                          |
| Nouvelle tâche          | Copier le template, incrémenter le numéro             |
| Changer de statut       | Mettre à jour `status` + `updated` puis `backlog.md`   |
| Dépendances             | `depends_on` et `blocks` mis à jour des deux côtés     |

## Arborescence canonique

```
.backlog/
├── project.md      # config projet + external_projects
├── AGENTS.md      # règles du backlog pour les agents
├── backlog.md     # tableau de bord groupé par statut
├── tasks/         # <PROJET>-<TYPE>-<NUM>.md (TASK, BUG, STORY, SPIKE, CHORE)
├── epics/
└── sprints/
```

Ce projet utilise le préfixe `DJC` (ex. `DJC-TASK-001`).

## Template de tâche

```yaml
---
id: "DJC-TASK-001"
type: "TASK"
status: "todo"          # todo | in-progress | blocked | review | done
priority: "medium"      # low | medium | high | critical
labels: []
components: []
story_points: null
epic: null
sprint: null
parent: null
children: []
depends_on: []
blocks: []
related_to: []
tokens_estimated: 0     # convention locale
tokens_used: 0          # convention locale
created: "YYYY-MM-DD"
updated: "YYYY-MM-DD"
---
## Summary
## Description
## Acceptance Criteria
- [ ] ...
```

## Règles agents

- `todo -> in-progress` seulement si toutes les `depends_on` sont `done`.
- Pas de cycle : valider le DAG à chaque écriture.
- Parser le frontmatter en YAML, jamais par regex approximative.
- Garder `backlog.md` synchronisé avec les fichiers de tâches.
- Dates au format `YYYY-MM-DD`, jamais de placeholder `<< >>` restant.

## Budget de tokens (convention du projet)

Avant d'attaquer une tâche : estimer l'ordre de grandeur dans
`tokens_estimated` (exploration + édition + vérification). En fin de tâche,
écrire `tokens_used` et une ligne de rétro dans `## Notes` si l'écart dépasse
2x. Une tâche dont l'estimation dépasse ~40k tokens doit être découpée en
sous-tâches (`children`).

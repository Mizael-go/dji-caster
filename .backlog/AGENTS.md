# Règles du backlog pour les agents

1. **Lire avant d'écrire** : `backlog.md` puis le ticket concerné.
2. **Un ticket, un fichier** dans `tasks/`, frontmatter YAML valide, jamais de
   placeholder `<< >>`.
3. **Transitions** : `todo → in-progress` uniquement si toutes les `depends_on`
   sont `done`. `in-progress → review → done`.
4. **DAG** : pas de cycle, `depends_on` et `blocks` symétriques.
5. **Synchronisation** : toute modification de statut met à jour
   `status`, `updated` dans le ticket **et** la ligne correspondante dans
   `backlog.md`.
6. **Budget** : estimer `tokens_estimated` avant de commencer, écrire
   `tokens_used` à la fin ; au-delà de ~40k estimés, découper en `children`.
7. **Preuves** : un ticket passe `done` avec la commande de vérification
   exécutée (build, test, APK produit) notée dans `## Notes`.
8. Les tickets décrivent le produit, pas la conversation : pas de « comme
   demandé plus haut ».

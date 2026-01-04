# 00 - Vue d'ensemble

But : fournir une description courte et factuelle de l'intention de ce
template, destinée à l'auteur pour une relecture future. Ce document
explique le pourquoi (raison d'être) et le périmètre — pas le comment.

Pourquoi ce template existe

- Fournir une base réutilisable, minimale et vérifiable pour démarrer un
  microservice HTTP en Go sans dépendances externes non nécessaires.
- Réduire le travail répétitif d'initialisation (CLI, config, serveur,
  middleware d'observabilité, endpoints opérationnels) pour accélérer
  les démarrages de projet.

Problèmes que ce template cherche explicitement à résoudre

- Standardiser la gestion d'un serveur HTTP (démarrage, shutdown, timeouts).
- Fournir des endpoints opérationnels stables (`/live`, `/ready`).
- Offrir une abstraction légère pour les métriques (interchangeable).
- Fournir un logging structuré minimal et reproductible (`slog`).

Ce que ce template n'essaie PAS de résoudre

- Il n'est pas un framework complet d'architecture métier ni une
  implémentation d'un domaine applicatif.
- Il n'impose aucun backend d'observabilité (metrics/logs/tracing).
- Il ne fournit pas d'intégration vendor-specific; ces exemples doivent
  être placés dans `examples/` si ajoutés.

Usage prévue

- Servir de base/template de démarrage (fork/clone) ou de référence
  pour de nouveaux microservices. Conserver la structure pour faciliter
  l'évolution et la revue technique.

Non-goals (récapitulatif)

- Ne pas fournir des patterns prêts à l'emploi pour le domaine métier.
- Ne pas décider d'un backend métrique/logging externe.
- Ne pas inclure des exemples vendor-specific dans la documentation
  centrale (README.md reste vendor-neutral).

Assumptions

- Aucune hypothèse opérationnelle non visible dans le code n'est faite
  dans ce document.

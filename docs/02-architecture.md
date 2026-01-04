# 02 - Architecture (séparation des responsabilités)

But : expliciter la séparation des couches observée dans le code et
décrire ce qui vit dans chaque couche. Document destiné à l'auteur pour
conserver la mémoire des choix réalisés.

Principes observés

- Séparer le transport (HTTP), l'application (services), le domaine et
  l'infrastructure (logging, metrics, serveur). Cette séparation est
  visible dans l'organisation des packages.

Cartographie vers les packages du dépôt

- Transport HTTP : `internal/transport/http`
  - `handler/` : handlers Gin (ex. `HealthHandler`, `UsersHandler`)
  - `middleware/` : middleware (RequestID, StructuredLogger, Metrics)
  - `router/` : construction du routeur et enregistrement des routes
- Configuration : `internal/config` (Viper + validation des valeurs)
- Server/lifecycle : `internal/server` (http.Server wrapper, shutdown)
- Observability/Infra :
  - `internal/logging` (slog wrapper)
  - `internal/observability/metrics` (Recorder abstraction + noop)
- Health checks : `internal/health` (Aggregator et checks)
- Domaine / application : `internal/domain` (pkg `apperr`), interfaces
  de services exposées vers les handlers (ex. `UserService` interface)

Ce qui vit dans chaque couche (concret)

- Handlers (`handler`) :
  - Lire la requête (binding), valider DTOs, appeler les services
  - Mapper erreurs via `errmap` et renvoyer des réponses via `response`
  - Ne pas contenir la logique métier ni accès direct aux infra
- Services / Domain (`internal/domain`, pas de dossier `service` explicite) :
  - Interfaces métier (ex. `UserService`) et implémentations (à fournir
    par le projet qui utilise le template)
- Infrastructure :
  - Logging, metrics, serveur HTTP, configuration, readiness checks

Ce qui est explicitement prévu / visible

- `UsersHandler` dépend d'une interface `UserService` — le template
  montre la dépendance inversée (handler -> interface), pas
  l'implémentation concrète.
- `internal/health.Aggregator` gère les checks de readiness et définit
  la politique « no checks = not ready » (implémentation explicite).

Ce qui est interdit (règles d'usage recommandées)

- Assumption : le template vise à garder la logique métier hors des
  handlers. Par conséquent, évitez d'implémenter des accès SQL/IO
  directement dans les handlers.

Pourquoi cette séparation

- Testabilité : les handlers deviennent testables en moquant les
  interfaces métier.
- Isolation des changements : changer l'infra (ex. remplacer metrics)
  n'implique pas de modifications profondes dans le domaine.

Limitation

- Ce dépôt démontre la structure et quelques patterns (handlers,
  middleware, abstractions d'observabilité). Il ne représente pas une
  implémentation complète du domaine métier — le service concret doit
  être fourni par le projet qui consomme ce template.

Assumptions

- La consigne « pas de SQL dans handlers » est un principe recommandé
  par ce template et non une règle appliquée par le code.

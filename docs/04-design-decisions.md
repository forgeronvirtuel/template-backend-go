# 04 - Decisions de conception

Ce fichier conserve la mémoire des décisions techniques visibles dans le
code ou explicitement mentionnées dans le README. Si la date n'est
pas disponible dans le dépôt, elle est indiquée comme **date inconnue**.

## date inconnue — Utiliser Gin pour le transport HTTP

### Decision

Utiliser `github.com/gin-gonic/gin` comme routeur HTTP et framework de
transport léger.

### Context

Le code du routeur (`internal/transport/http/router/router.go`) construit
un `gin.Engine` et enregistre les middleware et routes.

### Alternatives considered

- `net/http` standard library
- d'autres frameworks (Echo, Chi)

### Consequences

- - Avantage : router connu et concise gestion des routes/middleware
- - Inconvénient : dépendance externe mineure

## date inconnue — Abstraction metrics via `metrics.Recorder` + no-op

### Decision

Fournir une interface `Recorder` et une implémentation no-op par défaut.

### Context

`internal/observability/metrics/metrics.go` définit l'interface et la
no-op implémentation. Le router substitue `metrics.NewNoop()` si aucun
`Recorder` n'est fourni.

### Alternatives considered

- Forcer un backend (Prometheus, OTEL) au template
- Ne pas instrumenter du tout

### Consequences

- - Permet l'instrumentation sans imposer de fournisseur
- - Safe-by-default: pas d'échec si pas de backend
- - Le développeur doit fournir l'intégration concrète s'il veut des métriques

## date inconnue — Health endpoints `/live` et `/ready` séparés

### Decision

Exposer `/live` pour liveness et `/ready` pour readiness; les deux hors
du contrat API métier.

### Context

Routes définies dans `router.New` et handlers dans
`handler/health.go`. `Ready` utilise `internal/health.Aggregator`.

### Alternatives considered

- Un seul endpoint `/health` combiné

### Consequences

- - Conformité avec les attentes des orchestrateurs (probe separation)
- - Readiness permet d'exposer l'état des dépendances
- - Demande un soin supplémentaire pour configurer les checks

## date inconnue — Request ID via header `X-Request-Id` et propagation

### Decision

Générer/propager `X-Request-Id` en middleware (`RequestID`) et l'inclure
dans les réponses métiers via `meta.request_id`.

### Context

Middleware dans `internal/transport/http/middleware/request_id.go` et
utilisation dans `response.OK/Fail`.

### Alternatives considered

- Ne pas propager `request_id`
- Utiliser un autre header

### Consequences

- - Facilite corrélation logs <-> réponses
- - Nécessite que le client accepte/retienne l'en-tête si utile côté client

## date inconnue — Logging structuré avec `slog`

### Decision

Utiliser `log/slog` pour logs JSON structurés et exposer helpers dans
`internal/logging`.

### Context

`internal/logging/logger.go` initialise un `slog` JSON handler.

### Alternatives considered

- logrus, zap, zerolog

### Consequences

- - Pas de dépendance externe additionnelle (stdlib)
- - Sortie JSON prête pour ingestion
- - `slog` relativement récent ; certains outils tiers ont intégrations
    moins matures

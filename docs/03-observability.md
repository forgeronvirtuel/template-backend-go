# 03 - Observabilité

But : expliquer les choix d'observabilité implémentés (logs, metrics,
health) et leur justification technique telle qu'elle apparaît dans le
code.

Logs vs Metrics vs Health

- Logs : informations événementielles (audit, erreurs, diagnostic).
- Metrics : séries temporelles agrégées (compteurs, histogrammes, gauges).
- Health endpoints : sondes opérationnelles pour l'orchestrateur (`/live`, `/ready`).

Pourquoi `slog` pour le logging

- Le projet utilise `internal/logging` qui initialise `slog` et expose
  une API simple (`Init`, `Logger`, `LoggerFromContext`). Le choix de
  `slog` est visible dans le code et produit des logs JSON structurés
  sans dépendances externes.

Pourquoi un logging structuré et context-aware

- Un log structuré (JSON) facilite l'ingestion automatique par des
  systèmes d'agrégation et le filtrage.
- La propagation du `request_id` dans le contexte permet la corrélation
  entre logs et réponses HTTP (implémenté via `RequestID` middleware
  et `LoggerFromContext`).

Pourquoi une abstraction metrics backend-agnostic

- `internal/observability/metrics` définit une interface `Recorder`
  et une implémentation no-op. Cette abstraction rend le code
  indépendant d'un fournisseur (Prometheus, OTEL, Datadog...), comme
  indiqué explicitement dans les commentaires de `metrics.go`.

Pourquoi une implémentation no-op

- Le commentaire dans `metrics.go` indique la raison : garantir que le
  code s'exécute en sécurité même si aucun backend métrique n'est
  configuré. Cela offre un comportement « safe by default ». Le
  router substitute `metrics.NewNoop()` lorsque `Recorder` est nil.

Règles cardinalité pour les labels

- Le middleware metrics construit des labels `method`, `route` (via
  `c.FullPath()`) et `status` — explicitement pour maintenir une
  faible cardinalité.
- Les labels à haute cardinalité **ne doivent pas** être exposés
  (pas de `request_id`, `user_id`, IP, params bruts), ce qui est
  documenté dans `internal/observability/metrics/metrics.go`.

Différence et rôle de `/live` vs `/ready`

- `/live` : sonde de liveness (process is up). Retourne JSON simple
  `{"status":"ok"}` et éventuellement `version`.
- `/ready` : exécute les checks via `internal/health.Aggregator` et
  indique si les dépendances critiques sont prêtes. Politique
  observable : si aucun check n'est enregistré, l'Aggregator renvoie
  `Ready = false` (service NOT ready).

Notes opérationnelles

- Les endpoints opérationnels ne sont pas enveloppés par le contrat
  métier et doivent rester simples pour permettre aux orchestrateurs
  d'évaluer rapidement l'état du processus.

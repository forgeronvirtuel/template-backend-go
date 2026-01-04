# Brief technique pour un agent IA

But du document

- Fournir à un agent IA un résumé technique, factuel et actionnable du
  dépôt afin qu'il puisse donner des conseils de développement en
  respectant l'état actuel du code et les contraintes explicites.

Contrainte de production

- Ne rien inventer : tout doit être vérifiable dans le code ou le
  README. Si une information manque, demander plutôt que supposer.

Sources fiables à utiliser en priorité

- `cmd/serve.go`, `cmd/root.go`
- `internal/transport/http/router/router.go`
- `internal/transport/http/handler/health.go`
- `internal/transport/http/handler/users.go`
- `internal/transport/http/middleware/*`
- `internal/observability/metrics/metrics.go`
- `internal/logging/logger.go`
- `internal/config/config.go`
- `internal/transport/http/response/response.go`
- `internal/security/auth/auth.go` (authentification)

Résumé factuel du projet (faits vérifiés)

- Entrée : `main.go` appelle `cmd.Execute()` (Cobra).
- Commande principale : `serve` (sous-commande Cobra) démarre le
  serveur HTTP.
- Flags `serve` :
  - `-p` / `--port` (default `8080`) -> lié à `server.port` (Viper).
  - `-H` / `--host` (default `0.0.0.0`) -> lié à `server.host`.
  - `-l` / `--log-level` (default `info`) -> lié à `log.level`.
- Configuration : Viper lit `--config` si fourni, sinon `config.yaml`
  (recherche dans le répertoire courant) et accepte variables d'env.
- Router (Gin) construit dans `router.New` et exige `Health` handler.
  Middleware appliqués, dans l'ordre : `RequestID`, `StructuredLogger`, `Metrics`.
- Authentification (optionnelle) :
  - Par défaut désactivée (`auth.enabled=false`) mais applique strict mode :
    les routes protégées retournent 401.
  - Si activée : supporte `APIKeyAuthenticator` (clés statiques depuis
    `Authorization: Bearer` ou `X-API-Key`).
  - Interface `auth.Authenticator` permet d'implémenter JWT/OAuth2/etc.
  - Middleware `Auth` positionné sur groupe protégé sous `/api/v1`.
  - `/live` et `/ready` restent publics (pas de middleware auth).
- Routes exposées :
  - `GET /live` — liveness (réponse JSON minimale, option `version`) **PUBLIC**.
  - `GET /ready` — readiness (utilise `internal/health.Aggregator`) **PUBLIC**.
  - `POST /api/v1/users` — **PROTÉGÉE** (auth requise), disponible uniquement
    si un `UsersHandler` est fourni lors de la construction du router.
- Health semantics :
  - `Live` renvoie `{"status":"ok"}` (+ `version` si définie).
  - `Ready` exécute les checks et renvoie HTTP 200 si prêt, HTTP 503 si
    un check critique échoue. Si aucun check n'est enregistré,
    l'aggregator renvoie `not_ready` (policy explicite).
- Logging : wrapper `internal/logging` initialise `slog` JSON handler;
  middleware `StructuredLogger` logge `request_id`, `method`, `path`,
  `status_code`, `latency_ms`, `client_ip`.
- Request ID : middleware lit/écrit header `X-Request-Id`, génère une
  valeur si absente et la met dans le contexte (clé `request_id`).
- Metrics : abstraction `metrics.Recorder` avec type `Labels []Label` et
  une implémentation no-op `NewNoop()` ; middleware enregistre `method`,
  `route` (via `c.FullPath()`), `status` et latency (seconds).
- Responses métier : helpers `response.OK` et `response.Fail` produisent
  les formats JSON définis (`data/meta` pour succès, `error/meta` pour erreurs).
- Tests : présence de fichiers `_test.go` (exécution : `go test ./...`).

Contraintes explicites à respecter par l'agent IA

- Ne pas documenter ou supposer d'endpoints qui n'existent pas (p.ex.
  `/health` n'existe pas dans le code et ne doit pas être mentionné).
- `/live` et `/ready` sont hors contrat API métier (pas de wrapper
  `data/meta` ni `error/meta`).
- README doit rester vendor-neutral pour les métriques ; tout exemple
  vendor-specific doit être déplacé vers `examples/` si ajouté.
- Ne pas proposer de modifications qui ajoutent des labels métriques à
  haute cardinalité (`request_id`, `user_id`, IP, etc.).
- Pour l'authentification :
  - Ne jamais logger les credentials (clés API, tokens, mots de passe).
  - Ne jamais exposer les clés dans les logs, métriques ou réponses.
  - Respecter le strict mode : si auth désactivée, routes protégées → 401.
  - Ne pas appliquer le middleware Auth à `/live` et `/ready`.

Checklist rapide pour conseiller sur une PR

- Vérifier que les changements sont supportés par les tests existants
  et ajouter des tests unitaires si nécessaire.
- Si modification d'un handler, s'assurer que :
  - Les DTOs sont validés au niveau du handler (utilisez
    `validation.FromBindError`).
  - Les erreurs métiers sont mappées via `errmap` et renvoyées via
    `response.Fail`.
- Si changement d'observabilité :
  - Ne pas modifier `metrics.Recorder` public API sans justification.
  - Pour ajouter un backend, fournir une implémentation de
    `metrics.Recorder` dans `internal/observability` ou `examples/`.
  - Respecter les règles de cardinalité (method, route, status).
- Si changement de logging :
  - Utiliser `internal/logging.Init` pour initialiser le niveau.
  - Utiliser `Logger()` ou `LoggerFromContext()` ; ne pas remplacer
    globalement la manière dont le middleware attache le logger au contexte.
- Si ajout de routes :
  - Enregistrer les routes dans `router.New` ou via une fonction qui
    reçoit `router.Deps` pour garder la construction centralisée.
  - Si route protégée : l'enregistrer sous le groupe `protected` qui
    utilise `middleware.Auth`.
  - Si route publique sous `/api/v1` : l'enregistrer directement sur `v1`.
- Si modification de l'authentification :
  - Respecter l'interface `auth.Authenticator`.
  - Tester avec `DisabledAuthenticator` pour vérifier le strict mode.
  - Ajouter tests unitaires dans `internal/security/auth/*_test.go`.
  - Ne jamais logger les credentials.

Questions que l'agent IA doit poser avant d'agir (si non fournies)

1. Voulez-vous brancher une implémentation concrète pour `UserService` ?
2. Faut-il activer et intégrer un backend métrique (si oui, lequel) ?
3. Voulez-vous activer l'authentification ? Si oui, quelle implémentation
   (`APIKeyAuthenticator` simple ou JWT/OAuth2 custom) ?
4. Souhaitez-vous que `TODO.md` soit mis à jour pour refléter l'état
   actuel (quelques items semblent obsolètes) ?

Assumptions (si l'agent doit en faire, les lister explicitement)

- `buildinfo.Version` peut être utilisé pour exposer la version si
  la variable est renseignée au moment du build (le code logge cette
  valeur si disponible).

Requêtes d'actions types que l'agent peut proposer

- Proposer un patch minimal pour brancher un `UserService` mock dans
  `cmd/serve.go` pour démonstration locale (ne pas lier de dépendances
  réelles).
- Proposer tests unitaires pour un nouveau handler ou pour
  `internal/health.Aggregator`.
- Ajouter un exemple d'implémentation de `metrics.Recorder` dans
  `examples/` (le README doit rester vendor-neutral).

Emplacement du travail pour modifications typiques

- Ajouter service métier : `internal/domain` (implémentation) et
  brancher l'implémentation dans `cmd/serve.go` via `router.Deps`.
- Ajouter checks de readiness : créer un type implémentant
  `internal/health.ReadinessCheck` et le fournir à
  `health.NewReadinessAggregator` dans `cmd/serve.go`.
- Intégrer un backend métrique : ajouter une implémentation de
  `metrics.Recorder` et la passer dans `router.Deps.Metrics`.
- Implémenter auth custom (JWT, OAuth2) :
  - Créer une struct implémentant `auth.Authenticator` dans
    `internal/security/auth/` ou sous-package dédié.
  - Wire dans `cmd/serve.go` via `router.Deps.Auth`.
  - Ajouter tests dans `internal/security/auth/*_test.go`.
- Ajouter route protégée : enregistrer sous le groupe `protected` dans
  `router.New`.
- Ajouter route publique business : enregistrer directement sur `v1`
  (avant le groupe `protected`).

Format attendu des réponses de l'agent IA (préciser au modèle)

- Toujours citer les fichiers sources utilisés pour conclure.
- Si une proposition modifie le README, s'assurer qu'elle respecte
  la contrainte vendor-neutral (déplacer exemples vendor-specific vers
  `examples/`).
- Pour toute nouvelle hypothèse, ajouter une section `Assumptions`.

Fin.

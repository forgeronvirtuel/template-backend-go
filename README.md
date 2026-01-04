# template-backend-go

Ce dépôt est un template minimal de microservice HTTP en Go. Le contenu
du README suit strictement l'implémentation actuelle présente dans le
dépôt (routes exposées, flags CLI, configuration, comportement des
endpoints opérationnels, logging et abstraction des métriques).

**Important :** toute affirmation dans ce document est vérifiable dans
le code source. Aucune fonctionnalité non-présente dans le dépôt n'est
documentée ici.

**Langue:** ce README est en français. Les extraits de code sont en
anglais car ils reflètent des noms de symboles réels.

## Ce que garantit ce template

- Un binaire CLI avec une sous-commande `serve` qui lance un serveur HTTP.
- Endpoints opérationnels standardisés : `/live` (liveness) et `/ready` (readiness).
- Un routeur HTTP (Gin) avec middleware pour `request id`, logging structuré
  et instrumentation métrique (abstraction via `metrics.Recorder`).
- Logging structuré basé sur `slog` avec sortie JSON.
- Une abstraction simple pour les métriques (`internal/observability/metrics`).

## Ce que le template n'impose pas

- Aucune implémentation de backend métrique particulière n'est fournie
  (le template inclut une implémentation no-op). Le choix du fournisseur
  (Prometheus, OTEL, Datadog...) est laissé au projet.
- Aucune API métier n'est imposée — seules des structures et handlers
  d'exemple sont présentes.

## Structure importante du dépôt

- `main.go` : point d'entrée CLI qui appelle `cmd.Execute()`.
- `cmd/` : commandes Cobra. `serve` est la commande qui démarre le serveur.
- `internal/config` : lecture et validation de la configuration via Viper.
- `internal/server` : encapsulation du `http.Server` et gestion du shutdown.
- `internal/transport/http/router` : construction du routeur Gin et ordre des middleware.
- `internal/transport/http/handler` : handlers HTTP (ex. `HealthHandler`, `UsersHandler`).
- `internal/transport/http/middleware` : `RequestID`, `StructuredLogger`, `Metrics`.
- `internal/logging` : wrapper autour de `slog` (Init, Logger, context helpers).
- `internal/observability/metrics` : abstraction `Recorder` et implémentation no-op.
- `internal/health` : agrégateur de checks pour `/ready`.

## Endpoints exposés (réalité du code)

Les routes réellement enregistrées par le routeur (voir
`internal/transport/http/router/router.go`) sont :

- `GET /live` — endpoint de liveness opérationnelle.
- `GET /ready` — endpoint de readiness contenant le résumé des checks.
- `POST /api/v1/users` — présent uniquement si un `UsersHandler` est fourni
  lors de la construction du routeur ; le binaire fourni par défaut NE
  raccorde pas le `UsersHandler` (voir `cmd/serve.go`).

Remarques importantes :

- Seuls `/live` et `/ready` sont documentés comme endpoints opérationnels
  : il n'existe pas de `/health` dans le code, n'en parlez pas.
- Les endpoints opérationnels n'appliquent PAS le contrat API métier
  (pas de wrapper `data/meta` ni `error/meta`) : ils retournent du JSON
  minimal tel que défini dans `internal/transport/http/handler/health.go`.

### Comportement précis

- `GET /live` :

  - Réponse HTTP 200 avec body JSON minimal : `{"status":"ok"}` et
    éventuellement la clé `version` si fournie par le service.
  - Usage : vérifie que le processus est vivant.

- `GET /ready` :
  - Utilise `internal/health.Aggregator` pour exécuter les checks.
  - Si l'aggregator est construit sans checks (cas par défaut si vous
    n'enregistrez aucun check), la réponse indique que le service
    N'EST PAS prêt (`status: "not_ready"`) et la liste `checks` est vide.
  - Si un check critique échoue, l'endpoint renvoie HTTP 503,
    sinon HTTP 200.
  - Body JSON contient `status`, `checks` (liste d'objets nommés,
    `critical`, `status` et éventuellement `error`) et optionnellement `version`.

## Contrat API métier (ce qui existe dans le template)

Le template fournit un exemple de `UsersHandler` dans
`internal/transport/http/handler/users.go`. Les handlers métiers utilisent
les helpers suivants :

- `internal/transport/http/response` pour formater les réponses success/erreur
- `internal/transport/http/validation` pour extraire les détails de validation
- `internal/transport/http/errmap` pour mapper erreurs domaine -> HTTP

Par défaut, la commande `serve` fournie ne branche pas de `UsersHandler`.
Si vous ajoutez votre service métier, fournissez-le à `router.Deps{Users: ...}`
au moment de créer le router.

## Middleware et observabilité

- Order des middleware (voir `router.New`):

  1. `RequestID()` — positionnée globalement, ajoute/envoie l'en-tête
     `X-Request-Id` et stocke la valeur dans le contexte Gin.
  2. `StructuredLogger()` — log JSON structuré par requête via `slog`.
  3. `Metrics(...)` — instrumentation HTTP (compteurs, histogrammes, gauge).

- Logging :

  - Initialisé via `internal/logging.Init(level)` (appelé depuis `runServe`).
  - La middleware logge un attribut unique par requête contenant :
    `request_id`, `method`, `path`, `status_code`, `latency_ms`, `client_ip`.

- Metrics :
  - Interface `metrics.Recorder` définie dans `internal/observability/metrics`.
  - Le router accepte un `metrics.Recorder` optionnel ; si `nil`, on
    utilise `metrics.NewNoop()` (aucune donnée envoyée).
  - Middleware metrics crée les labels `method`, `route` (template via `c.FullPath()`)
    et `status`. Le nom des métriques utilisées est visible dans
    `middleware/metrics.go` (`http_requests_total`,
    `http_request_duration_seconds`, `http_requests_in_flight`).
  - Important : les métriques n'incluent PAS de labels à haute
    cardinalité (pas de `request_id`, `user_id`, IP, etc.).

Note : par design, ce README reste agnostique vis-à-vis d'un backend
de métriques spécifique. Des exemples de backends (Prometheus/OTEL/Datadog)
sont présents en commentaire dans le code d'abstraction des métriques ;
comme politique du template, le README ne contient aucun exemple
spécifique à un fournisseur. Si vous souhaitez des snippets d'intégration
vendor-specific, créez `examples/` dans le dépôt.

## Configuration

- Fichier de configuration : `config.yaml` (recherché dans le répertoire
  courant par défaut). Le CLI a aussi un flag `--config` pour en préciser
  un autre chemin (voir plus bas).
- Viper est utilisé pour la lecture (`root.initConfig`) et appelle
  `viper.AutomaticEnv()` : les variables d'environnement peuvent
  surcharger les valeurs (conformes aux clés Viper utilisées dans le code).

### Clés et valeurs par défaut (vérifiables dans `internal/config/config.go`)

- `server.host` : par défaut `0.0.0.0`
- `server.port` : par défaut `8080`
- `log.level` : par défaut `info` (défini via flags/viper dans `cmd/serve.go`)

Exemple minimal de `config.yaml` :

```yaml
server:
  host: "0.0.0.0"
  port: 8080
```

### Flags CLI (exactement comme définis dans `cmd/`)

- `--config` : chemin vers le fichier de config (flag persistant défini
  dans `root.go`).

Sous-commande `serve` (flags exacts dans `cmd/serve.go`) :

- `-p`, `--port` : port (uint16) — default `8080`
- `-H`, `--host` : host bind — default `0.0.0.0`
- `-l`, `--log-level` : log level (`debug|info|warn|error`) — default `info`

Ces flags sont liés à Viper : `server.port`, `server.host`, `log.level`.

## Exemples d'utilisation (Quickstart)

Prérequis : Go (1.20+ recommandé).

Lancer rapidement sans installer :

```bash
go run main.go serve
```

Ou construire un binaire puis lancer :

```bash
go build -o template-backend-go .
./template-backend-go serve --port 8080 --host 127.0.0.1
```

Spécifier un fichier de config :

```bash
./template-backend-go --config ./config.yaml serve
```

Vérifier les endpoints opérationnels :

```bash
curl -sS http://127.0.0.1:8080/live
curl -sS http://127.0.0.1:8080/ready
```

## Tests

Ce dépôt contient des tests unitaires et d'intégration (fichiers `_test.go`).
Pour exécuter tous les tests :

```bash
go test ./...
```

## Roadmap / TODO

Le fichier `TODO.md` à la racine contient les tâches à réaliser. Certains
éléments présents dans ce dépôt sont déjà implémentés (ex. `RequestID`
middleware qui pose `X-Request-Id` et le renvoie dans la réponse). Mettez
à jour `TODO.md` pour refléter l'état actuel si vous ajoutez/terminez des
travaux.

## Opérations (ops)

- Endpoints opérationnels : `/live` et `/ready` (voir plus haut).
- Logs : format JSON via `slog` (stdout), niveau contrôlé par `--log-level`.
- Metrics : abstraction via `metrics.Recorder` — par défaut no-op.

## Remarques finales

Tout ce qui est documenté ci‑dessus est tiré directement du code source
(`cmd/`, `internal/`). Ne pas ajouter d'items non présents dans le code
ou la configuration. Si vous désirez des exemples d'intégration de backends
de métriques (Prometheus, OTEL, Datadog), créez un dossier `examples/`
et placez-y des snippets d'intégration ; le README restera volontairement
vendor-neutral.

---

Pour toute clarification sur ce README, consultez les fichiers sources
cit és ci-dessus : `cmd/serve.go`, `internal/transport/http/router/router.go`,
`internal/transport/http/handler/health.go`, `internal/observability/metrics/metrics.go`,
`internal/logging/logger.go`, `internal/config/config.go`.

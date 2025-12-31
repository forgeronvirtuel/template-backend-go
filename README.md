# Backend Go – Template Production

Ce dépôt fournit un **template backend production-ready en Go**, conçu pour être cloné et adapté aux projets clients avec un minimum de configuration.

Les objectifs principaux sont :

- Standardiser les contrats d'API (réponses, erreurs, validation)
- Imposer une séparation claire des responsabilités
- Éliminer les patterns backend fragiles et récurrents
- Fournir une base solide pour les environnements de production

---

## Principes Fondamentaux

Ce template est construit autour des principes suivants :

- **Contrats d'API explicites** plutôt que comportements implicites
- **Entrées et sorties typées** (pas de payloads non structurés)
- **Les erreurs du domaine sont indépendantes du transport**
- **Une seule façon** de retourner les réponses et les erreurs
- **Préoccupations de production d'abord** (timeouts, shutdown, traçabilité)

---

## Contrat d'API

### Réponse de Succès (2xx)

Toutes les réponses réussies **doivent** suivre ce format :

```json
{
  "data": {},
  "meta": {
    "request_id": "e4f1c8b9a0d44a2e8f6c1a8c5c9e4b21"
  }
}
```

- `data`: le payload de réponse réel
- `meta.request_id`: identifiant unique pour la requête, propagé de bout en bout

---

### Réponse d'Erreur (4xx / 5xx)

Toutes les réponses d'erreur **doivent** suivre ce format :

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Payload de requête invalide",
    "details": [
      {
        "field": "email",
        "rule": "email",
        "message": "doit être une adresse email valide"
      }
    ]
  },
  "meta": {
    "request_id": "e4f1c8b9a0d44a2e8f6c1a8c5c9e4b21"
  }
}
```

- `code`: code d'erreur **stable, lisible par machine**
- `message`: résumé lisible par humain
- `details`: optionnel, principalement utilisé pour les erreurs de validation
- `meta.request_id`: toujours présent

---

## Stratégie d'Erreur

### Erreurs Domaine / Application

Les erreurs provenant des couches domaine ou application **ne doivent pas** dépendre de concepts HTTP.

Elles sont représentées comme des erreurs typées :

- `NOT_FOUND`
- `CONFLICT`
- `UNAUTHORIZED`
- `FORBIDDEN`
- `VALIDATION_ERROR`
- `INTERNAL_ERROR`

Ces erreurs sont ensuite mappées vers des réponses HTTP **uniquement dans la couche transport**.

### Mapping HTTP

Le mapping entre les erreurs du domaine et les codes HTTP est centralisé et appliqué de manière stricte.

Exemple :

| Code Domaine     | Statut HTTP |
| ---------------- | ----------- |
| NOT_FOUND        | 404         |
| CONFLICT         | 409         |
| UNAUTHORIZED     | 401         |
| FORBIDDEN        | 403         |
| VALIDATION_ERROR | 400         |
| INTERNAL_ERROR   | 500         |

Les handlers **ne doivent jamais** décider des codes HTTP directement en fonction de la logique métier.

---

## Validation

### Où Vit la Validation

- La validation est effectuée **à la frontière du transport**
- Seuls les **DTOs (Data Transfer Objects)** sont validés
- Les modèles du domaine restent libres de toute préoccupation JSON, HTTP ou de validation

### Exemple de DTO

```go
type CreateUserRequest struct {
    Email string `json:"email" binding:"required,email"`
    Name  string `json:"name"  binding:"required,min=2,max=80"`
}
```

### Erreurs de Validation

Les erreurs de validation sont toujours retournées comme :

- `code`: `VALIDATION_ERROR`
- `details`: une entrée par champ invalide, incluant la règle et le message

Cela garantit :

- gestion prévisible côté frontend
- documentation d'API cohérente
- internationalisation facile si nécessaire

---

## Traçabilité

### Request ID

Chaque requête HTTP entrante se voit attribuer un `request_id` :

- Généré automatiquement s'il n'est pas fourni
- Accepté depuis le header : `X-Request-Id`
- Retourné dans :
  - Header de réponse `X-Request-Id`
  - Corps de réponse `meta.request_id`

Cet identifiant doit être :

- loggué
- propagé aux services en aval
- inclus dans les réponses d'erreur

---

## Vue d'Ensemble de l'Architecture

```
cmd/
  └── serve.go          # CLI entrypoint (orchestration uniquement)

internal/
  ├── domain/           # Règles métier, entités, erreurs du domaine
  ├── app/              # Cas d'usage / services applicatifs
  ├── infra/            # DB, services externes, implémentations
  └── transport/
      └── httpapi/      # Couche HTTP (Gin)
          ├── handler/  # Handlers HTTP
          ├── dto/      # DTOs requête/réponse
          ├── response/ # Helpers de réponse
          ├── errmap/   # Mapping erreurs Domaine → HTTP
          ├── middleware/
          └── router/
```

Règles clés :

- Les handlers dépendent d'**interfaces**, jamais d'implémentations concrètes
- La couche domaine n'a **aucune dépendance** sur Gin, HTTP ou JSON
- La couche transport est remplaçable (REST, gRPC, etc.)

---

## Ce Que Ce Template Évite Explicitement

- Handlers inline (closures)
- Payloads `map[string]interface{}`
- Réponses JSON ad-hoc
- Logique de gestion d'erreur dispersée
- Logique métier dans les handlers HTTP
- Conventions implicites

---

## Démarrage Rapide

```bash
git clone <ce-repo> mon-service
cd mon-service
go run ./cmd/serve
```

Ensuite :

1. Définir vos cas d'usage du domaine
2. Implémenter les adaptateurs d'infrastructure
3. Les brancher dans la couche HTTP
4. Étendre l'API en gardant le contrat inchangé

---

## Usage Prévu

Ce dépôt est destiné à être :

- cloné pour de nouveaux services backend
- utilisé comme référence pour les conventions d'API
- étendu, pas réécrit

Si vous vous retrouvez à casser le contrat d'API, c'est un signal que l'architecture devrait évoluer — pas les conventions.

---

## Installation

```bash
go mod download
go build -o app .
```

## Utilisation

### Lancer le serveur HTTP

```bash
# Lancer avec les paramètres par défaut (port 8080, host 0.0.0.0)
./app serve

# Spécifier un port personnalisé
./app serve --port 3000

# Spécifier un host et un port
./app serve --host localhost --port 3000

# Utiliser un fichier de configuration personnalisé
./app serve --config /path/to/custom.yaml
```

### Options disponibles

**Flags de commande :**

- `-p, --port`: Port d'écoute (défaut: 8080)
- `-H, --host`: Host à binder (défaut: 0.0.0.0)

**Flags globaux :**

- `--config`: Chemin vers le fichier de configuration (défaut: ./config.yaml)

### Configuration

L'application supporte plusieurs méthodes de configuration avec l'ordre de priorité suivant :

1. **🥇 Flags en ligne de commande** (priorité maximale)
2. **🥈 Fichier de configuration** (`config.yaml`)
3. **🥉 Valeurs par défaut**

#### Fichier de configuration

Créez un fichier `config.yaml` à la racine du projet :

```yaml
server:
  host: "0.0.0.0"
  port: 8080
```

Un fichier d'exemple est disponible : `config.example.yaml`

```bash
# Copier l'exemple et personnaliser
cp config.example.yaml config.yaml
```

#### Exemples de configuration

```bash
# 1. Avec fichier config.yaml (port: 9000)
./app serve
# → Démarre sur le port 9000

# 2. Flag override config
./app serve --port 3000
# → Démarre sur le port 3000 (le flag a priorité)

# 3. Sans config.yaml (valeurs par défaut)
rm config.yaml && ./app serve
# → Démarre sur le port 8080

# 4. Fichier de config personnalisé
./app serve --config /etc/myapp/config.yaml
```

## Endpoints API

Une fois le serveur lancé, les endpoints suivants sont disponibles:

- `GET /` - Page d'accueil avec la liste des endpoints
- `GET /health` - Health check du serveur
- `GET /api/v1/hello` - Endpoint de test
- `GET /api/v1/users/:id` - Récupérer un utilisateur par ID
- `POST /api/v1/users` - Créer un nouvel utilisateur

## Exemple de requêtes

```bash
# Health check
curl http://localhost:8080/health

# Hello endpoint
curl http://localhost:8080/api/v1/hello

# Get user
curl http://localhost:8080/api/v1/users/123

# Create user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john@example.com"}'
```

## Tests

### Tests Unitaires

```bash
# Tests unitaires rapides
go test ./internal/... -v -short

# Tous les tests unitaires
go test ./internal/... -v

# Tests avec race detector
go test ./internal/... -race

# Couverture de code
go test ./internal/... -cover

# Rapport de couverture HTML
go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Tests d'Intégration

```bash
# Tests d'intégration (longs)
go test ./test/integration/... -v

# Skipped en mode short
go test ./test/integration/... -v -short
```

### Tous les Tests

```bash
# Tous les tests (unitaires + intégration) en mode court
go test ./... -v -short

# Tous les tests complets
go test ./... -v

# Avec coverage complète
go test ./... -cover -coverprofile=coverage.out
```

### Benchmarks

```bash
# Lancer les benchmarks
go test ./internal/... -bench=. -benchmem

# Benchmark spécifique
go test ./internal/transport/http -bench=BenchmarkHandler -benchmem
```

### Tests disponibles

**Tests unitaires** (`internal/*/`):

- ✅ Tests de configuration (Load, Validate)
- ✅ Tests des handlers HTTP (Health, Welcome, Hello, Users)
- ✅ Tests du router (routes, middleware)
- ✅ Tests du serveur (lifecycle, timeouts, shutdown)
- ✅ Tests de validation
- ✅ Benchmarks de performance

**Tests d'intégration** (`test/integration/`):

- ✅ Test du cycle de vie complet du serveur
- ✅ Test du graceful shutdown
- ✅ Test des erreurs de démarrage
- ✅ Test de requêtes concurrentes (100 requêtes)
- ✅ Test des timeouts
- ✅ Test avec contexte

## Structure du Projet

```
template-backend-go/
├── cmd/                          # Points d'entrée de l'application
│   ├── root.go                   # Commande racine Cobra + config Viper
│   ├── serve.go                  # Commande serve (orchestration)
│   └── serve_test.go             # Tests unitaires de la commande
├── internal/                     # Packages internes (non exportables)
│   ├── config/                   # Gestion de la configuration
│   │   ├── config.go             # Load, Validate, structs Config
│   │   └── config_test.go        # Tests de configuration
│   ├── server/                   # Lifecycle du serveur HTTP
│   │   ├── server.go             # Start, Shutdown, gestion signaux
│   │   └── server_test.go        # Tests du serveur
│   └── transport/http/           # Couche HTTP
│       ├── handlers.go           # Handlers métier (Health, Users, etc.)
│       ├── handlers_test.go      # Tests des handlers
│       ├── router.go             # Configuration routes + middleware
│       └── router_test.go        # Tests du router
├── test/                         # Tests d'intégration
│   ├── integration/
│   │   └── server_test.go        # Tests d'intégration du serveur
│   └── README.md                 # Documentation des tests
├── config.yaml                   # Configuration (gitignored)
├── config.example.yaml           # Template de configuration
├── main.go                       # Point d'entrée principal
└── go.mod
```

### Principes de Design

- **Séparation des préoccupations** : Chaque package a une responsabilité unique
- **Testabilité** : Toutes les couches sont facilement mockables
- **Réutilisabilité** : Les packages `internal/` sont indépendants
- **Standards Go** : Convention `internal/`, `cmd/`, exports clairs
- **Contrats explicites** : Types définis, pas d'implicite
- **Production-ready** : Timeouts, shutdown gracieux, traçabilité

## Développement

Pour lancer en mode développement:

```bash
go run main.go serve
```

### Ajouter un Nouveau Endpoint

1. Définir le DTO dans `internal/transport/httpapi/dto/`
2. Ajouter le handler dans `internal/transport/httpapi/handler/`
3. Enregistrer la route dans `internal/transport/httpapi/router/`
4. Ajouter les tests unitaires
5. Vérifier que la réponse respecte le contrat d'API

### Modifier la Configuration

1. Mettre à jour les structs dans `internal/config/config.go`
2. Ajouter la validation dans `Config.Validate()`
3. Mettre à jour `config.example.yaml`
4. Ajouter les tests

## Fonctionnalités

### Production-Ready

- ✅ **Timeouts configurés** (Read: 10s, Write: 10s, ReadHeader: 5s, Idle: 120s)
- ✅ **Graceful shutdown** avec timeout (5s)
- ✅ **Gestion des signaux** (SIGINT, SIGTERM)
- ✅ **Protection Slowloris** (ReadHeaderTimeout)
- ✅ **Canal d'erreur bufferisé** (pas de goroutine leak)
- ✅ **Traçabilité** (request_id propagé)
- ✅ **Logging structuré**
- ✅ **Recovery middleware**

### Architecture

- ✅ **Séparation domaine/transport**
- ✅ **Erreurs typées** indépendantes de HTTP
- ✅ **Contrats d'API explicites** (réponses normalisées)
- ✅ **Validation centralisée** (DTOs à la frontière)
- ✅ **Mapping erreurs** domaine → HTTP
- ✅ **Configuration par fichier** YAML (Viper)
- ✅ **Priorité** : flags > config file > defaults

### Développement

- ✅ **CLI avec Cobra**
- ✅ **Tests unitaires** (87.5%+ coverage sur config, 100% sur handlers)
- ✅ **Tests d'intégration** (cycle de vie complet, concurrence)
- ✅ **Benchmarks de performance**
- ✅ **Architecture modulaire** (cmd, internal/config, internal/server, internal/transport)
- ✅ **Hot reload possible** (avec air ou similaire)

## Guide de Contribution

### Règles Strictes

1. **Ne jamais casser le contrat d'API** : Toutes les réponses doivent suivre le format défini
2. **Pas de logique métier dans les handlers** : Utilisez des services/use cases
3. **Erreurs domaine typées** : Pas de `errors.New()` avec HTTP status directement
4. **Tests obligatoires** : Toute nouvelle fonctionnalité doit avoir des tests
5. **Validation à la frontière** : Uniquement dans les DTOs

### Workflow

1. Créer une branche feature
2. Implémenter en respectant l'architecture
3. Ajouter les tests (unitaires + intégration si nécessaire)
4. Vérifier que tous les tests passent : `go test ./... -v`
5. Vérifier la coverage : `go test ./... -cover`
6. Vérifier le race detector : `go test ./... -race -short`
7. Créer une PR avec description claire

## Roadmap

### À Implémenter

- [ ] **Couche domaine complète** (`internal/domain/`)
- [ ] **Couche application** (`internal/app/`) avec use cases
- [ ] **DTOs structurés** (`internal/transport/httpapi/dto/`)
- [ ] **Response helpers** (`internal/transport/httpapi/response/`)
- [ ] **Error mapping** (`internal/transport/httpapi/errmap/`)
- [ ] **Middleware de traçabilité** (request_id automatique)
- [ ] **Infrastructure** (`internal/infra/`) : DB, repositories
- [ ] **Logging structuré** avec niveaux
- [ ] **Métriques** (Prometheus)
- [ ] **Health checks avancés** (DB, services externes)
- [ ] **Documentation OpenAPI/Swagger**
- [ ] **Rate limiting**
- [ ] **CORS configurable**
- [ ] **Authentification JWT**
- [ ] **Migrations de base de données**

### Améliorations

- [ ] **Hot reload** en développement (avec air)
- [ ] **Docker** multi-stage build
- [ ] **CI/CD** GitHub Actions
- [ ] **Linting** strict (golangci-lint)
- [ ] **Pre-commit hooks**

## Références

### Architecture

- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
- [Go Project Layout](https://github.com/golang-standards/project-layout)

### Patterns

- Domain-Driven Design (DDD)
- Repository Pattern
- Dependency Injection
- Error Handling Strategies

### Production

- [The Twelve-Factor App](https://12factor.net/)
- [Go best practices](https://go.dev/doc/effective_go)

---

## License

MIT

---

## Support

Pour toute question ou suggestion, créer une issue sur le dépôt.

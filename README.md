# Template Backend Go

Une application CLI en Go avec un serveur HTTP utilisant Cobra et Gin.

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
```

### Options disponibles

- `-p, --port`: Port d'écoute (défaut: 8080)
- `-H, --host`: Host à binder (défaut: 0.0.0.0)

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

### Lancer les tests unitaires

```bash
# Tests rapides (sans intégration)
go test ./cmd -v -short

# Tous les tests
go test ./cmd -v

# Tests avec race detector
go test ./cmd -race -short

# Couverture de code
go test ./cmd -cover -short

# Rapport de couverture HTML
go test ./cmd -coverprofile=coverage.out -short
go tool cover -html=coverage.out
```

### Benchmarks

```bash
# Lancer les benchmarks
go test ./cmd -bench=. -benchmem

# Benchmark spécifique
go test ./cmd -bench=BenchmarkHealthEndpoint -benchmem
```

### Tests disponibles

**Tests unitaires** (`serve_test.go`):

- ✅ Tests des handlers HTTP (GET, POST)
- ✅ Tests de validation des entrées
- ✅ Tests des cas d'erreur (JSON invalide, 404)
- ✅ Tests des méthodes HTTP non autorisées
- ✅ Benchmarks de performance

**Tests d'intégration** (`integration_test.go`):

- ✅ Test du cycle de vie complet du serveur
- ✅ Test du graceful shutdown
- ✅ Test des erreurs de démarrage
- ✅ Test de requêtes concurrentes
- ✅ Test des timeouts

## Développement

Pour lancer en mode développement:

```bash
go run main.go serve
```

## Fonctionnalités

- ✅ CLI avec Cobra
- ✅ Serveur HTTP avec Gin
- ✅ Graceful shutdown
- ✅ Middleware de logging et recovery
- ✅ Gestion des erreurs
- ✅ Timeouts configurés (Read, Write, ReadHeader, Idle)
- ✅ Routes groupées avec versioning API
- ✅ Tests unitaires et d'intégration
- ✅ Benchmarks de performance
- ✅ Protection contre Slowloris (ReadHeaderTimeout)
- ✅ Gestion propre des signaux (SIGINT, SIGTERM)
- ✅ Canal d'erreur bufferisé (pas de goroutine leak)

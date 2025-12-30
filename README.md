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
- ✅ Configuration par fichier YAML (Viper)
- ✅ Priorité : flags > config file > defaults
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

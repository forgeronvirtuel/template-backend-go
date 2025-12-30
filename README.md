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

## Architecture

Le projet suit une architecture modulaire clean avec séparation des responsabilités :

```
template-backend-go/
├── cmd/                          # Points d'entrée de l'application
│   ├── root.go                   # Commande racine Cobra + config Viper
│   ├── serve.go                  # Commande serve (orchestration)
│   ├── serve_test.go             # Tests unitaires des endpoints
│   └── integration_test.go       # Tests d'intégration serveur
├── internal/                     # Packages internes (non exportables)
│   ├── config/                   # Gestion de la configuration
│   │   └── config.go             # Load, Validate, structs Config
│   ├── server/                   # Lifecycle du serveur HTTP
│   │   └── server.go             # Start, Shutdown, gestion signaux
│   └── transport/http/           # Couche HTTP
│       ├── handlers.go           # Handlers métier (Health, Users, etc.)
│       └── router.go             # Configuration routes + middleware
├── config.yaml                   # Configuration (gitignored)
├── config.example.yaml           # Template de configuration
├── main.go                       # Point d'entrée principal
└── go.mod
```

### Principes de design

- **Séparation des préoccupations** : Chaque package a une responsabilité unique
- **Testabilité** : Toutes les couches sont facilement mockables
- **Réutilisabilité** : Les packages `internal/` sont indépendants
- **Standards Go** : Convention `internal/`, `cmd/`, exports clairs

## Développement

Pour lancer en mode développement:

```bash
go run main.go serve
```

### Ajouter un nouveau endpoint

1. Ajouter le handler dans `internal/transport/http/handlers.go`
2. Enregistrer la route dans `internal/transport/http/router.go`
3. Ajouter les tests dans `cmd/serve_test.go`

### Modifier la configuration

1. Mettre à jour les structs dans `internal/config/config.go`
2. Ajouter la validation dans `Config.Validate()`
3. Mettre à jour `config.example.yaml`

## Fonctionnalités

- ✅ CLI avec Cobra
- ✅ Serveur HTTP avec Gin
- ✅ Architecture modulaire (cmd, internal/config, internal/server, internal/transport)
- ✅ Configuration par fichier YAML (Viper)
- ✅ Priorité : flags > config file > defaults
- ✅ Validation de configuration
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

## Structure du code

### cmd/serve.go - Orchestration

```go
func runServe() error {
    cfg := config.Load()           // Charge la configuration
    handler := http.NewHandler()   // Crée les handlers métier
    router := http.NewRouter()     // Configure le router
    srv := server.New()            // Crée le serveur
    return srv.Start()             // Démarre avec lifecycle
}
```

### internal/config - Configuration

- `Load()` : Charge depuis Viper (flags > file > defaults)
- `Validate()` : Valide les valeurs (port, host)
- Types `Config` et `ServerConfig` avec mapstructure

### internal/transport/http - Couche HTTP

- `handlers.go` : Logique métier (Health, Users, etc.)
- `router.go` : Configuration des routes et middleware

### internal/server - Lifecycle serveur

- `New()` : Crée http.Server avec timeouts
- `Start()` : Démarre + gestion des signaux
- `Shutdown()` : Arrêt graceful avec timeout

# Tests

Ce dossier contient les tests d'intégration pour le projet.

## Structure

```
test/
└── integration/
    └── server_test.go    # Tests d'intégration du serveur HTTP
```

## Tests d'intégration

Les tests d'intégration vérifient le comportement complet du système, incluant :

- **Cycle de vie du serveur** : Démarrage, exécution, et arrêt gracieux
- **Endpoints HTTP** : Vérification des routes et réponses
- **Gestion des signaux** : SIGTERM, SIGINT pour l'arrêt gracieux
- **Requêtes concurrentes** : Test de charge avec 100 requêtes simultanées
- **Timeouts du serveur** : Vérification des configurations de timeout
- **Erreurs de démarrage** : Tests avec configurations invalides

## Exécution des tests

### Tests d'intégration uniquement

```bash
# Mode court (skip les tests d'intégration)
go test ./test/integration/... -v -short

# Exécution complète des tests d'intégration
go test ./test/integration/... -v

# Avec le race detector
go test ./test/integration/... -v -race
```

### Tous les tests (unitaires + intégration)

```bash
# Tous les tests en mode court
go test ./... -v -short

# Tous les tests (long)
go test ./... -v

# Avec coverage
go test ./... -cover -coverprofile=coverage.out
```

## Notes

- Les tests d'intégration sont automatiquement skippés en mode `-short`
- Chaque test utilise un port dynamique pour éviter les conflits
- Les tests vérifient l'arrêt gracieux avec un timeout de 5 secondes
- Les tests de concurrence envoient 100 requêtes simultanées

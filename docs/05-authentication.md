# 05 - Authentification

But : documenter le système d'authentification optionnel ajouté au template,
ses décisions de conception et son utilisation.

## Vue d'ensemble

Le template fournit une capacité d'authentification **optionnelle** et
**vendor-neutral** qui peut être activée via configuration. Par défaut,
l'authentification est désactivée mais applique une politique stricte
(strict mode) : les routes protégées retournent HTTP 401 même si l'auth
est désactivée.

## Principes de conception

### Vendor-neutral
- Aucune dépendance externe (stdlib uniquement).
- L'interface `auth.Authenticator` permet de brancher n'importe quelle
  implémentation (JWT, OAuth2, SAML, etc.) sans modifier le template.
- Le template fournit deux implémentations de base :
  - `DisabledAuthenticator` (strict mode : toujours 401)
  - `APIKeyAuthenticator` (validation de clés API simples)

### Strict mode (sécurité par défaut)
- Si `auth.enabled=false` : les routes protégées retournent 401.
- Si `auth.enabled=true` mais aucune clé configurée : retour 401.
- Rationale : éviter les déploiements accidentels sans auth sur des
  routes supposées protégées.

### Séparation endpoints opérationnels / business
- `/live` et `/ready` restent publics (pas de middleware auth).
- Seules les routes business sous `/api/v1` peuvent être protégées.
- Rationale : les orchestrateurs doivent toujours pouvoir vérifier
  liveness/readiness sans credentials.

### Intégration au contrat API
- Les erreurs d'authentification utilisent `response.Fail` et retournent
  le format standard `error/meta` avec `request_id`.
- Codes d'erreur : `UNAUTHORIZED` (401), `FORBIDDEN` (403).
- Rationale : cohérence avec le reste de l'API métier.

## Architecture

### Package `internal/security/auth`

**Interface `Authenticator`**
```go
type Authenticator interface {
    Authenticate(ctx context.Context, r *http.Request) (Identity, error)
}
```

**Type `Identity`**
```go
type Identity struct {
    Subject string   // Identifiant (ex: "api_key", user ID)
    Scopes  []string // Permissions optionnelles
}
```

**Erreurs sentinelles**
- `ErrUnauthenticated` : credentials manquantes ou invalides → 401
- `ErrForbidden` : credentials valides mais permissions insuffisantes → 403

**Context helpers**
- `ContextWithIdentity(ctx, identity)` : attacher l'identité au contexte
- `IdentityFromContext(ctx)` : extraire l'identité du contexte

### Implémentations fournies

**`DisabledAuthenticator`**
- Retourne toujours `ErrUnauthenticated`.
- Utilisé quand `auth.enabled=false` ou si aucune clé n'est configurée.
- Garantit le strict mode.

**`APIKeyAuthenticator`**
- Valide des clés API statiques depuis :
  - Header `Authorization: Bearer <key>`
  - Header `X-API-Key: <key>`
- Ne logge JAMAIS les clés (sécurité).
- Retourne `Identity{Subject: "api_key"}` en cas de succès.
- Limitation : clés statiques, pas de rotation. Pour production,
  envisager JWT ou OAuth2.

### Middleware `middleware.Auth`

Localisation : `internal/transport/http/middleware/auth.go`

Comportement :
1. Appelle `authenticator.Authenticate(ctx, request)`.
2. Succès :
   - Attache `Identity` au contexte via `ContextWithIdentity`.
   - Appelle `c.Next()` pour continuer le traitement.
3. Échec :
   - Mappe l'erreur vers un code d'erreur business (`UNAUTHORIZED` / `FORBIDDEN`).
   - Retourne une réponse JSON via `response.Fail` (format `error/meta`).
   - Appelle `c.Abort()` pour terminer la chaîne de traitement.

Le middleware panic si `authenticator` est `nil` (fail-fast).

### Intégration au router

Fichier : `internal/transport/http/router/router.go`

Structure :
```go
type Deps struct {
    Users   *handler.UsersHandler
    Health  *handler.HealthHandler
    Metrics metrics.Recorder
    Auth    auth.Authenticator // Optionnel
}
```

Si `Deps.Auth` est `nil`, le router le remplace automatiquement par
`auth.NewDisabledAuthenticator()` pour garantir le strict mode.

Routes publiques (sans auth) :
- `GET /live`
- `GET /ready`

Routes protégées (avec auth) :
- Groupe `/api/v1` → sous-groupe protégé avec `middleware.Auth`
- Exemple : `POST /api/v1/users`

## Configuration

### Clés Viper

**`auth.enabled`** (bool, default `false`)
- Active ou désactive l'authentification.
- Si `false` : `DisabledAuthenticator` est utilisé (strict 401).

**`auth.api_keys`** (string ou []string)
- Liste des clés API autorisées.
- Supporte deux formats :
  - Liste YAML : `["key1", "key2"]`
  - String comma-separated : `"key1,key2"`
- Si vide alors que `enabled=true` : strict 401.

### Exemple de configuration

Fichier `config.yaml` :
```yaml
auth:
  enabled: true
  api_keys:
    - "secret-key-1"
    - "secret-key-2"
```

Ou via variables d'environnement :
```bash
AUTH_ENABLED=true
AUTH_API_KEYS=secret-key-1,secret-key-2
```

### Avertissement sécurité
- Ne jamais committer de vraies clés API dans le dépôt.
- Utiliser des secrets managers en production (Vault, AWS Secrets Manager, etc.).
- Préférer des systèmes robustes (JWT, OAuth2) pour les APIs publiques.

## Utilisation pour les développeurs

### Accéder à l'identité authentifiée

Dans un handler :
```go
func (h *MyHandler) MyEndpoint(c *gin.Context) {
    identity, ok := auth.IdentityFromContext(c.Request.Context())
    if !ok {
        // Ne devrait pas arriver si le middleware Auth est actif
        response.Fail(c, 500, requestID, "INTERNAL_ERROR", "Identity not found", nil)
        return
    }
    
    logger.Info("Request authenticated",
        slog.String("subject", identity.Subject),
    )
    
    // Traiter la requête...
}
```

### Implémenter un authenticator personnalisé

Exemple : validation JWT
```go
type JWTAuthenticator struct {
    verifier jwt.Verifier
}

func (j *JWTAuthenticator) Authenticate(ctx context.Context, r *http.Request) (auth.Identity, error) {
    tokenString := extractBearerToken(r)
    if tokenString == "" {
        return auth.Identity{}, auth.ErrUnauthenticated
    }
    
    claims, err := j.verifier.Verify(tokenString)
    if err != nil {
        return auth.Identity{}, auth.ErrUnauthenticated
    }
    
    return auth.Identity{
        Subject: claims.Subject,
        Scopes:  claims.Scopes,
    }, nil
}
```

Puis wirer dans `cmd/serve.go` :
```go
authenticator := &JWTAuthenticator{verifier: jwtVerifier}
deps := router.Deps{
    Auth: authenticator,
    // ...
}
```

### Désactiver l'auth pour certaines routes

Si vous souhaitez des routes publiques sous `/api/v1`, créez un groupe
sans middleware auth :
```go
v1 := r.Group("/api/v1")

// Routes publiques
v1.GET("/public", publicHandler)

// Routes protégées
protected := v1.Group("")
protected.Use(middleware.Auth(authenticator))
protected.POST("/users", usersHandler)
```

## Tests

### Tests unitaires

Fichiers :
- `internal/security/auth/auth_test.go` : tests des authenticators
- `internal/transport/http/middleware/auth_test.go` : tests du middleware

Exécution :
```bash
go test ./internal/security/auth/... -v
go test ./internal/transport/http/middleware/... -v
```

### Tests d'intégration

Tester avec curl (auth désactivée) :
```bash
# /live et /ready restent accessibles
curl http://localhost:8080/live
curl http://localhost:8080/ready

# Route protégée retourne 401 (strict mode)
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","name":"Test"}'
# Retourne: {"error":{"code":"UNAUTHORIZED",...},"meta":{"request_id":"..."}}
```

Tester avec auth activée (`config.yaml`) :
```yaml
auth:
  enabled: true
  api_keys:
    - "test-key-123"
```

```bash
# Sans auth : 401
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","name":"Test"}'

# Avec auth valide : 200/201 (si UsersHandler est branché)
curl -X POST http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer test-key-123" \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","name":"Test"}'

# Ou avec X-API-Key
curl -X POST http://localhost:8080/api/v1/users \
  -H "X-API-Key: test-key-123" \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","name":"Test"}'
```

## Limitations et considérations

### Limitations de `APIKeyAuthenticator`
- Clés statiques (pas de rotation).
- Pas d'expiration.
- Pas de rate limiting intégré.
- Pas de gestion granulaire des permissions (scopes vides).

Pour production, envisager :
- JWT avec expiration et refresh tokens.
- OAuth2 / OIDC pour délégation.
- mTLS pour service-to-service.

### Sécurité

**Ce qui est fait :**
- Strict mode par défaut.
- Les clés ne sont JAMAIS loggées.
- Erreurs d'auth loggées au niveau WARN (pas ERROR pour éviter le spam).
- Format de réponse cohérent avec le contrat API.

**Ce qui n'est PAS fait (à ajouter selon besoins) :**
- Rate limiting par IP/clé.
- Audit des tentatives d'authentification échouées.
- Blocage automatique après X échecs.
- Rotation automatique des clés.

### Performance
- `APIKeyAuthenticator` utilise une map pour O(1) lookup.
- Pas d'allocation supplémentaire si contexte déjà existant.
- Middleware auth positionné après RequestID et StructuredLogger pour
  garantir la traçabilité même des requêtes rejetées.

## Décisions de conception (récapitulatif)

### Pourquoi vendor-neutral ?
- Permet aux projets de choisir leur stratégie d'auth sans réécrire
  le template.

### Pourquoi strict mode par défaut ?
- Évite les déploiements accidentels sans protection.
- Principe de sécurité : deny by default.

### Pourquoi stdlib-only ?
- Pas de dépendances externes pour une fonctionnalité optionnelle.
- Le projet peut ajouter JWT/etc. selon ses besoins.

### Pourquoi ne pas protéger /live et /ready ?
- Conformité avec les attentes des orchestrateurs Kubernetes/Docker.
- Les health checks doivent être accessibles sans credentials.

### Pourquoi logger au niveau WARN et pas ERROR ?
- Les tentatives d'auth échouées sont attendues (bots, typos).
- Évite de polluer les logs d'erreurs critiques.
- Permet quand même de tracer les tentatives pour audit.

## Roadmap future (suggestions)

- [ ] Ajouter un exemple d'implémentation JWT dans `examples/auth/jwt/`.
- [ ] Ajouter un exemple OAuth2 dans `examples/auth/oauth2/`.
- [ ] Documenter l'intégration avec des secret managers.
- [ ] Ajouter un middleware de rate limiting optionnel.
- [ ] Support des scopes/permissions dans `APIKeyAuthenticator`.

## Références

- Code : `internal/security/auth/auth.go`
- Middleware : `internal/transport/http/middleware/auth.go`
- Router : `internal/transport/http/router/router.go`
- Configuration : `cmd/serve.go` (fonction `configureAuth`)
- Tests : `internal/security/auth/auth_test.go`, `internal/transport/http/middleware/auth_test.go`

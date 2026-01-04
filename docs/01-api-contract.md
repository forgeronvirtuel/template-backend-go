# 01 - Contrat API

But : documenter les choix de contrat pour les APIs métiers présents
dans le template et les décisions explicites liées au transport.

Pourquoi définir le contrat avant l'architecture

- Un contrat HTTP/JSON stable (format success/erreur) facilite la
  séparation des responsabilités entre transport et domaine : le
  transport sait comment sérialiser/désérialiser et mapper les erreurs,
  tandis que le domaine reste agnostique du transport.

Formats exacts utilisés (extraits de `internal/transport/http/response`)

Succès (exemple)

```json
{
  "data": { "id": "123", "email": "a@b.com" },
  "meta": { "request_id": "e4f1c8b9..." }
}
```

Erreur (exemple)

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request payload",
    "details": [
      {
        "field": "email",
        "rule": "email",
        "message": "must be a valid email address"
      }
    ]
  },
  "meta": { "request_id": "e4f1c8b9..." }
}
```

Remarques :

- Les structures `Success` et `Failure` sont définies dans
  `internal/transport/http/response/response.go` et utilisées par les
  helpers `OK(...)` et `Fail(...)`.
- Le champ `meta.request_id` est toujours présent dans les réponses
  métiers (transport boundary responsibility).

Endpoints opérationnels en dehors du contrat

- Important : `/live` et `/ready` sont explicitement hors du contrat
  métier. Ils renvoient des JSON opérationnels minimaux (voir
  `internal/transport/http/handler/health.go`) et ne sont pas enveloppés
  dans `data/meta` ou `error/meta`.

Stratégie d'erreur

- Les erreurs du domaine et de l'application restent des erreurs Go
  typées/structurées et sont mappées au transport via `errmap` (voir
  `internal/transport/http/errmap`). Le transport effectue :
  - mapping erreur -> (http status, code, message)
  - émission d'une réponse formatée via `response.Fail(...)`.

Validation

- La validation est effectuée à la frontière du transport (ex : binding
  dans les handlers). Le package `validation` convertit les erreurs de
  binding/validator en `response.ErrorDetail`.
- Les DTOs (ex. `CreateUserRequest`) sont validés au niveau du handler
  (transport boundary). Le domaine ne dépend pas des tags JSON/binding.

`request_id` : rôle et propagation

- Header : middleware lit/écrit `X-Request-Id` (see
  `internal/transport/http/middleware/request_id.go`).
- Corps de réponse : le transport inclut `meta.request_id` via
  `response.OK` / `response.Fail`.
- Raison : permet la corrélation entre logs, traces et réponses HTTP ;
  facilite le diagnostic côté client et côté serveur.

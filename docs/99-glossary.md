# 99 - Glossaire

Operational endpoint

- Endpoint réseau utilisé par les orchestrateurs ou systèmes d'observabilité
  pour vérifier l'état opérationnel d'un service (ex. `/live`, `/ready`).

API contract

- Convention JSON/HTTP partagée entre le transport et les consommateurs
  de l'API pour formater les réponses de succès et d'erreur (ici `data/meta`
  et `error/meta`).

Domain error

- Erreur issue de la logique métier ou des règles de l'application. Doit
  être mappée au transport par une couche dédiée (`errmap`).

Transport layer

- Couche responsable de l'interface réseau (HTTP), sérialisation,
  validation des DTOs et mapping des erreurs vers des réponses HTTP.

Readiness vs Liveness

- Liveness : indique que le processus est vivant (`/live`), utile pour
  détecter des processus gelés.
- Readiness : indique si le service est prêt à traiter du trafic
  (dépendances externes disponibles). Ici `/ready` exécute des checks via
  `internal/health.Aggregator`.

Structured logging

- Logs émis au format structuré (JSON) avec paires clé/valeur, facile
  à ingérer et à filtrer par les systèmes d'observabilité.

Metrics cardinality

- Cardinalité des labels désigne le nombre de valeurs distinctes pour
  un label. Faible cardinalité est recherchée pour éviter l'explosion
  de séries métriques (exclure `request_id`, `user_id`, IP, etc.).

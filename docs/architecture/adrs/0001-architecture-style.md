# Architecture Decision Record (ADR)

## ADR-0001: Architecture Style - Modular Monolith to Microservices

### Status
Accepted

### Context
Le projet CanBankX demande de démontrer les avantages et les performances d'une architecture orientée microservices (avec API Gateway, Load Balancing et Caching) par rapport à une approche plus classique (monolithique).

### Decision
Nous adopterons une approche "Monolithe Modulaire" (**Modular Monolith**) pour la Phase 1. 
Dans ce modèle :
1. Chaque "Bounded Context" (Identité, Comptes, Paiements) est implémenté dans un *package Go séparé*.
2. Il n'y a qu'un **seul et unique point d'entrée (`main.go`)** hébergeant toutes les API REST.
3. Toutes les interactions entre les domaines se font via des appels de fonctions ordinaires (interfaces Go) en mémoire, et non via le réseau.

Le système sera ensuite transformé par étapes :
- Étape 2 : Ajout de Load Balancing horizontal (Nginx) et de Caching (Redis) au monolithe.
- Étape 3 : Découpage physique en vrais microservices orchestrés via API Gateway (Kong).

### Consequences
**Positives** :
- Facilité d'établir une "ligne de base" (baseline) pour les tests de latence et de débit (P95, etc.).

**Négatives** :
- Risque de couplage fort si les règles de découpage (Modular Monolith) ne sont pas respectées.

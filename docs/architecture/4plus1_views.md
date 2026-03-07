# Vues 4+1 (Modèle de Kruchten) - CanBankX

Le modèle 4+1 décrit l'architecture selon cinq perspectives distinctes pour répondre aux besoins de différents intervenants.

---

## 1. Vue Logique (Logical View)
La solution repose sur un **Monolithe Modulaire** partitionné par métier.

![Logical View Diagram](puml/logical_view.puml)
- **Module Identity :** `User`, `Role`, `Session`, `KYCDocument`.
- **Module Accounts :** `Account`, `AccountBalance`.
- **Module Transfers :** `Transfer`, `Payee`.
- **Module Ledger :** `LedgerEntry`, `Transaction`.

Ces modules communiquent via des interfaces Go strictement définies (`ports`) sans partager directement leurs modèles de base de données.

---

## 2. Vue Processus (Process View)

**Workflow (Paiement) :**

![Process View Diagram](puml/process_view.puml)
1. Requête HTTP entrante.
2. Le **Router (Gin)** délègue au Middleware d'Authentification (validation JWT).
3. Le **Contrôleur (Handler)** vérifie l'`Idempotency-Key` via le module *Transfers*.
4. Si la clé est nouvelle, le module *Transfers* orchestre :
   - Requête de verrouillage pessimiste sur les comptes source et destination via le module *Accounts*.
   - Écriture d'une `Transaction` et de deux `LedgerEntry` dans la base via le module *Ledger*.
5. Commit de la transaction SQL globale.
6. Retour HTTP `201 Created`.

*Note : Sous charge (k6), la base de données MySQL applique des verrous de lignes (`FOR UPDATE`) sérialisant l'accès au même compte, affectant potentiellement le RPS sur les comptes "hot spot".*

---

## 3. Vue Développement (Development/Implementation View)

## 4. Vue Déploiement (Physical/Deployment View)
*Dédiée aux administrateurs systèmes et ingénieurs DevOps (Phase 1 Baseline).*

**Actuellement (Phase 1) :**

![Deployment View Diagram Phase 1](puml/deployment_view_phase1.puml)
- Node 1: Serveur applicatif exécutant le binaire Go (Monolithe).
- Node 2: Base de données MySQL.

**Évolution prévue (Phase 2) :**

![Deployment View Diagram Phase 2](puml/deployment_view_phase2.puml)
- **Load Balancer :** Nginx exposant l'application.
- **App Nodes :** Instances de conteneurs Docker exécutant l'application Go.
- **Cache Node :** Redis pour la mise en cache.
- **Monitoring :** Prometheus / Grafana scrutant le `/metrics` des instances.

**Évolution prévue (Phase 3) :**

![Deployment View Diagram Phase 3](puml/deployment_view_phase3.puml)
- **API Gateway :** KrakenD routant les demandes.
- **Services :** Microservices backend Go.
- **Bases de données :** Une base de données isolée par service.

---

## +1. Vue Scénarios (Use Cases View)

**Scénario de preuve (Virement) :**
Le client s'authentifie, obtient un JWT, puis appelle `POST /api/v1/transfers`.
Le monolithe valide le JWT (Vue Logique), exécute la logique de paiement avec idempotence (Vue Processus), interagit avec le domaine *Transfers* défini en interne (Vue Développement), et stocke les données dans le conteneur MySQL (Vue Déploiement).

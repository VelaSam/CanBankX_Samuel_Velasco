# arc42 Modèle d'Architecture - CanBankX

Source: 
https://github.com/NetworkedAssets/arc42-in-markdown-template/blob/master/2.%20Single%20File/Arc42%20Template%20in%20Markdown.md

## 1. Introduction and Goals

### 1.1 Requirements Overview
Le projet **CanBankX** vise à concevoir une plateforme bancaire commerciale canadienne numérique et sécurisée. Le système permet la gestion de comptes chèques et épargne, l'exécution de paiements et transferts, la consultation de soldes, ainsi qu'une conformité stricte (AML, KYC) et des clôtures journalières (EOD). 

Le projet est divisé en phases évolutives :
1. Baseline Monolithe Modulaire (Archivé).
2. Optimisation via Caching (Redis) & Load Balancing (Archivé).
3. Architecture Microservices orchestrée via une API Gateway KrakenD **(Version Courante sur main)**.

### 1.2 Quality Goals
1. **Performance (Latence/Débit):** Latence P95 < 250-500ms ; RPS > 600-1000.
2. **Intégrité et Auditabilité:** Écritures "exactly-once" et logs de transactions "append-only".
3. **Observabilité:** 4 Golden Signals monitorés avec Prometheus/Grafana.
4. **Disponibilité:** Supérieure à 95% (simulée).

### 1.3 Stakeholders
- **Clients (Particuliers/PME):** Utilisent l'API (UX fluide, fiable).
- **Évaluateurs (Professeur LOG430):** Juge de la traçabilité architecturale et des mesures quantitatives NFR.
- **Développeur:** Maintiennent un code propre en Go.

---

## 2. Architecture Constraints
- **Technologiques:** Go (Backend), MySQL (Persistance).
- **Organisationnelles:** Démontrer chaque étape architecturale (évolution) de façon chiffrée.

---

## 3. System Scope and Context

### 3.1 Business Context
- **Client:** Envoie des requêtes HTTPS REST au système bancaire.
- **CanBankX (Le Système):** Traite les inscriptions, virements, et logs.
- **Régulateurs/Partenaires (Simulés):** FINTRAC (vérifications de conformité), Interac (virements externes).

### 3.2 Technical Context
- Interface HTTPS REST sécurisée par Token (JWT).
- Base de données SQL hébergée localement / conteneurisée.
- Outils de métriques (Prometheus) *pullant* les données.

---

## 4. Solution Strategy
L'approche initiale reposait sur un **Monolithe Modulaire** écrit en Go. Cette décision (ADR-0001) a permis de bâtir une fondation saine. Ce monolithe a ensuite été optimisé par du Load-Balancing (Nginx) et de la mise en cache rapide avec Redis (Phase 2). Finalement, le système a été complètement découpé en une **Architecture Microservices** (Phase 3). Les bases de données ont été séparées (Database-per-Microservice) pour assurer une isolation parfaite des données (ADR-0002), et le trafic est désormais orchestré de manière transparente par l'API Gateway KrakenD. Les écritures critiques sont assurées par des mécanismes idempotents et des requêtes HTTP internes synchrones garantissant le respect de l'ACID.

---

## 5. Building Block View
Voir [Vues 4+1 (Vue Logique & Développement)](../4plus1_views.md) pour les détails. Le système est décomposé en *Bounded Contexts* :
- `Identity` (KYC, Auth)
- `Accounts` (Gestion des comptes)
- `Transfers` (Paiements intra et inter)
- `Ledger` (Traçabilité financière immuable)
- `Compliance` (Vérifications).

---

## 6. Runtime View
Les scénarios de fonctionnement principal (comme le virement) sont décrits dans la *Vue Processus* des [Vues 4+1](../4plus1_views.md).

---

## 7. Deployment View
Voir la [Vue Déploiement](../4plus1_views.md) des Vues 4+1 correspondantes aux différentes étapes architecturales.

---

## 8. Cross-cutting Concepts
1. **Idempotence:** Utilisée pour empêcher les virements multiples. L'appelant doit fournir un `Idempotency-Key` dans les en-têtes.
2. **Observabilité:** Tous les modules enregistrent leurs logs (structurés JSON) vers Stdout et exposent un endpoint `/metrics` en format Prometheus.
3. **Gestion d'Erreurs:** Format RFC 7807 (voir ADR-0003).

---

## 9. Design Decisions
Consolidées dans les ADRs dédiés :
- [ADR-0001: Architecture Style (Monolithe Modulaire)](../adrs/0001-architecture-style.md)
- [ADR-0002: Persistence Strategy & Transactionality](../adrs/0002-persistence-strategy.md)
- [ADR-0003: Error Handling & Versioning](../adrs/0003-error-handling-and-versioning.md)

---

## 10. Risques et Dettes Techniques
- **Technical Debt Intentionnelle (Résolue en Phase 2 et 3):** Les goulots d'étranglement de la base de données monolithique ont été résolus en séparant physiquement la base de données relationnelle en 3 conteneurs distincts (Identity, Accounts, Transactions).
- **Risque d'Orchestration:** L'état distribué des données oblige des appels HTTP internes synchrones entre les services lors d'un transfert (Ex: Transactions -> Accounts). Une défaillance réseau entre les deux services internes nécessiterait des mécanismes de compensation (Sagas) qui ne sont pas encore implémentés.

---

## 11. Glossary
Voir la documentation [Analyse Métier & DDD](../analyse_metier.md) pour le détail du vocabulaire du domaine d'affaires.

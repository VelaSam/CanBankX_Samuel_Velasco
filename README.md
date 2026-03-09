# CanBankX - Architecture Microservices (Phase 3 Finale)

Salut ! Bienvenue sur le dépôt de CanBankX. Honnêtement, c'est le plus gros projet que j'ai fait à date à l'ÉTS ! Petite anecdote : au début du lab, je pensais qu'il fallait *uniquement* faire la version microservices (donc sauter direct à la phase 3 en ignorant les phases 1 et 2). J'avais presque tout fini quand je me suis rendu compte de mon erreur et j'ai dû recommencer le projet à zéro... Mais bon, au final, c'était un mal pour un bien ! Vu que j'avais déjà construit l'architecture complexe, refaire la version monolithique de base pour ensuite l'upgrader vers cette version microservices finale, c'était un vrai piece of cake!

Cette branche contient l'aboutissement du projet : une **Architecture Microservices Répartie** complète, avec 3 APIs indépendantes en Go, 3 bases de données MySQL isolées (patron *Database-per-Microservice*), un cache Redis distribué et le Gateway KrakenD pour orchestrer.
---

## 🚀 Quickstart

### Prérequis
- `git`
- `docker` et `docker compose` installés localement ou sur votre VM.
- `Go 1.22+` installé sur votre machine hôte (nécessaire **uniquement** pour exécuter le script de génération de seed de la base de données).

### 1. Cloner le Projet
Assurez-vous d'être sur la branche contenant la version finale.

```bash
git clone https://github.com/VelaSam/CanBankX_Samuel_Velasco CanBankX
cd CanBankX
```

### 2. Démarrage de l'Infrastructure (Docker Compose)
Toute l'infrastructure (les 3 APIs Go, les 3 BDs MySQL, Redis, KrakenD, Prometheus et Grafana) est entièrement conteneurisée.
Aucun fichier `.env` n'est requis pour le lancement de base ; le fichier `docker-compose.yml` injecte les variable par défaut.

Exécutez la commande suivante à la racine du projet :
```bash
docker compose up -d --build
```
### 3. Peupler les Bases de Données (Data Seeder)
Pour que l'API Gateway et les tests de charge fonctionnent, il faut d'abord injecter des données dans les 3 bases de données isolées. J'ai conçu un script Go qui s'occupe de tout : création de 1000 utilisateurs, ouverture de leurs comptes, et dépôt initial de 10 000$ en exploitant la logique distribuée.

Exécutez la commande suivante depuis votre terminal hôte (à la racine du projet) :
```bash
go run cmd/seeder/main.go
```
*Note : Le script exportera automatiquement les identifiants de comptes requis par k6 vers le fichier `tests/k6/users.json`.*

### 4. Tests et Validation de l'API Gateway (KrakenD)

L'API Gateway KrakenD est maintenant le point d'entrée unique et sécurisé de tout le système sur le port `8080`. Les microservices internes (`8081`, `8082`, `8083`) ne sont plus exposer directement aux clients.

Vous pouvez tester le bon acheminement du système en appelant son point de terminaison de santé public :
```bash
curl http://localhost:8080/health
```
Si ça retourne "UP", la gateway effectue le routage correctement vers les APIs !

#### Validation détaillée via Postman :
Puisque le système est maintenant composé de multiples microservices. Vous pouvez importer la collection Postman mise à jour :
- **Fichier à importer dans Postman** : `docs/architecture/postman/postman-collection.json`
- Cette collection inclut des tests pour reproduire l'entièreté d'un parcours client : Inscription, Login, MFA, Création de compte et les virements ACID inter-services.

### 5. Tests de Charge Extrêmes (k6)
L'architecture microservicesa été conçue pour supporter une charge répartie élevée. 

Pour lancer la simulation de stress test (il y a 4 fichiers de stress test) via k6, exécutez le conteneur k6 officiel. Ce dernier lira vos différents scripts et le pool d'utilisateurs que le `seeder` a généré :

**Sur Linux / Mac / VM :**
```bash
docker run --rm -i -v $(pwd)/tests/k6:/scripts --network canbankx_canbankx-net grafana/k6 run -e API_URL=http://krakend:8080 /scripts/3_hard.js
```

**Sur Windows (PowerShell) :**
```powershell
docker run --rm -i -v ${PWD}/tests/k6:/scripts --network canbankx_canbankx-net grafana/k6 run -e API_URL=http://krakend:8080 /scripts/3_hard.js
```

### 6. Observabilité (Monitoring)
Les métriques Golden Signals de la passerelle et de chaque microservice peuvent être consultées :
- **Prometheus** : `http://localhost:9090`
- **Grafana** : `http://localhost:3000` *(Login : `admin` / Password : `admin`)*. Dans Grafana, connectez Prometheus dans le panneau *Data Sources* en ciblant l'URL interne Docker `http://prometheus:9090`.

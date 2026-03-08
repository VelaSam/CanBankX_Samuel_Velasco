# CanBankX

CanBankX (sur cette branche) est une API REST développée en Go en Monolithe Modulaire. Son architecture est hautement résiliente (via le lock ACID des tables) et traçable (via un système de comptabilité / *Ledger* en *append-only*).

## Quickstart

### Prérequis
- `git`
- `docker` et `docker-compose` d'installés localement (sur l'hôte ou une Machine Virtuelle).
- *(Facultatif : `Go 1.26+` si vous souhaitez développer ou compiler nativement).*

### 1. Cloner le Projet
Ce README est uniquement présent sur la branche `canbankx_phase1_monolithique`, il sera modifié pour les autres version
Clonez ce repo et aller sur la branche de la phase 1 :
```bash
cd CanBankX
git checkout canbankx_phase1_monolithique
```

### 2. Configuration (Fichier `.env`)
Le projet utilise Docker, mais ses containers ont besoin de valeurs spécifiques pour l'API et la bd. Un squelette vide est fourni, qu'il faut peupler :

**Sur Linux/Mac/WSL :**
```bash
cp .env-example .env
```
_Sur Windows (shame) : 
```bash
Copy-Item .env-example .env
```

Ouvrez le fichier `.env` nouvellement créé et modifiez les valeurs selon votre environnement :
- **`DB_USER`** et **`DB_PASSWORD`** : Les accès qui seront automatiquement configurés pour l'image MySQL `mysql:8.0`. Changez le mot de passe pour plus de sécurité.
- **`JWT_SECRET`** : Clé secrète de cryptographie servant à signer et valider de façon intègre les jetons. Il est crucial en prod de mettre une chaîne longue et aléatoire.
- **`GIN_MODE`** : Utilisez la valeur `debug` si vous testez localement pour voir les logs, ou `release` sur la vm

### 3. Démarrage de l'Application Bancaire
Toute l'infrastructure démarre instantanément avec Docker :
```bash
docker compose up -d --build
```
L'API devrait maintenant être en ligne sur : `http://localhost:8080/health`.

## Tests Unitaires et Fonctionnels (Postman)
Une collection complète est mise à disposition pour valider les 5 CU architecturaux.
Fichier à importer dans Postman : `docs/architecture/postman/postman-collection.json`

## Tests de Charge (k6) et Observabilité (Grafana)

- **Data Time-Series (Prometheus)** : `http://localhost:9090`
- **Dashboards Dynamiques (Grafana)**: `http://localhost:3000` *(Login : `admin` / Password : `admin`)*. Connectez Prometheus dans le panneau *Data Sources* avec l'URL interne de Docker `http://prometheus:9090`.

### Scénarios de tests disponibles
Le dossier `tests/k6` comprend 4 paliers d'évaluation de l'architecture :
1. `1_easy.js` : Modéré (50 Users). Devrait toujours réussir.
2. `2_medium.js` : Difficile (300 Users). Conçu pour stresser la phase monolithique et passer la Phase Redis+LB.
3. `3_hard.js` : Extrême (1000 Users) et difficile. Conçu pour stresser la phase Redis+LB et passer la Phase Microservices.
4. `4_massive_stress.js` : Crash Test Ultime.

### Lancer la simulation (sans installation logicielle locale)
Assurez-vous que l'API roule (`docker compose up`) et utilisez le conteneur k6 officiel pour simuler la charge par-dessus le réseau canbankx :

**Sur Windows (PowerShell) :**
```powershell
Get-Content tests/k6/1_monolith_baseline.js | docker run --rm -i --network canbankx_canbankx-net grafana/k6 run -e API_URL=http://api:8080 -
```

**Sur Linux / Mac / Serveur VM :**
```bash
cat tests/k6/1_monolith_baseline.js | docker run --rm -i --network canbankx_canbankx-net grafana/k6 run -e API_URL=http://api:8080 -
```
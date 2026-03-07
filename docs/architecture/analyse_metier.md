# Analyse Métier et DDD - CanBankX

## Périmètre Fonctionnel (Bounded Contexts)

### 1. Clients & Identité (Identity)
Gestion des profils clients, vérification KYC/AML, authentification forte (MFA) et rôles.

### 2. Comptes Bancaires (Accounts)
Gestion des comptes chèques et épargne, tenue des soldes sécurisés et application des intérêts.

### 3. Paiements & Transferts (Transfers)
Virements internes et paiements externes, incluant des contrôles de soldes et d'AML avant exécution.

### 4. Comptabilité & Règlements (Ledger)
Système central des écritures comptables, fournissant un journal immuable (append-only) et traitant la clôture journalière (EOD).

### 5. Conformité & Audit (Compliance)
Surveillance transactionnelle, archivage légal et détection des patterns anormaux (AML).

---

## Les 5 Cas d'Utilisation (UC) Principaux

Dans le cadre de cette architecture, nous nous concentrons sur les 5 Use Cases (UC) indispensables listés dans le cahier des charges :

![Diagramme des Cas d'Utilisation](puml/uc_diagram.puml)

### UC-01 : Inscription & Vérification d'identité (KYC)
- **Objectif :** Permettre à un client d’ouvrir un profil bancaire conforme aux exigences KYC/AML canadiennes.
- **Acteur principal :** Client
- **Flux principal :**
  1. Le Client fournit des données personnelles (nom, adresse, NAS).
  2. Le Système valide les données et crée un profil *Pending*.
  3. Le Client confirme via OTP/MFA.
  4. Le Système valide le KYC et active le profil.
- **Résultat attendu :** Un profil client actif.

### UC-02 : Authentification & MFA
- **Objectif :** Garantir un accès sécurisé aux services bancaires.
- **Acteur principal :** Client
- **Flux principal :**
  1. Saisie de l'identifiant et mot de passe.
  2. Vérification MFA.
  3. Création d’une session/jeton de sécurité (JWT).
- **Résultat attendu :** Session valide créée.

### UC-03 : Ouverture d’un compte bancaire
- **Objectif :** Permettre l’ouverture d’un compte chèque ou épargne.
- **Acteur principal :** Client
- **Flux principal :**
  1. Le Client (authentifié) choisit le type de compte.
  2. Le Système crée le compte avec solde initial à 0.
  3. Journalisation comptable de la création.
- **Résultat attendu :** Compte actif disponible.

### UC-04 : Consultation des soldes et historiques
- **Objectif :** Permettre la consultation des comptes et des transactions.
- **Acteur principal :** Client
- **Flux principal :**
  1. Le Client sélectionne un de ses comptes.
  2. Le Système récupère le solde calculé (ou mis en cache) et la liste des transactions récentes depuis le journal comptable.
- **Résultat attendu :** Informations affichées avec exactitude.

### UC-05 : Virement bancaire (interne)
- **Objectif :** Permettre un transfert de fonds d'un compte à un autre.
- **Acteur principal :** Client
- **Flux principal :**
  1. Le Client saisit le compte bénéficiaire et le montant.
  2. Le Système vérifie le solde du compte source, appliquant une "idempotency key" pour éviter les doublons.
  3. Création des écritures comptables sous garantie *exactly-once* (transaction DB).
- **Résultat attendu :** Soldes mis à jour, confirmation envoyée.

---

## Glossaire Métier
- **KYC (Know Your Customer) :** Exigences réglementaires d'identification des clients.
- **AML (Anti-Money Laundering) :** Contrôles contre le blanchiment d'argent.
- **MFA (Multi-Factor Authentication) :** Mécanisme d'authentification forte.
- **Idempotency Key :** Clé unique passée par le client pour s'assurer qu'une requête de paiement n'est exécutée qu'une seule fois.
- **Append-Only Ledger :** Journal comptable où les entrées ne peuvent être qu'ajoutées (aucune modification ou suppression de transaction passée).

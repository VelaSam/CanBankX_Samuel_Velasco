# Architecture Decision Record (ADR)

## ADR-0002: Stratégie de Persistance et Transactions (Exactly-Once)

### Status
Accepted

### Context
Le cahier des charges bancaires stipule que :
- Les écritures comptables doivent garantir un fonctionnement "exactly-once" ou "effectively-once".
- Le journal comptable doit être en mode "append-only".
- Les bases de données doivent garantir l'intégrité via un schéma strict, incluant des FK, index et transactions.

### Decision
**MySQL** a été choisi comme SGBD relationnel principal.
1. **Idempotency Key :** Toute requête modifiant l'état (Paiement, Transfert) nécessitera une `Idempotency-Key` dans les *headers* HTTP de la requête cliente. Cette clé sera stockée dans une table SQL dédiée aux requêtes idempotentes. Si une clé a déjà été traitée, le système retournera le résultat mis en cache de la requête précédente sans re-traiter la procédure.
2. **Transactions ACID :** Chaque modification (ex: réduire le solde A, augmenter le solde B, ajouter les logs comptables) se fera obligatoirement à l'intérieur d'une seule et même transaction de base de données (BEGIN ... COMMIT). 
3. **Journalisation Append-Only :** La table `ledger_entries` ne permettra aucune mise à jour (`UPDATE`) ou suppression (`DELETE`), uniquement des insertions (`INSERT`), garantissant de ce fait un historique financier parfaitement auditable.

### Consequences
**Positives :**
- Répond aux critères du cahier de charges (sécurité, transactions fiables).
- Protection contre les requêtes répétées accidentelles de l'interface client (grâce à l'A idempotency key).

**Négatives :**
- MySQL introduira une latence réseau et une contrainte de scalabilité pour les verrous (mutex/lock?) sur les écritures fréquentes. Il faut s'attendre à pouvoir le lire dans les métriques sous la forme d'un plafond dans le nombre de *Request Per Second* (RPS), justifiant plus tard l'implémentation de Caching (Redis).
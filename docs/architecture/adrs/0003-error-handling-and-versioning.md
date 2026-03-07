# Architecture Decision Record (ADR)

## ADR-0003: Stratégie d'Erreurs et de Versionnage d'API

### Status
Accepted

### Context
Pour garantir une intégration claire, sécurisée et évolutive, nous devons standardiser la manière dont notre API répond (succès et erreurs) et évolue dans le temps sans briser les contrats existants.

### Decision
1. **Versionnage par URI :** Toutes les routes de l'API REST incluront une version majeure dans le chemin de l'URL, ex: `/api/v1/accounts`.
2. **Codes HTTP Standards :** Le système utilisera scrupuleusement les codes de statut HTTP :
   - `200 OK`, `201 Created`, `204 No Content`
   - `400 Bad Request` (erreur de validation)
   - `401 Unauthorized` (token manquant/invalide)
   - `403 Forbidden` (accès refusé, ex: mauvais rôle)
   - `404 Not Found` (ressource absente)
   - `409 Conflict` (idempotency conflict, état incohérent)
   - `500 Internal Server Error` (erreur technique masquée pour l'utilisateur)
3. **Format Standardisé d'Erreur (RFC 7807) :** Toutes les erreurs retourneront un objet JSON structuré contenant un `code` interne, un `message` lisible pour le développeur, et éventuellement des `details` (ex: champs de formulaire invalides).

Exemple de réponse d'erreur 400:
```json
{
  "error_code": "INVALID_FUNDS",
  "message": "Solde insuffisant pour effectuer ce transfert.",
  "details": ["solde_actuel: 100", "montant_demande: 500"]
}
```

### Consequences
**Positives :**
- (Ceci ne va pas arriver pour ce projet mais dans un vrai projet ça serait le cas) Les développeurs frontend (mobile ou web) peuvent coder leur logique de *retry* et d'affichage d'erreurs en fonction de codes HTTP fiables.
- Le versionnage par URI permet l'introduction d'une v2 en parallèle pendant une période de transition sans impacter les anciens clients.

**Négatives :**
- Ajoute de la *boilerplate* dans les contrôleurs pour formater chaque erreur de la même manière. Nous devrons écrire un *middleware* ou une fonction utilitaire en Go pour forcer ce contrat.

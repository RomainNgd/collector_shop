# Exigences fonctionnelles et backlog

## 1. Exigences fonctionnelles de Collector.shop

| ID | Exigence reformulée | Priorité |
|---|---|---|
| RF-01 | Un visiteur peut consulter le catalogue et le détail des objets sans être authentifié. | V1 |
| RF-02 | Un particulier doit créer un compte pour acheter ou vendre. Un même compte peut être acheteur et vendeur. | V1 |
| RF-03 | L'espace personnel présente les achats, ventes, historiques, évaluations, centres d'intérêt et préférences de notification. | V1 |
| RF-04 | Un vendeur peut gérer plusieurs boutiques et publier des objets avec photos, description, prix et frais de port. Son statut de particulier reste visible. | V1 |
| RF-05 | Une annonce n'est publiée qu'après un contrôle automatique ou une validation par Collector. | V1 |
| RF-06 | L'acheteur et le vendeur peuvent échanger dans un chat modérable. Le système bloque l'échange d'email et de numéro de téléphone. | V1 |
| RF-07 | L'administrateur gère les catégories, annonces, comptes vendeurs et opérations de modération. Lui seul peut créer une catégorie. | V1 |
| RF-08 | Le système recommande des objets selon les centres d'intérêt renseignés par l'acheteur. Le parcours de navigation sera ajouté en V2. | V1/V2 |
| RF-09 | L'utilisateur choisit les notifications reçues dans l'application et par email : nouvelle annonce ciblée ou modification de prix. | V1 |
| RF-10 | Le paiement par carte est réalisé sur la plateforme ; le paiement direct est interdit et Collector prélève 5 % de commission. | V1 |
| RF-11 | Un acheteur et un vendeur peuvent s'évaluer après une transaction. | V1 |
| RF-12 | Chaque changement de prix est historisé et transmis au système de notification et au composant de détection de fraude. | V1 |
| RF-13 | L'application peut communiquer avec une solution externe de détection des prix anormaux et vendeurs suspects. | V1 |
| RF-14 | L'interface prend en charge l'internationalisation et l'accessibilité. | V1 |
| RF-15 | Le système peut transmettre des campagnes ciblées aux plateformes publicitaires partenaires. | V1 |

L'architecture doit permettre l'ajout ultérieur d'enchères, de ventes en direct, d'un assistant avant-vente et d'outils d'analyse des ventes.

## 2. Backlog du prototype implémenté

Le prototype couvre le parcours d'achat, le paiement et l'administration d'un catalogue. Les fonctions marketplace vendeur, boutiques, chat, évaluations, recommandations, notifications, commission et fraude restent hors de son périmètre actuel.

| ID | User story | Critères d'acceptation |
|---|---|---|
| US-01 | En tant que visiteur, je veux consulter le catalogue afin de découvrir les objets disponibles. | **Étant donné** un visiteur non connecté, **quand** il ouvre le catalogue ou une fiche produit, **alors** les produits, catégories, descriptions, images et prix applicables sont affichés sans demander d'authentification. |
| US-02 | En tant que visiteur, je veux créer un compte et me connecter afin de passer une commande. | Une adresse déjà utilisée est refusée ; le mot de passe est stocké avec bcrypt ; des identifiants valides créent une session JWT de 24 h dans un cookie `HttpOnly`, `SameSite=Lax` et `Secure` en HTTPS ; des identifiants invalides renvoient une erreur sans ouvrir de session. |
| US-03 | En tant qu'acheteur, je veux gérer mon panier afin de préparer ma commande. | L'utilisateur peut ajouter, retirer et modifier la quantité d'un produit ; une quantité doit être un entier positif ; le total tient compte des promotions actives ; un panier vide ne peut pas créer de commande. |
| US-04 | En tant qu'acheteur authentifié, je veux créer et consulter mes commandes afin de suivre mes achats. | La création exige une session valide ; l'API recalcule les produits, quantités et prix ; la commande apparaît dans l'historique et son détail est accessible uniquement à son propriétaire ou à l'administrateur. |
| US-05 | En tant qu'acheteur, je veux payer ma commande par carte afin de finaliser mon achat. | Une session Stripe Checkout est créée pour une commande autorisée ; seules les URLs de retour configurées sont acceptées ; aucune donnée bancaire n'est saisie dans Collector Shop ; le webhook signé met à jour l'état du paiement. |
| US-06 | En tant qu'administrateur, je veux gérer les catégories et les produits afin de maintenir le catalogue. | Un utilisateur non administrateur reçoit une réponse `403` ; l'administrateur peut créer, modifier et supprimer une catégorie ou un produit ; le prix doit être positif ; l'image envoyée doit être un fichier image conforme à la taille autorisée. |
| US-07 | En tant qu'administrateur, je veux gérer les promotions afin d'appliquer une réduction aux produits. | Une promotion peut être fixe ou en pourcentage ; sa valeur est positive et un pourcentage ne dépasse pas 100 ; elle peut cibler tous les produits ou une sélection ; seules les promotions actives modifient le prix affiché et commandé. |

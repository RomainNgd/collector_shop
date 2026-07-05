# Plan de remédiation sécurité

## 1. Objectif et périmètre

Ce plan traite les risques observés sur la V1 de Collector Shop.

Périmètre analysé :

- API Go et authentification JWT ;
- application SvelteKit ;
- pipeline GitHub Actions ;
- images Docker ;
- déploiement K3s et Argo CD ;
- PostgreSQL, Prometheus et Grafana.

Le plan distingue les protections déjà en place des risques résiduels. Une action est considérée comme terminée uniquement après un test de validation.

## 2. Méthode de priorisation

Chaque risque reçoit deux notes de 1 à 4 :

- **probabilité** : facilité et fréquence possible de l'incident ;
- **impact** : conséquence sur la confidentialité, l'intégrité ou la disponibilité.

Le score est calculé ainsi :

```text
Score = probabilité × impact
```

|   Score | Priorité      | Traitement attendu                              |
| ------: | ------------- | ----------------------------------------------- |
| 12 à 16 | P0 — critique | Bloquer une livraison et corriger immédiatement |
|  8 à 11 | P1 — haute    | Corriger avant la prochaine version             |
|   4 à 7 | P2 — moyenne  | Planifier dans une prochaine itération          |
|   1 à 3 | P3 — faible   | Accepter temporairement ou surveiller           |

## 3. Protections déjà en place

| Contrôle                         | État    | Preuve technique                                                                              |
| -------------------------------- | ------- | --------------------------------------------------------------------------------------------- |
| Chiffrement des flux publics     | Couvert | Ingress TLS avec cert-manager et Let's Encrypt                                                |
| Stockage des mots de passe       | Couvert | Hachage bcrypt dans `go-api/services/auth_service.go`                                         |
| Authentification et autorisation | Couvert | JWT, middleware serveur et contrôle du rôle administrateur                                    |
| Protection des commandes         | Couvert | Contrôle du propriétaire et recalcul des prix côté API                                        |
| Paiement                         | Couvert | Stripe Checkout et validation cryptographique du webhook                                      |
| Validation des images            | Couvert | Taille, extension, signature binaire et nom généré en UUID                                    |
| Détection automatisée            | Couvert | Trivy bloque les résultats `HIGH` et `CRITICAL` corrigibles ; scan local propre le 05/07/2026 |
| Secrets de production            | Couvert | Secrets créés dans K3s et absents des manifests versionnés                                    |

## 4. Synthèse des risques

| ID     | Risque                                               | Probabilité | Impact | Score | Priorité | État                                        |
| ------ | ---------------------------------------------------- | ----------: | -----: | ----: | -------- | ------------------------------------------- |
| SEC-01 | La CI accepte les vulnérabilités hautes et critiques |           3 |      4 |    12 | P0       | Corrigé localement, CI à confirmer          |
| SEC-02 | SonarCloud peut être ignoré sans faire échouer la CI |           2 |      4 |     8 | P1       | Corrigé localement, SonarCloud à confirmer  |
| SEC-03 | Aucune limitation des tentatives de connexion        |           3 |      3 |     9 | P1       | À traiter                                   |
| SEC-04 | En-têtes HTTP de sécurité incomplets                 |           3 |      3 |     9 | P1       | À traiter                                   |
| SEC-05 | Déploiement d'images utilisant le tag `latest`       |           2 |      4 |     8 | P1       | À traiter                                   |
| SEC-06 | Sauvegarde et restauration PostgreSQL non définies   |           2 |      4 |     8 | P1       | À traiter                                   |
| SEC-07 | Durcissement et isolation Kubernetes incomplets      |           2 |      3 |     6 | P2       | Corrigé localement, déploiement à confirmer |
| SEC-08 | Cycle de vie des JWT limité                          |           2 |      3 |     6 | P2       | Planifié                                    |
| SEC-09 | Détection et alertes de sécurité insuffisantes       |           2 |      3 |     6 | P2       | Planifié                                    |

## 5. Actions détaillées

### SEC-01 — Rendre les scans Trivy bloquants

**Constat initial :** les trois scans utilisaient `exit-code: "0"` dans `.github/workflows/reusable-build-and-scan.yml`. Une vulnérabilité haute ou critique produisait un rapport, mais la pipeline restait verte.

**Risque :** une dépendance ou une image vulnérable peut être publiée sur Docker Hub.

**Correction technique :**

1. Remplacer `exit-code: "0"` par `exit-code: "1"` pour le dépôt et les deux images.
2. Conserver l'envoi du rapport SARIF avec une condition `always()`.
3. Autoriser une exception uniquement avec un identifiant CVE, une justification, un responsable et une date d'expiration.
4. Interdire les exceptions globales ou sans échéance.

**Validation :**

```sh
trivy fs --severity HIGH,CRITICAL --exit-code 1 .
```

La CI doit échouer si Trivy détecte une vulnérabilité interdite. Après correction ou exception documentée, elle doit redevenir verte.

**Responsable :** DevOps — **Échéance :** immédiate.

**Avancement :**

- les scans sont bloquants uniquement pour les résultats `HIGH` et `CRITICAL` disposant d'un correctif ;
- `pgx/v5`, `quic-go`, `golang.org/x/crypto` et `golang.org/x/net` ont été mis à jour vers leurs versions corrigées ;
- l'image API est compilée avec Go 1.26.4 afin de corriger les vulnérabilités de la bibliothèque standard présentes dans Go 1.26.3 ;
- `npm`, `npx`, Corepack et Yarn ont été retirés de l'image front d'exécution : ils ne sont pas utilisés par `node build` et contenaient les dépendances globales vulnérables `picomatch` et `sigstore` ;
- le scan Trivy local du 05/07/2026 retourne 0 vulnérabilité, 0 secret et 0 mauvaise configuration sur le dépôt ;
- les images Docker sont construites avec succès ; leur scan final reste à confirmer dans la prochaine CI.

### SEC-02 — Activer réellement SonarCloud

**Constat initial :** le workflow terminait avec succès lorsque `SONAR_TOKEN` ou `SONAR_ORGANIZATION` manquait. L'analyse était alors marquée comme ignorée.

**Risque :** les vulnérabilités de code, la duplication et la complexité ne sont pas contrôlées.

**Correction technique :**

1. Ajouter `SONAR_TOKEN` dans les secrets GitHub du dépôt.
2. Définir l'organisation `romainngd` dans `sonar-project.properties`.
3. Rendre le token obligatoire dans `reusable-sonarcloud.yml`.
4. Utiliser `SonarSource/sonarqube-scan-action@v6`.
5. Supprimer le comportement silencieux `Skip SonarCloud when not configured` sur `main`.
6. Ajouter `-Dsonar.qualitygate.wait=true` au scan.
7. Configurer une Quality Gate : aucune nouvelle vulnérabilité, couverture du nouveau code au moins égale à 80 % et duplication inférieure à 3 %.

**Validation :** la dernière pipeline doit afficher l'étape `SonarCloud scan` comme exécutée. Une Quality Gate rouge doit faire échouer le job.

**Responsable :** Lead developer — **Échéance :** avant la prochaine livraison.

**Avancement :**

- workflow rendu obligatoire et Quality Gate bloquante ;
- la publication Docker attend maintenant la réussite de SonarCloud ;
- l'organisation `romainngd` est versionnée dans la configuration du projet et l'action SonarCloud utilise la version sécurisée `v6` ;
- les secrets JWT statiques utilisés par les tests ont été remplacés par des valeurs aléatoires générées à l'exécution ;
- le prochain scan SonarCloud doit confirmer la disparition de l'alerte `go:S6437` et le bon fonctionnement du jeton `SONAR_TOKEN`.

### SEC-03 — Limiter les tentatives de connexion

**Constat :** `POST /auth/login` est public et ne possède ni temporisation ni limitation de débit.

**Risque :** attaque par force brute, credential stuffing et consommation excessive des ressources bcrypt.

**Correction technique :**

1. Ajouter un middleware Gin spécifique à `/auth/login`.
2. Utiliser comme clé l'adresse IP source et l'email normalisé.
3. Autoriser au maximum 5 échecs sur une période de 15 minutes.
4. Retourner `429 Too Many Requests` avec un en-tête `Retry-After`.
5. Réinitialiser le compteur après une authentification réussie.
6. Conserver le même message d'erreur pour un email absent ou un mot de passe incorrect.
7. Configurer explicitement les proxys de confiance Gin afin de ne pas accepter un faux `X-Forwarded-For`.

Un stockage mémoire suffit tant que l'API reste mono-réplique. Redis sera nécessaire si plusieurs instances de l'API partagent la charge.

**Validation :** un test d'intégration envoie six connexions invalides. Les cinq premières reçoivent `401`, la sixième reçoit `429`. Une connexion valide reste possible après expiration du délai.

**Responsable :** développeur back-end — **Échéance :** prochaine version.

### SEC-04 — Ajouter les en-têtes HTTP de sécurité

**Constat :** les réponses publiques du front et de l'API ne présentent pas systématiquement HSTS, CSP, protection anti-framing et politique de référent.

**Risque :** clickjacking, chargement de ressources non autorisées et réduction de la protection du navigateur contre les injections.

**Correction technique :** créer un `Middleware` Traefik et l'associer à l'Ingress :

```yaml
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: security-headers
  namespace: collector-shop-prod
spec:
  headers:
    frameDeny: true
    contentTypeNosniff: true
    referrerPolicy: strict-origin-when-cross-origin
    stsSeconds: 31536000
    stsIncludeSubdomains: true
    stsPreload: true
    permissionsPolicy: "camera=(), microphone=(), geolocation=()"
```

La CSP doit d'abord être déployée en mode `Content-Security-Policy-Report-Only`, puis rendue bloquante après vérification des ressources SvelteKit et Stripe.

**Validation :**

```sh
curl -I https://collector-app.romainnigond.fr
curl -I https://collector-api.romainnigond.fr/products
```

Les réponses doivent contenir les en-têtes attendus sans casser l'affichage, l'authentification ni le paiement.

**Responsable :** DevOps et développeur front-end — **Échéance :** prochaine version.

### SEC-05 — Déployer des images immuables

**Constat :** la CI publie un tag lié au SHA Git, mais les manifests K3s utilisent `romain2311/go-api:latest` et `romain2311/collector-spa:latest`.

**Risque :** le contenu réellement déployé n'est pas identifiable avec certitude. Le rollback est difficile à reproduire.

**Correction technique :**

1. Remplacer `latest` par le SHA complet du commit validé dans les manifests.
2. Faire modifier les manifests par une étape GitOps contrôlée ou par Argo CD Image Updater.
3. Conserver l'ancienne valeur du tag pour permettre le rollback.
4. À terme, utiliser le digest `sha256` de l'image plutôt qu'un tag mutable.

**Validation :**

```sh
kubectl -n collector-shop-prod get deployment go-api \
  -o jsonpath='{.spec.template.spec.containers[0].image}'
kubectl -n collector-shop-prod rollout history deployment/go-api
```

L'image affichée doit correspondre au commit Git présenté. Un retour à la révision précédente doit être testé.

**Responsable :** DevOps — **Échéance :** prochaine livraison.

### SEC-06 — Sauvegarder PostgreSQL et tester la restauration

**Constat :** PostgreSQL utilise un volume persistant, mais aucun processus de sauvegarde n'est défini. Un volume persistant ne constitue pas une sauvegarde.

**Risque :** perte des comptes, commandes et états de paiement après corruption ou perte du serveur.

**Correction technique :**

1. Exécuter quotidiennement `pg_dump -Fc` depuis un `CronJob` Kubernetes.
2. Chiffrer et copier l'archive vers un stockage externe au VPS.
3. Conserver 7 sauvegardes quotidiennes et 4 sauvegardes hebdomadaires.
4. Contrôler le code retour, la taille du fichier et la date de la dernière sauvegarde.
5. Effectuer un test mensuel de restauration dans une base isolée.
6. Ne jamais écrire le mot de passe PostgreSQL dans le manifeste ou les logs.

Objectifs proposés : **RPO de 24 heures** et **RTO de 2 heures**.

**Validation :** restaurer une sauvegarde dans une base temporaire, puis vérifier le nombre d'utilisateurs, de produits, de commandes et d'éléments de commande.

**Responsable :** DevOps — **Échéance :** avant utilisation de données réelles.

### SEC-07 — Durcir les workloads Kubernetes

**Correction technique :** ajouter à chaque workload compatible :

```yaml
securityContext:
  runAsNonRoot: true
  seccompProfile:
    type: RuntimeDefault
containers:
  - securityContext:
      allowPrivilegeEscalation: false
      readOnlyRootFilesystem: true
      capabilities:
        drop: ["ALL"]
```

Les répertoires devant rester modifiables, comme `/app/upload` et éventuellement `/tmp`, doivent utiliser un volume dédié.

Ajouter ensuite des `NetworkPolicy` : refus par défaut, accès Ingress vers le front et l'API, accès API vers PostgreSQL, accès Prometheus vers le port `9090` et accès DNS nécessaire.

**Validation :** les pods démarrent avec les nouveaux paramètres, Trivy ne signale plus ces mauvaises configurations et un pod non autorisé ne peut pas joindre PostgreSQL.

**Responsable :** DevOps — **Échéance :** itération suivante.

**Avancement :** les workloads du front, de l'API et de PostgreSQL utilisent maintenant un utilisateur numérique non-root, `RuntimeDefault`, une racine en lecture seule, aucune élévation de privilèges et aucune capability Linux. Les chemins modifiables utilisent des volumes dédiés. Le scan Trivy local ne signale plus les règles `KSV-0014` et `KSV-0118`. Le démarrage dans K3s reste à valider après publication des nouvelles images.

### SEC-08 — Renforcer le cycle de vie des JWT

**Correction technique :**

- n'accepter explicitement que l'algorithme `HS256` ;
- ajouter et vérifier les claims `iss`, `aud` et `iat` ;
- réduire la durée du jeton d'accès ;
- prévoir la rotation de `JWT_SECRET` ;
- utiliser ensuite un jeton de rafraîchissement révocable si la durée de session doit rester longue.

**Validation :** les tests doivent refuser un jeton expiré, signé avec un autre algorithme, destiné à une autre audience ou signé avec l'ancien secret après la période de rotation.

**Responsable :** développeur back-end — **Échéance :** itération suivante.

### SEC-09 — Ajouter des alertes de sécurité

**Constat :** les métriques HTTP sont collectées, mais Alertmanager est désactivé et aucune alerte de sécurité n'est définie.

**Correction technique :** créer des alertes sur :

- hausse des réponses `401` sur `/auth/login` ;
- taux de réponses `5xx` supérieur à 5 % pendant 5 minutes ;
- indisponibilité de l'API ;
- absence de sauvegarde récente ;
- certificat TLS proche de l'expiration.

Exemple PromQL pour les échecs de connexion :

```promql
sum(rate(collector_http_requests_total{route="/auth/login",status="401"}[5m])) > 1
```

**Validation :** provoquer l'alerte dans un environnement de test et vérifier sa réception, son acquittement et son retour à l'état normal.

**Responsable :** DevOps — **Échéance :** itération suivante.

## 6. Ordre de mise en œuvre

### Phase 1 — Sécuriser la livraison

1. Rendre Trivy bloquant.
2. Activer SonarCloud et sa Quality Gate.
3. Conserver les rapports de scan comme preuves.

### Phase 2 — Protéger l'application et les données

1. Ajouter le rate limiting de la connexion.
2. Déployer les en-têtes HTTP.
3. Utiliser des images immuables.
4. Automatiser et tester la sauvegarde PostgreSQL.

### Phase 3 — Réduire les risques résiduels

1. Durcir les pods et ajouter les règles réseau.
2. Renforcer le cycle de vie des JWT.
3. Ajouter les alertes de sécurité.

## 7. Suivi du plan

### Journal de traitement

| Date       | Risque | Action réalisée                                                                                | Preuve                                                     | Statut                         |
| ---------- | ------ | ---------------------------------------------------------------------------------------------- | ---------------------------------------------------------- | ------------------------------ |
| 05/07/2026 | SEC-01 | Scans bloquants, dépendances mises à jour, Go 1.26.4 et outils npm inutiles retirés du runtime | Tests réussis et scan Trivy du dépôt à 0 résultat bloquant | Validation CI attendue         |
| 05/07/2026 | SEC-02 | Quality Gate obligatoire, action v6, organisation corrigée et secrets de test aléatoires       | Tests Go réussis ; nouveau scan SonarCloud attendu         | Validation SonarCloud attendue |
| 05/07/2026 | SEC-07 | `securityContext`, utilisateur non-root et volumes inscriptibles ajoutés                       | Trivy : 0 mauvaise configuration haute ou critique         | Validation K3s attendue        |

Pour passer une action à l'état **Corrigé**, il faut conserver :

- le lien vers la pull request ou le commit ;
- le résultat de la CI ;
- le rapport du scan concerné ;
- le test fonctionnel ou de sécurité réalisé ;
- la date et le nom de la personne ayant validé.

À la soutenance, les actions terminées seront présentées comme des risques traités. Les autres seront présentées comme des risques résiduels connus, priorisés et associés à une méthode de validation.

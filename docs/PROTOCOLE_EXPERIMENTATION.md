# Protocole d'expérimentation technique

## 1. Objectif

Avant le développement du POC, les plateformes support des composants à développer ont été testées en bac à sable afin de valider ou rejeter leur adoption. Quatre expérimentations ont été menées :

1. l'orchestrateur de conteneurs : Minikube et K3s ;
2. la plateforme CI/CD : GitLab CI et GitHub Actions ;
3. le système d'observabilité : Prometheus et Grafana déployés par Argo CD ;
4. le provisionnement automatique de certificats TLS : cert-manager et Let's Encrypt sur l'Ingress Traefik.

Chaque expérimentation précise l'environnement utilisé, les étapes permettant de la reproduire, les résultats observés, puis les difficultés et limites identifiées.

Le choix entre une architecture microservices et une API modulaire est une décision de conception et non une expérimentation de plateforme : il est traité séparément en annexe.

## 2. Environnement d'essai commun

- poste de développement Windows 11 avec Git, Docker et `kubectl` ;
- VPS Linux mono-nœud de taille limitée (2 vCPU, 4 Go de RAM) ;
- dépôt source hébergé sur GitHub ;
- API Go, application SvelteKit et base PostgreSQL comme charges de test ;
- manifests Kubernetes du dossier `prod/k3s` et Applications Argo CD du dossier `prod/argocd`.

## 3. Expérimentation 1 — Orchestrateur : Minikube et K3s

### Solutions comparées

- Minikube pour exécuter Kubernetes sur le poste de développement ;
- K3s pour déployer Kubernetes directement sur le VPS cible.

### Protocole de reproduction

1. Sur le poste local, démarrer Minikube et déployer la base applicative :

   ```sh
   minikube start --memory 4096 --cpus 2
   kubectl apply -k prod/k3s
   kubectl -n collector-shop-prod get pods -w
   ```

2. Sur le VPS, installer K3s puis appliquer les mêmes manifests :

   ```sh
   curl -sfL https://get.k3s.io | sh -
   sudo kubectl get nodes
   kubectl apply -k prod/k3s
   ```

3. Dans les deux environnements, vérifier les objets utilisés par le POC : `Deployment`, `Service`, `Ingress`, `Secret`, volumes persistants, probes `/healthz` et `/readyz`, et `HorizontalPodAutoscaler` :

   ```sh
   kubectl top pods -n collector-shop-prod        # vérifie metrics-server, requis par le HPA
   kubectl -n collector-shop-prod get hpa collector-spa -w
   ```

4. Comparer la consommation mémoire/CPU au repos et la facilité d'installation et d'exposition publique.

### Résultats

Les deux distributions exécutent les mêmes manifests sans modification. Minikube n'a pas été retenu pour le serveur cible :

1. Minikube ajoute une machine virtuelle ou un conteneur intermédiaire et consomme davantage de mémoire et de CPU ; il est conçu pour un poste de développement, pas pour une exposition publique.
2. K3s fournit les API Kubernetes nécessaires dans une distribution allégée adaptée à un VPS, avec Traefik intégré comme contrôleur Ingress et `metrics-server` fourni par défaut, ce dont dépend le HPA.

K3s a donc été retenu pour l'hébergement ; Minikube reste utilisable localement pour tester les manifests avant un déploiement.

### Difficultés et limites

- le HPA reste `<unknown>` tant que `metrics-server` ne répond pas : la commande `kubectl top pods` doit être vérifiée avant tout test de charge ;
- la configuration DNS et l'ouverture des ports 80/443 sont nécessaires avant l'exposition publique ;
- cluster mono-nœud : la haute disponibilité du serveur lui-même n'est pas couverte.

## 4. Expérimentation 2 — Plateforme CI/CD : GitLab CI et GitHub Actions

### Solutions comparées

- GitLab CI avec un fichier `.gitlab-ci.yml` ;
- GitHub Actions avec des workflows YAML réutilisables.

### Protocole de reproduction

Une pipeline minimale identique a été montée sur les deux plateformes :

1. tests Go avec une base PostgreSQL de service (`services:` dans les deux syntaxes) ;
2. tests et vérifications SvelteKit (`npm test`, `npm run check`, `npm run lint`) ;
3. construction des deux images Docker ;
4. scan de sécurité du dépôt et des images ;
5. publication des images vers Docker Hub, conditionnée à la réussite des étapes précédentes.

Pour reproduire la version retenue : pousser une branche sur GitHub et observer l'exécution de `.github/workflows/ci.yml`, qui appelle les workflows réutilisables `reusable-*.yml` du même dossier.

### Résultats

Les deux plateformes couvrent les besoins. GitHub Actions a été retenu :

1. le code est déjà hébergé sur GitHub : aucun dépôt miroir ni synchronisation de secrets vers GitLab n'est nécessaire ;
2. le catalogue d'actions maintenues couvre Go, Node.js, Docker Buildx, SonarCloud, Trivy et l'envoi de rapports SARIF vers l'onglet Sécurité du dépôt ;
3. le découpage en workflows réutilisables permet d'imposer l'ordre tests → build → scan → publication, la publication n'étant exécutée que sur `main`.

### Difficultés et limites

- la configuration des secrets (`DOCKERHUB_TOKEN`, `SONAR_TOKEN`) et des permissions (`contents`, `security-events`) doit être explicite pour chaque workflow appelé ;
- un service PostgreSQL de job nécessite des healthchecks pour éviter des tests qui démarrent avant la base ;
- dépendance aux quotas et à la disponibilité des runners hébergés par GitHub.

## 5. Expérimentation 3 — Observabilité : Prometheus et Grafana via Argo CD

### Objectif

Valider qu'une pile d'observabilité complète (collecte de métriques + tableaux de bord) peut être installée et maintenue par GitOps sur le VPS, sans dépasser ses ressources, et qu'elle collecte réellement les métriques applicatives exposées par l'API (`collector_http_requests_total`, `collector_http_request_duration_seconds`).

### Protocole de reproduction

1. Installer Argo CD sur le cluster :

   ```sh
   kubectl create namespace argocd
   kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
   kubectl -n argocd rollout status deployment/argocd-server
   ```

2. Déclarer le DNS `grafana.romainnigond.fr` vers l'IP du VPS, puis appliquer les deux Applications Argo CD qui installent les charts Helm officiels `prometheus` (29.14.0) et `grafana` (10.5.14) dans le namespace `monitoring` :

   ```sh
   kubectl apply -f prod/argocd/monitoring-application.yaml
   kubectl -n argocd get applications prometheus grafana
   ```

3. Vérifier la collecte et l'accès :

   ```sh
   kubectl -n monitoring get pods
   kubectl -n monitoring get certificate grafana-tls -w
   kubectl -n monitoring get secret grafana -o jsonpath='{.data.admin-password}' | base64 -d
   ```

4. Générer du trafic sur l'API (parcours du catalogue ou tir k6 court) et vérifier dans Grafana que le tableau de bord « Collector Shop - Vue rapide » affiche le débit, le taux d'erreurs et la latence p95.

### Résultats

- l'ensemble de la pile est décrit dans un seul fichier versionné : un redéploiement de l'application ne supprime pas la supervision, et une dérive manuelle est corrigée automatiquement (`selfHeal: true`) ;
- Prometheus scrape l'endpoint `/metrics` de l'API (serveur dédié, port 9090) et `node-exporter` pour le serveur ;
- Grafana provisionne automatiquement la datasource Prometheus et trois tableaux de bord : vue applicative Collector Shop, Node Exporter Full et Go Runtime ;
- l'accès est limité : pas d'inscription, pas d'accès anonyme, cookies `Secure`, HSTS, et Prometheus reste interne au cluster (seul Grafana a un Ingress).

L'expérimentation valide l'adoption de Prometheus + Grafana gérés par Argo CD comme composante d'observabilité du POC (collecte de métriques).

### Difficultés et limites

- les valeurs par défaut des charts dépassent les capacités du VPS : il a fallu désactiver Alertmanager, kube-state-metrics, la pushgateway et les scrapes Kubernetes non utilisés, fixer des `resources.requests/limits` et limiter la rétention à 7 jours / 4 Go ;
- `initChownData` du chart Grafana a dû être désactivé pour respecter le `securityContext` non-root ;
- sans `enforce_domain` et `root_url`, les redirections de connexion Grafana étaient incorrectes derrière Traefik ;
- limite assumée : Alertmanager est désactivé, donc aucune alerte n'est émise — ce point est repris dans le plan de remédiation (SEC-09).

## 6. Expérimentation 4 — TLS automatique : cert-manager et Let's Encrypt

### Objectif

Valider le provisionnement et le renouvellement automatiques de certificats TLS pour les trois sous-domaines publics (`collector-app`, `collector-api`, `grafana`), sans manipulation manuelle de certificats sur le serveur.

### Protocole de reproduction

1. Ouvrir les ports 80 (challenge HTTP-01) et 443, et faire pointer les DNS vers le VPS.
2. Installer cert-manager et vérifier ses trois déploiements :

   ```sh
   kubectl -n cert-manager rollout status deployment/cert-manager
   kubectl -n cert-manager rollout status deployment/cert-manager-webhook
   kubectl -n cert-manager rollout status deployment/cert-manager-cainjector
   ```

3. Créer un `ClusterIssuer` ACME `letsencrypt-prod` (résolveur HTTP-01 via la classe Ingress `traefik`), puis annoter l'Ingress applicatif avec `cert-manager.io/cluster-issuer: letsencrypt-prod` et déclarer la section `tls` (voir `prod/k3s/ingress.yaml`).
4. Observer l'émission puis tester la chaîne :

   ```sh
   kubectl get clusterissuer letsencrypt-prod
   kubectl -n collector-shop-prod get certificate -w
   curl -vI https://collector-api.romainnigond.fr/products 2>&1 | grep -E "subject|issuer|expire"
   ```

### Résultats

- le certificat `collector-shop-tls` passe à l'état `Ready` et couvre les deux domaines applicatifs ; le certificat `grafana-tls` est émis automatiquement à partir de l'Ingress du chart Grafana, sans configuration supplémentaire : le mécanisme est bien réutilisable pour tout nouveau sous-domaine ;
- le renouvellement est automatique avant expiration (durée de vie Let's Encrypt de 90 jours) ;
- l'ensemble des flux publics du POC est servi en HTTPS, ce qui couvre l'exigence de chiffrement en transit.

L'expérimentation valide l'adoption de cert-manager comme composant de sécurité du cluster, géré globalement et indépendamment des manifests applicatifs.

### Difficultés et limites

- le challenge HTTP-01 échoue tant que le DNS ne pointe pas vers le VPS ou que le port 80 est fermé : l'ordre DNS → pare-feu → issuer doit être respecté ;
- Let's Encrypt applique des limites de débit : les essais ont d'abord été réalisés avec l'issuer de staging avant de passer à `letsencrypt-prod` ;
- limite identifiée : aucune alerte n'existe si un renouvellement échoue ; une alerte d'expiration est prévue dans le plan de remédiation (SEC-09).

## 7. Synthèse des décisions

| Sujet | Solution retenue | Justification principale |
|---|---|---|
| Orchestration | K3s | Kubernetes léger adapté au VPS, Traefik et metrics-server intégrés |
| CI/CD | GitHub Actions | Intégration directe au dépôt et catalogue d'actions pour Go, Docker, SonarCloud et Trivy |
| Observabilité | Prometheus + Grafana via Argo CD | Collecte des métriques applicatives et serveur, gérée par GitOps, dimensionnée pour le VPS |
| TLS | cert-manager + Let's Encrypt | Provisionnement et renouvellement automatiques pour tous les sous-domaines |

Ces choix sont adaptés au périmètre actuel du POC. Ils pourront être réévalués si le trafic, la taille de l'équipe ou les besoins de disponibilité augmentent.

## Annexe — Décision d'architecture : API modulaire plutôt que microservices

Cette comparaison relève de la conception applicative et non d'une expérimentation de plateforme ; elle est conservée ici comme trace de la décision.

### Options étudiées

- microservices séparés pour l'authentification, le catalogue, les commandes et le paiement ;
- API unique organisée en modules métier (contrôleurs, services, modèles).

### Analyse

1. Chaque microservice ajoute une image, une configuration, une pipeline, une supervision et une gestion de versions. Cette charge opérationnelle est disproportionnée pour une petite équipe et un POC.
2. Le catalogue, les commandes et le paiement manipulent des données fortement liées : leur séparation imposerait des échanges réseau supplémentaires, de la duplication de données et des mécanismes de cohérence distribuée.

### Décision

Le POC utilise une API Go unique découpée en contrôleurs, services et modèles. Ce découpage interne préserve la possibilité d'extraire un service autonome plus tard si la charge ou l'organisation de l'équipe le justifie, conformément à l'exigence d'évolutivité du contexte (enchères, ventes en direct, analyse des ventes).

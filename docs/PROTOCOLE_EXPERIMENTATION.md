# Protocole d'expérimentation technique

## 1. Objectif

Cette expérimentation valide trois choix structurants pour Collector Shop :

- l'architecture applicative ;
- l'orchestrateur de conteneurs ;
- la plateforme CI/CD.

Les solutions ont été comparées dans un environnement isolé avant de retenir celles utilisées par le POC.

## 2. Environnement d'essai

- poste de développement avec Git et Docker ;
- API Go, application SvelteKit et base PostgreSQL ;
- dépôt source hébergé sur GitHub ;
- serveur Linux cible de taille limitée ;
- manifests Kubernetes du dossier `prod/k3s`.

## 3. Architecture microservices

### Solutions comparées

- microservices séparés pour l'authentification, le catalogue, les commandes et le paiement ;
- API unique organisée en modules métier.

### Protocole

1. Découper le parcours d'achat en quatre domaines métier.
2. Définir les API nécessaires entre ces domaines.
3. Étudier leur déploiement dans des conteneurs indépendants.
4. Comparer la gestion des données, des erreurs et du débogage avec une API modulaire unique.

### Résultat

L'approche microservices est techniquement possible, mais elle n'a pas été retenue pour la V1.

Deux contraintes principales ont été identifiées :

1. Chaque service demande une image, une configuration, une pipeline, une supervision et une gestion de versions supplémentaires. Cette charge est trop importante pour une petite équipe et un POC.
2. Le catalogue, les commandes et le paiement utilisent des données fortement liées. Leur séparation imposerait des échanges réseau, de la duplication de données et des mécanismes de cohérence distribuée.

Le POC utilise donc une API Go unique, découpée en contrôleurs, services et modèles. Ce découpage reste suffisamment modulaire pour extraire un service plus tard si la charge ou l'organisation le justifie.

### Difficultés et limites

- définition des frontières entre les services ;
- gestion des erreurs entre plusieurs composants ;
- absence de test à grande échelle avec plusieurs équipes autonomes.

## 4. Minikube et K3s

### Solutions comparées

- Minikube pour exécuter Kubernetes localement ;
- K3s pour déployer Kubernetes sur le serveur cible.

### Protocole

1. Préparer le déploiement de PostgreSQL, de l'API et du front.
2. Vérifier les `Deployment`, `Service`, `Ingress`, `Secret` et volumes persistants.
3. Tester les probes de santé et le `HorizontalPodAutoscaler`.
4. Comparer les ressources nécessaires et la facilité d'installation.

### Résultat

Minikube permet de tester les objets Kubernetes, mais il n'a pas été retenu pour le serveur cible.

Deux raisons principales :

1. Minikube ajoute une couche locale et consomme davantage de mémoire et de CPU. Il est surtout adapté à un poste de développement.
2. K3s fournit les API Kubernetes nécessaires dans une distribution plus légère. Il est adapté à un VPS et intègre notamment Traefik pour l'Ingress.

K3s a donc été retenu. Les mêmes manifests permettent de gérer les services, les secrets, TLS, les probes et la montée en charge du front.

### Difficultés et limites

- configuration du DNS et des certificats TLS ;
- dépendance à `metrics-server` pour le HPA ;
- cluster composé d'un seul nœud, sans haute disponibilité du serveur lui-même.

## 5. GitLab CI et GitHub Actions

### Solutions comparées

- GitLab CI avec un fichier `.gitlab-ci.yml` ;
- GitHub Actions avec des workflows YAML réutilisables.

### Protocole

La comparaison a porté sur une pipeline minimale contenant :

1. les tests Go et SvelteKit ;
2. le lint et l'analyse statique ;
3. la construction des images Docker ;
4. le scan de sécurité ;
5. la publication des images.

### Résultat

Les deux solutions couvrent les besoins du projet. GitHub Actions a été retenu pour deux raisons :

1. Le code est déjà hébergé sur GitHub. Il n'est donc pas nécessaire de migrer ou de synchroniser le dépôt et les secrets avec GitLab.
2. GitHub Actions fournit un catalogue important d'actions prêtes à l'emploi pour Go, Node.js, Docker, SonarCloud, Trivy et l'envoi des rapports de sécurité.

La pipeline finale est découpée en workflows réutilisables dans `.github/workflows`. Elle exécute les tests, le lint, le build Docker, les scans et la publication des images.

### Difficultés et limites

- configuration des secrets Docker Hub et SonarCloud ;
- gestion précise des permissions GitHub ;
- dépendance à la disponibilité et aux quotas de la plateforme GitHub.

## 6. Synthèse des décisions

| Sujet | Solution retenue | Justification principale |
|---|---|---|
| Architecture | API modulaire | Moins de complexité opérationnelle et données plus simples à maintenir |
| Orchestration | K3s | Kubernetes léger et adapté au VPS cible |
| CI/CD | GitHub Actions | Intégration directe au dépôt et catalogue d'actions disponible |

Ces choix sont adaptés au périmètre actuel du POC. Ils pourront être réévalués si le trafic, la taille de l'équipe ou les besoins de disponibilité augmentent.

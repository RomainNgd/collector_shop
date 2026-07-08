# Qualité du code

## Objectif

Ces métriques permettent d'évaluer la qualité du code de Collector Shop en s'appuyant sur plusieurs caractéristiques d'ISO/IEC 25010 : la maintenabilité, la fiabilité, la sécurité et l'adéquation fonctionnelle.

Leur suivi régulier permet de détecter rapidement une dégradation du code et d'éviter l'accumulation de dette technique.

## 1. Couverture des tests automatisés

### Pourquoi l'utiliser ?

La couverture indique quelle proportion du code est exécutée par les tests automatisés. Elle aide à identifier les parties de l'application qui risquent de contenir des erreurs non détectées.

Cette métrique contribue principalement à la **fiabilité**, à l'**adéquation fonctionnelle** et à la **maintenabilité**.

### Comment l'évaluer ?

La couverture est calculée avec la formule suivante :

```text
Couverture = lignes de code exécutées par les tests / lignes de code exécutables × 100
```

Outils utilisés :

- `go test -cover` pour l'API Go ;
- Vitest avec l'option `--coverage` pour le front SvelteKit ;
- SonarCloud pour centraliser et suivre l'évolution de la couverture.

### Critères d'acceptation

- couverture du nouveau code supérieure ou égale à **80 %** ;
- couverture supérieure ou égale à **90 %** pour les fonctions critiques : authentification, commandes, promotions et paiements ;
- tous les tests doivent réussir.

### Prévention de la dette technique

Contrôler la couverture du nouveau code empêche l'ajout progressif de fonctionnalités non testées. Les futures modifications sont ainsi plus sûres et les régressions plus faciles à détecter.

## 2. Complexité du code

### Pourquoi l'utiliser ?

La complexité mesure la difficulté à comprendre et à tester une fonction. Une fonction contenant trop de conditions, de boucles ou de branches devient plus coûteuse à modifier et augmente le risque d'erreur.

Cette métrique est principalement liée à la **maintenabilité** et à la **testabilité**.

### Comment l'évaluer ?

La complexité cyclomatique correspond au nombre de chemins d'exécution possibles dans une fonction. Elle augmente notamment avec les instructions `if`, `switch`, `for` et les conditions imbriquées.

Outils utilisés :

- SonarCloud pour mesurer la complexité cognitive et cyclomatique ;
- `go vet`, ESLint et les revues de code pour compléter l'analyse.

### Critères d'acceptation

- complexité cyclomatique inférieure ou égale à **10 par fonction** ;
- aucune nouvelle fonction signalée comme excessivement complexe par SonarCloud ;
- les fonctions dépassant le seuil doivent être découpées ou faire l'objet d'une justification.

### Prévention de la dette technique

Le suivi de la complexité évite l'apparition de fonctions toujours plus longues et difficiles à maintenir. Le découpage régulier du code réduit le coût des évolutions et facilite l'ajout de tests.

## 3. Duplication du code

### Pourquoi l'utiliser ?

La duplication apparaît lorsqu'une même logique est copiée à plusieurs endroits. Une correction doit alors être répétée dans chaque copie, avec le risque d'oublier une occurrence ou de créer des comportements différents.

Cette métrique contribue à la **maintenabilité**, à la **réutilisabilité** et à la **fiabilité**.

### Comment l'évaluer ?

SonarCloud compare les blocs de code et calcule la densité de duplication :

```text
Taux de duplication = lignes dupliquées / nombre total de lignes × 100
```

### Critères d'acceptation

- taux de duplication inférieur ou égal à **3 % sur le nouveau code** ;
- aucune duplication de règles métier critiques, par exemple le calcul des prix ou des promotions ;
- toute duplication importante doit être factorisée dans une fonction, un service ou un composant partagé.

### Prévention de la dette technique

Cette métrique empêche la multiplication de copies difficiles à synchroniser. La factorisation rend les corrections plus rapides et garantit qu'une règle métier reste identique dans toute l'application.

## 4. Vulnérabilités de sécurité

### Pourquoi l'utiliser ?

Cette métrique mesure le nombre de vulnérabilités détectées dans le code, les dépendances et les images Docker. Une vulnérabilité non corrigée représente un risque pour les comptes utilisateurs, les commandes et les paiements.

Elle répond principalement aux exigences de **sécurité** et de **fiabilité**.

### Comment l'évaluer ?

Les vulnérabilités sont classées par niveau de gravité : faible, moyenne, haute ou critique.

Outils utilisés :

- SonarCloud pour les problèmes de sécurité présents dans le code ;
- Trivy pour les dépendances, les secrets, la configuration et les images Docker.

### Critères d'acceptation

- **aucune vulnérabilité critique ou haute** non corrigée ;
- aucun secret ou identifiant sensible présent dans le dépôt ;
- les vulnérabilités moyennes doivent être corrigées ou accompagnées d'une justification et d'une échéance.

### Prévention de la dette technique

Le suivi du nombre et de l'ancienneté des vulnérabilités empêche leur accumulation au fil des mises à jour. Les problèmes sont corrigés pendant qu'ils restent limités, avant de devenir plus difficiles ou plus coûteux à traiter.

## Synthèse

| Métrique | Outil principal | Critère d'acceptation |
|---|---|---|
| Couverture des tests | Go test, Vitest, SonarCloud | Au moins 80 % sur le nouveau code et 90 % sur le code critique |
| Complexité du code | SonarCloud | Complexité maximale de 10 par fonction |
| Duplication du code | SonarCloud | Maximum 3 % sur le nouveau code |
| Vulnérabilités | SonarCloud, Trivy | Aucune vulnérabilité critique ou haute |

Ces quatre métriques doivent être observées à chaque évolution du projet. Leur tendance est aussi importante que leur valeur : une dégradation régulière doit entraîner une correction avant l'ajout de nouvelles fonctionnalités.

## Mesures relevées le 08/07/2026

| Métrique | Valeur mesurée | Méthode de mesure | Conforme ? |
|---|---|---|---|
| Couverture API Go | **81,3 %** au global (153 fonctions de test, toutes vertes) | `go test ./... -coverprofile` puis `go tool cover -func` | Oui pour le seuil global de 80 % |
| Couverture front | **94,7 %** des instructions (97 tests dans 19 fichiers, tous verts) | `npm run test:coverage` (Vitest, rapport v8) | Oui |
| Complexité et duplication | Suivies par la Quality Gate SonarCloud, bloquante en CI ; relevé à capturer sur le tableau de bord à chaque exécution | SonarCloud (`sonar.qualitygate.wait=true`) | Oui tant que la Quality Gate est verte |
| Vulnérabilités | **0** vulnérabilité haute ou critique, 0 secret, 0 mauvaise configuration (scan du dépôt du 05/07/2026) | Trivy `vuln,secret,misconfig`, bloquant en CI | Oui |

Détail de la couverture Go par paquet : `models` 100 %, `routes` 98,2 %, `config` 94,9 %, `pkg/metrics` 92,3 %, `middlewares` 90,3 %, `controllers` 85,7 %, `services` 85,1 %, `database` 78,3 %.

### Axes d'amélioration mis en évidence

1. **Code critique sous le seuil renforcé** : `controllers` (85,7 %) et `services` (85,1 %) portent l'authentification, les commandes et les paiements, dont le seuil cible est 90 %. Les prochains tests doivent viser en priorité les branches non couvertes de ces deux paquets.
2. **Paquet `database` à 78,3 %**, sous le seuil global de 80 % : les chemins d'erreur des migrations et du seed sont les moins testés ; à compléter avant d'enrichir le schéma.
3. Ces deux écarts sont traités pendant qu'ils restent petits : c'est exactement le mécanisme anti-dette technique décrit plus haut — le seuil sur le *nouveau* code empêche l'écart de grandir, et la tendance par paquet indique où investir l'effort de test.

## Référence

- [ISO/IEC 25010:2023 — Modèle de qualité du produit](https://www.iso.org/fr/standard/78176.html)

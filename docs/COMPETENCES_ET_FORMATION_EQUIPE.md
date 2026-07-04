# Compétences et formation de l'équipe

## Cartographie des compétences

Le projet peut être réalisé par une petite équipe composée de profils complémentaires. Chaque membre possède une spécialité principale sans devoir maîtriser tous les domaines.

| Profil | Nombre | Compétences principales |
|---|---:|---|
| Product Owner | 1 | Besoins métier, backlog, critères d'acceptation et priorisation |
| Développeur back-end Go | 1 | Go, Gin, API REST, PostgreSQL, GORM, JWT et tests automatisés |
| Développeur front-end | 1 | SvelteKit, TypeScript, HTML/CSS, accessibilité et Vitest |
| Ingénieur DevOps/DevSecOps | 1 | GitHub Actions, Docker, Kubernetes/K3s, Argo CD, SonarCloud et Trivy |
| Testeur QA | 1 | Plans de test, tests d'intégration, tests E2E, sécurité et performance avec k6 |

Les développeurs peuvent se relire mutuellement, mais les sujets spécialisés restent attribués au profil compétent. Cette répartition évite de rechercher un développeur unique expert en développement, sécurité, infrastructure, tests et gestion de produit.

## Besoin de renforcement

La compétence à renforcer en priorité est l'intégration de la sécurité dans le développement. Les développeurs savent produire et tester du code, mais doivent mieux comprendre les contrôles SonarCloud, Trivy et les portes de sécurité de la pipeline CI/CD.

## Action de formation proposée

### Atelier DevSecOps appliqué à Collector Shop

- **Participants :** développeurs, DevOps et testeur QA ;
- **Durée :** deux jours ;
- **Format :** rappels courts puis exercices directement sur le dépôt ;
- **Contenu :** analyse SonarCloud, scan Trivy, gestion des secrets, correction d'une vulnérabilité et configuration d'une porte de qualité ;
- **Résultat attendu :** chaque participant sait identifier un résultat de sécurité, évaluer sa gravité et proposer une correction adaptée.

La formation est validée par un exercice pratique : corriger une vulnérabilité volontairement introduite et obtenir une pipeline entièrement conforme. Une courte revue trois mois plus tard permet de vérifier que ces pratiques sont toujours appliquées.

# Load Tests

Ce dossier regroupe les tests de charge transverses du monorepo.

Convention retenue:

- un dossier par application cible
- un sous-dossier par outil de test
- la documentation et les scripts d'observation au meme endroit que le scenario

Structure actuelle:

- `tests/load/collector-spa/k6`: scenario `k6` pour provoquer un scale-out du front
- `tests/load/collector-spa`: scripts d'observation du scale-out Kubernetes

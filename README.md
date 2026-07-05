# Collector Shop

Collector Shop est une application e-commerce de demonstration pour des produits de collection.

Le projet est organise en monorepo:

- `go-api`: API Go/Gin, base PostgreSQL, authentification JWT, commandes, promotions et paiement Stripe optionnel.
- `collector-spa`: front SvelteKit avec rendu serveur, panier, espace client et administration.
- `dev`: scripts pour lancer l'environnement local.
- `build`: scripts et Dockerfiles pour construire les images.
- `prod/k3s`: base Kubernetes k3s pour un deploiement de production.
- `tests/load`: scenarios de charge k6.

## Prerequis

- Docker
- Go 1.26.4
- Node.js 22
- npm
- PowerShell pour les scripts Windows fournis

## Lancer en local

Depuis la racine:

```powershell
powershell -ExecutionPolicy Bypass -File .\dev\start-dev.ps1
```

Le script relance PostgreSQL, charge les fixtures, demarre l'API et demarre le front.

URLs locales:

- API: `http://localhost:8080`
- Front: `http://localhost:5173`

Voir [dev/README.md](dev/README.md) pour le detail, notamment Stripe en local.

## Tests et qualite

API Go:

```sh
cd go-api
go test ./...
```

Front SvelteKit:

```sh
cd collector-spa
npm ci
npm run check
npm test
npm run lint
```

La CI GitHub Actions est decoupee en workflows reutilisables:

- tests Go
- checks SvelteKit
- lint Go/front
- SonarCloud
- build Docker
- scan Trivy
- push Docker Hub

Le workflow principal est dans [.github/workflows/ci.yml](.github/workflows/ci.yml).

## Build Docker

```powershell
.\build\scripts\build-images.ps1
```

Voir [build/README.md](build/README.md) pour les tags et registres.

## Production

Une base k3s est disponible dans `prod/k3s`.

```sh
kubectl apply -k prod/k3s
```

Voir [prod/k3s/README.md](prod/k3s/README.md) avant de deployer: les secrets, le DNS, le registre Docker et les points de readiness doivent etre adaptes a l'environnement cible.

## Notes

- `JWT_SECRET` doit etre identique entre `go-api` et `collector-spa`.
- `API_BASE_URL` sert aux appels serveur internes de SvelteKit.
- `API_PUBLIC_BASE_URL` sert aux URLs visibles depuis le navigateur.
- Les vrais secrets ne doivent pas etre commits.

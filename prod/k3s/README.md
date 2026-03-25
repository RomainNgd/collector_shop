# k3s Production Baseline

Ce dossier contient une base Kubernetes pour lancer `postgres`, `go-api` et `collector-spa` sur k3s.

## 1) Prerequis

- Les images de prod sont tirees depuis Docker Hub:
  - `romain2311/go-api:latest`
  - `romain2311/collector-spa:latest`
- Comme les images sont privees, cree le secret Docker Hub directement dans le cluster.
- Avec Argo CD, ce secret ne doit pas etre commit dans le repo.
- Nom attendu par les deployments: `dockerhub-pull-secret` dans le namespace `collector-shop-prod`.
- Exemple de creation:

```powershell
kubectl create secret docker-registry dockerhub-pull-secret `
  --namespace collector-shop-prod `
  --docker-server=https://index.docker.io/v1/ `
  --docker-username=romain2311 `
  --docker-password='<dockerhub-token>' `
  --dry-run=client -o yaml | kubectl apply -f -
```
- Modifier les secrets dans `secret.yaml`.
- `metrics-server` doit etre installe si tu veux que le `HorizontalPodAutoscaler` fonctionne.
- Verifie que `kubectl top pods -n collector-shop-prod` repond avant de lancer un test de charge.

## 2) Configurer le DNS

L'Ingress expose maintenant deux sous-domaines:

- `collector-app.romainnigond.fr` pour le front `collector-spa`
- `collector-api.romainnigond.fr` pour l'API `go-api`

Ils doivent pointer vers l'IP publique de ton noeud k3s ou de ton load balancer.

Exemple hosts local:

```text
<NODE_IP> collector-app.romainnigond.fr
<NODE_IP> collector-api.romainnigond.fr
```

## 3) Deployer

```sh
kubectl apply -k prod/k3s
kubectl -n collector-shop-prod get pods
kubectl -n collector-shop-prod get ingress
```

Si tu deployes avec Argo CD, cree d'abord le secret Docker Hub dans le cluster, puis laisse Argo CD synchroniser le reste du dossier `prod/k3s`.

Si tu republies une nouvelle image avec le tag `latest`, les pods doivent etre recrees pour forcer un nouveau pull:

```sh
kubectl -n collector-shop-prod rollout restart deployment/go-api deployment/collector-spa
```

## 4) Acces

- Front: `http://collector-app.romainnigond.fr`
- API: `http://collector-api.romainnigond.fr/products`, `http://collector-api.romainnigond.fr/auth/login`, etc.

## 5) Test de charge et scale-out

Le front `collector-spa` dispose d'un `HorizontalPodAutoscaler` dans `collector-spa-hpa.yaml`.

Pour observer un scale-out:

```sh
kubectl -n collector-shop-prod get hpa collector-spa -w
kubectl -n collector-shop-prod get pods -l app=collector-spa -w
```

Puis lance le scenario `k6` range dans `tests/load/collector-spa`.

Exemple:

```sh
k6 run tests/load/collector-spa/k6/scale-up.js -e BASE_URL=http://collector-app.romainnigond.fr
```

Si tu attaques l'Ingress via l'IP du noeud plutot que via `collector-app.romainnigond.fr`, ajoute `-e HOST_HEADER=collector-app.romainnigond.fr`.

## Notes

- `API_BASE_URL` est interne au cluster (`http://go-api:8080`) pour le SSR.
- `API_PUBLIC_BASE_URL` est publique (`http://collector-api.romainnigond.fr`) pour les URLs d'images dans le navigateur.
- Le scenario de charge est prevu pour montrer le scale-out du front. L'API `go-api` reste mono-replica dans cette base k3s.

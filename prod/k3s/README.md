# k3s Production Baseline

Ce dossier contient une base Kubernetes pour lancer `postgres`, `go-api` et `collector-spa` sur k3s.

## 1) Prerequis

- Les images doivent exister dans un registre accessible par k3s, ou etre importees dans les noeuds.
- Adapter les tags d'images dans `go-api.yaml` et `collector-spa.yaml` si besoin.
- Modifier les secrets dans `secret.yaml`.
- `metrics-server` doit etre installe si tu veux que le `HorizontalPodAutoscaler` fonctionne.
- Verifie que `kubectl top pods -n collector-shop-prod` repond avant de lancer un test de charge.

## 2) Adapter le domaine local de test

Le front est expose sur `collector.local` via Ingress.
Ajoute une entree DNS ou hosts vers l'IP de ton noeud k3s.

Exemple hosts:

```text
<NODE_IP> collector.local
```

## 3) Deployer

```sh
kubectl apply -k prod/k3s
kubectl -n collector-shop-prod get pods
kubectl -n collector-shop-prod get ingress
```

## 4) Acces

- Front: `http://collector.local`
- API: `http://collector.local/products`, `http://collector.local/auth/login`, etc.

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
k6 run tests/load/collector-spa/k6/scale-up.js -e BASE_URL=http://collector.local
```

Si tu attaques l'Ingress via l'IP du noeud plutot que via `collector.local`, ajoute `-e HOST_HEADER=collector.local`.

## Notes

- `API_BASE_URL` est interne au cluster (`http://go-api:8080`) pour le SSR.
- `API_PUBLIC_BASE_URL` est publique (`http://collector.local`) pour les URLs d'images dans le navigateur.
- Le scenario de charge est prevu pour montrer le scale-out du front. L'API `go-api` reste mono-replica dans cette base k3s.

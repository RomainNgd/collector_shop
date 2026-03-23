# k3s Production Baseline

Ce dossier contient une base Kubernetes pour lancer `postgres`, `go-api` et `collector-spa` sur k3s.

## 1) Prerequis

- Les images doivent exister dans un registre accessible par k3s, ou etre importees dans les noeuds.
- Adapter les tags d'images dans `go-api.yaml` et `collector-spa.yaml` si besoin.
- Modifier les secrets dans `secret.yaml`.

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

## Notes

- `API_BASE_URL` est interne au cluster (`http://go-api:8080`) pour le SSR.
- `API_PUBLIC_BASE_URL` est publique (`http://collector.local`) pour les URLs d'images dans le navigateur.

# Production VPS

Cette documentation decrit le deploiement cible pour l'exercice: un VPS, k3s, Argo CD, Traefik et TLS Let's Encrypt.

Domaines utilises:

- Front: `https://collector-app.romainnigond.fr`
- API: `https://collector-api.romainnigond.fr`

## 1. Preparer le VPS

Sur le VPS, prevoir une distribution Linux simple, par exemple Ubuntu ou Debian.

Ouvrir au minimum:

- `22/tcp` pour SSH
- `80/tcp` pour le challenge HTTP Let's Encrypt
- `443/tcp` pour HTTPS

Exemple avec `ufw`:

```sh
sudo ufw allow OpenSSH
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

Configurer ensuite les DNS pour pointer vers l'IP publique du VPS:

```text
collector-app.romainnigond.fr -> <IP_DU_VPS>
collector-api.romainnigond.fr -> <IP_DU_VPS>
```

## 2. Installer k3s

Installation simple:

```sh
curl -sfL https://get.k3s.io | sh -
```

Verifier le cluster:

```sh
sudo kubectl get nodes
sudo kubectl get pods -A
```

Pour utiliser `kubectl` sans `sudo`:

```sh
mkdir -p ~/.kube
sudo cp /etc/rancher/k3s/k3s.yaml ~/.kube/config
sudo chown "$USER:$USER" ~/.kube/config
chmod 600 ~/.kube/config
```

## 3. Installer cert-manager

cert-manager fournit les certificats Let's Encrypt utilises par l'Ingress.

```sh
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/latest/download/cert-manager.yaml
kubectl -n cert-manager rollout status deployment/cert-manager
kubectl -n cert-manager rollout status deployment/cert-manager-webhook
kubectl -n cert-manager rollout status deployment/cert-manager-cainjector
```

Le `ClusterIssuer` est versionne dans `prod/k3s/clusterissuer.yaml`.

## 4. Installer Argo CD

```sh
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
kubectl -n argocd rollout status deployment/argocd-server
```

Pour acceder a l'interface Argo CD depuis ta machine:

```sh
kubectl -n argocd port-forward svc/argocd-server 8080:443
```

Mot de passe initial:

```sh
kubectl -n argocd get secret argocd-initial-admin-secret \
  -o jsonpath="{.data.password}" | base64 -d
```

Si le repo GitHub est prive, ajoute une cle SSH deploy key dans GitHub et configure le repo dans Argo CD avant de creer l'application.

## 5. Creer les secrets

Les manifests attendent un namespace `collector-shop-prod`.

```sh
kubectl create namespace collector-shop-prod
```

Secret applicatif:

```sh
kubectl -n collector-shop-prod create secret generic collector-shop-secrets \
  --from-literal=DB_NAME=ecommerce \
  --from-literal=DB_USER=golang \
  --from-literal=DB_PASSWORD='<db-password>' \
  --from-literal=JWT_SECRET='<jwt-secret>' \
  --from-literal=STRIPE_SECRET_KEY='' \
  --from-literal=STRIPE_WEBHOOK_SECRET='' \
  --dry-run=client -o yaml | kubectl apply -f -
```

Si les images Docker Hub sont privees:

```sh
kubectl -n collector-shop-prod create secret docker-registry dockerhub-pull-secret \
  --docker-server=https://index.docker.io/v1/ \
  --docker-username=romain2311 \
  --docker-password='<dockerhub-token>' \
  --dry-run=client -o yaml | kubectl apply -f -
```

## 6. Deployer avec Argo CD

Verifier la branche dans `prod/argocd/application.yaml`.
Par defaut elle pointe sur `main`.

```sh
kubectl apply -f prod/argocd/application.yaml
kubectl -n argocd get application collector-shop
```

Argo CD synchronise ensuite le dossier `prod/k3s`.

## 7. Verification

```sh
kubectl -n collector-shop-prod get pods
kubectl -n collector-shop-prod get ingress
kubectl -n collector-shop-prod get certificate
```

Les URLs attendues sont:

- `https://collector-app.romainnigond.fr`
- `https://collector-api.romainnigond.fr/products`

## Notes

- `API_BASE_URL` reste interne au cluster: `http://go-api:8080`.
- `API_PUBLIC_BASE_URL` est en HTTPS pour le navigateur.
- `GIN_MODE=release` est active en production.
- Les vrais secrets ne sont pas necessaires dans Git pour cet exercice; cree-les dans le cluster.

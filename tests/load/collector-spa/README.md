# Collector SPA Load Test

Ce dossier contient un scenario de charge `k6` pour montrer qu'un second pod `collector-spa`
peut etre cree quand la charge HTTP monte suffisamment.

## Prerequis

- le cluster k3s est deploie avec `kubectl apply -k prod/k3s`
- `metrics-server` fonctionne
- `kubectl top pods -n collector-shop-prod` repond
- l'Ingress `collector-app.romainnigond.fr` pointe bien vers ton noeud k3s

## Lancer l'observation du scale-out

Dans un terminal PowerShell:

```powershell
powershell -ExecutionPolicy Bypass -File .\tests\load\collector-spa\assert-scale-up.ps1
```

Ou depuis Linux/macOS:

```sh
./tests/load/collector-spa/assert-scale-up.sh
```

Ce script attend qu'un pod supplementaire `collector-spa` devienne `Ready`.
La version bash accepte les memes reglages via variables d'environnement:
`NAMESPACE`, `DEPLOYMENT`, `HPA_NAME`, `TIMEOUT_SECONDS`, `POLL_SECONDS`.

## Lancer le test de charge

Avec `k6` installe localement:

```powershell
k6 run .\tests\load\collector-spa\k6\scale-up.js -e BASE_URL=https://collector-app.romainnigond.fr
```

Si tu vises l'IP du noeud k3s au lieu du host `collector-app.romainnigond.fr`, surcharge le header Host:

```powershell
k6 run .\tests\load\collector-spa\k6\scale-up.js -e BASE_URL=http://192.168.1.50 -e HOST_HEADER=collector-app.romainnigond.fr
```

## Variables utiles

- `BASE_URL`: URL d'entree du front. Defaut `https://collector-app.romainnigond.fr`
- `PATH_TO_HIT`: chemin HTTP cible. Defaut `/`
- `HOST_HEADER`: header `Host` optionnel si tu passes par l'IP du noeud
- `PEAK_VUS`: charge max. Defaut `180`
- `BATCH_SIZE`: nombre de requetes paralleles par iteration. Defaut `4`

## Conserver les preuves d'une campagne

Pour archiver le resultat d'un tir (preuve pour le plan de remediation et la soutenance):

```sh
k6 run tests/load/collector-spa/k6/scale-up.js \
  -e BASE_URL=https://collector-app.romainnigond.fr \
  --summary-export "k6-summary-$(date +%Y%m%d-%H%M).json"
```

Completer avec des captures du dashboard Grafana `Collector Shop - Vue rapide`
(debit, erreurs 5xx, latence p95) prises pendant le tir.

## Notes

- Le test vise volontairement le front `collector-spa`, car il peut scaler sans la contrainte
  du volume persistant actuellement monte par `go-api`.
- La page `/` declenche aussi des appels SSR vers `go-api`, donc tu exerces une charge
  realiste sur la pile sans avoir a ecrire un scenario plus artificiel.

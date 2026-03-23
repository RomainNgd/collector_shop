# Image Builds

Ce dossier sert uniquement a fabriquer des images Docker propres pour `collector-spa` et `go-api`.

## Images produites

- `collector-shop/go-api:<tag>`
- `collector-shop/collector-spa:<tag>`

## Prerequis

- Docker installe
- Les variables d'environnement runtime seront fournies plus tard par k3s ou par votre manifeste Kubernetes

## Build sous PowerShell

```powershell
.\build\scripts\build-images.ps1
.\build\scripts\build-images.ps1 -Tag v0.1.0
.\build\scripts\build-images.ps1 -Tag v0.1.0 -Registry registry.example.com/collector-shop
```

## Build sous shell POSIX

```sh
sh ./build/scripts/build-images.sh
TAG=v0.1.0 sh ./build/scripts/build-images.sh
TAG=v0.1.0 REGISTRY=registry.example.com/collector-shop sh ./build/scripts/build-images.sh
```

## Notes

- Le front SvelteKit est maintenant construit avec `adapter-node` pour produire un runtime Node explicite.
- L'API Go est construite en multi-stage et embarque uniquement le binaire final.
- Le dossier d'upload du backend n'est pas integre dans l'image finale comme donnee persistante. Il devra etre monte en volume dans Kubernetes si necessaire.
- Ces scripts ne poussent pas les images. Ils se contentent de les construire localement avec un tag propre.

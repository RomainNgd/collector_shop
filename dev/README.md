# Dev

Depuis la racine du projet, lance:

```powershell
powershell -ExecutionPolicy Bypass -File .\dev\start-dev.ps1
```

Le script:

- vide le volume Postgres local dans `go-api`
- relance la base via `docker compose`
- attend que PostgreSQL reponde
- charge les fixtures avec `go run . seed`
- ouvre l'API Go dans une fenetre PowerShell
- ouvre le front Vite dans une autre fenetre PowerShell
- transmet au front le meme `JWT_SECRET` que l'API pour que SvelteKit verifie les cookies JWT cote serveur

Mode simulation:

```powershell
powershell -ExecutionPolicy Bypass -File .\dev\start-dev.ps1 -DryRun
```

## Stripe en local

Pour activer le paiement Stripe sur la demo:

1. copie `go-api/.env.example` vers `go-api/.env` si ce n'est pas deja fait
2. renseigne `STRIPE_ENABLED=true`
3. renseigne `STRIPE_SECRET_KEY` avec une cle sandbox/test Stripe
4. renseigne `STRIPE_CHECKOUT_ALLOWED_ORIGINS=http://localhost:5173` pour autoriser les retours Checkout vers le front local
5. lance l'API
6. dans un autre terminal, utilise la Stripe CLI officielle:

```powershell
stripe listen --events checkout.session.completed,checkout.session.expired,checkout.session.async_payment_failed,checkout.session.async_payment_succeeded --forward-to localhost:8080/payments/stripe/webhook
```

La CLI affiche une valeur `whsec_...`: reporte-la dans `STRIPE_WEBHOOK_SECRET`, puis redemarre l'API.

Cartes de test utiles dans Stripe Checkout:

- `4242 4242 4242 4242` pour un paiement carte reussi
- une date future valide
- n'importe quel CVC a 3 chiffres

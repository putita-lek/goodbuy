# Deploying GoodBuy

Live at **https://goodbuy.sohamjadhav.com**. Everything runs on free tiers.

| Part | Where | What it runs |
|---|---|---|
| Website + `/api` proxy | Cloudflare Worker `goodbuy` (`front-end/wrangler.jsonc`) | The built React app; forwards `/api/*` to the API; pings `/health` every 10 min so the API never sleeps |
| API | Render web service `goodbuy-api` (https://goodbuy-api.onrender.com), Frankfurt, Free | Docker image `ghcr.io/s0hambot/goodbuy-api:latest` |
| Database | TiDB Cloud Starter cluster `goodbuy` (AWS Frankfurt), database `goodbuy` | MySQL-compatible; needs TLS (`DB_TLS=true`) |

The API image is deployed from GitHub Packages rather than from this repo, so Render
never downloads the ~800 MB of Kaggle CSVs stored in Git LFS.

## Update the website

From `front-end/` (needs `npx wrangler login` once):

```bash
npm run deploy
```

## Update the API

From `back-end/` (needs `docker login ghcr.io`, e.g. with a `gh` token that has `write:packages`):

```bash
docker build -t ghcr.io/s0hambot/goodbuy-api:latest .
docker push ghcr.io/s0hambot/goodbuy-api:latest
```

Then in Render: **goodbuy-api → Manual Deploy → Deploy latest reference**.

## Render environment variables

`DB_HOST`, `DB_PORT=4000`, `DB_USER`, `DB_PASSWORD`, `DB_NAME=goodbuy`, `DB_TLS=true`,
`SEED_DATA=false`, `CORS_ORIGINS=https://goodbuy.sohamjadhav.com`. Health check path: `/health`.

## Reload the database

Put the TiDB connection values in `back-end/tidb.local.env` (git-ignored; same keys as
above), then from `back-end/`:

```bash
docker run --rm --env-file tidb.local.env -v "$PWD:/src" -w /src golang:1.22-alpine go run ./cmd/seed
docker run --rm --env-file tidb.local.env -v "$PWD:/src" -v "$PWD/../kaggle:/kaggle:ro" -w /src golang:1.22-alpine go run ./cmd/import-kaggle -dir /kaggle
```

Re-running the import replaces the imported prices; it doesn't duplicate them.

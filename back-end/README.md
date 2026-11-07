Requirement:
- It is a website which user input their preference and their shopping list items. then, the back-end must suggest options for each item on where should the user buy based on their preference and data of the related items in the database.
- Data: product (product item, product price, and product store) [in the next phase, the database might also store user profile. but as of now (for simplicity) we will not store it)
- API Spec: https://docs.google.com/spreadsheets/d/1TMo_HwhqbyV9DZ2_VAMGc-wQZoiK4hf-CWIuBg7jkY0/edit?gid=0#gid=0
Tech Stack: Go, Go Fiber, GORM, MySQL
---

# GoodBuy backend

Go Fiber + GORM + MySQL API that suggests where to buy each item on a shopping list, based on the user's location, priorities, sustainability preferences and budget.

## Run

With Docker (starts MySQL and the API, and loads sample Edinburgh stores and products):

```bash
docker compose up --build
```

Without Docker for the API (needs Go 1.22+ and a MySQL 8 database):

```bash
cp .env.example .env   # then export the variables, or set them in your shell
go run ./cmd/server
```

Run unit tests: `go test ./...`. Example calls are in `requests.http`.

| Variable | Default | Meaning |
|---|---|---|
| `PORT` | `8080` | HTTP port |
| `DB_HOST` / `DB_PORT` | `127.0.0.1` / `3306` | MySQL address |
| `DB_USER` / `DB_PASSWORD` / `DB_NAME` | `goodbuy` | MySQL credentials and database |
| `SEED_DATA` | `true` | Load the sample catalogue when the `stores` table is empty |
| `CORS_ORIGINS` | `*` | Allowed browser origins for the website |

## API

| Method | Path | Purpose |
|---|---|---|
| POST | `/submit-preference` | Save preference and shopping list, compute suggestions, return `session_id` |
| GET | `/shopping-item/option-result?session_id=` | Up to 3 ranked options per shopping list item |
| POST | `/shopping-item/submit-option` | Save the option the user chose for each item |
| GET | `/shopping-item/summary?session_id=` | Chosen option per item, total price and whether it fits the budget |
| GET | `/health` | Health check |

Errors return `{"error": "..."}` with status 400 (bad input), 404 (unknown session) or 500. `submit-option` returns `{"is_success": false, "error": "..."}` on failure.

### Differences from the spec sheet

- Every option uses `price_gbp` (the sheet mixes `price_gbp` and `price_per_unit_gbp`).
- The summary always uses `selected_option` (the sheet uses `options` for the second item); it is `null` for items with no choice yet.
- `option_id` is unique within each item (1, 2, 3), ordered best first; the sheet repeats ids.
- `is_local_businesss` keeps the sheet's spelling so the frontend matches it.
- Extra fields: `item_quantity` and `distance_miles` on results; `total_price_gbp`, `budget_gbp` and `within_budget` on the summary.
- The sheet's example request has a trailing comma after `"organic": 3`, which is not valid JSON; clients must send valid JSON.

## How suggestions are chosen

1. **Match:** the item text is split into words (filler words like "a", "of", "in" are dropped and plurals made singular). A product matches if more than half of those words appear in its name or keywords.
2. **Filter:** drop products whose store is further than `maximum_distance_from_user_location_miles` (straight-line distance) or whose `price × quantity` is over `budget_gbp`.
3. **Score:** match quality + the priority ranks (`best_price`, `closest_location`: rank 1 counts more than rank 2) + a smaller bonus for sustainability ranks (`local_business`, `recycleable`, `organic`: rank 1 to 3). The bonus is capped below the top priority's weight, so it breaks ties but never overrides the user's first priority. A rank of 0 means "don't care".
4. **Return** the top 3, best first.

Suggestions are saved with the session as a snapshot, so later price changes don't alter a session's results or summary.

## Data model

- `stores`: name, latitude, longitude, is_local_business
- `products`: name, keywords, price_gbp, store_id, is_recyclable, is_organic
- `shopping_sessions`, `session_items`, `session_item_options`: one shopping flow and its suggestions. These are not user profiles and hold no personal details besides the submitted location.

## Layout

```
cmd/server         entry point
internal/config    environment config
internal/database  connection and migrations
internal/models    GORM models
internal/dto       request and response JSON
internal/repository database access
internal/service   matching, ranking and session logic
internal/handler   Fiber routes
internal/seed      sample data
```

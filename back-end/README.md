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

## Importing real supermarket prices (Kaggle)

The sample catalogue only has a few hand-written products. `cmd/import-kaggle` loads real prices from the Kaggle dataset [Time Series UK Supermarket Data](https://www.kaggle.com/datasets/declanmcalinden/time-series-uk-supermarket-data) (Aldi, ASDA, Morrisons, Sainsbury's and Tesco).

> **The prices are from 2024, not today.** The dataset has daily snapshots from 9 Jan to 13 Apr 2024 and is not updated. Each imported product stores its `price_date`, so the app can show how old a price is. Lidl, M&S and other chains are not in the dataset and keep their seed data.

### 1. Get the files

Download the dataset from Kaggle and unzip it into `goodbuy/kaggle/`, next to `back-end/`:

```
goodbuy/
  back-end/
  kaggle/
    All_Data_Aldi.csv  All_Data_ASDA.csv  All_Data_Morrisons.csv  All_Data_Sains.csv  All_Data_Tesco.csv
```

The files are about 800 MB in total, so keep `kaggle/` out of git.

With the Kaggle CLI (needs an API token in `~/.kaggle/kaggle.json`), from the `goodbuy` folder:

```bash
kaggle datasets download -d declanmcalinden/time-series-uk-supermarket-data -p kaggle --unzip
```

### 2. Run the import

Start the database and API first, so the tables exist and the sample stores are seeded (`docker compose up --build`). Then, from `back-end/`:

```bash
go run ./cmd/import-kaggle -dir ../kaggle -convenience-markup 8
```

The command reads the same `DB_*` variables as the server, and their defaults match `docker-compose.yml`, which publishes MySQL on `localhost:3306`. A full import takes 1–2 minutes and adds about 80,000 products.

| Flag | Default | Meaning |
|---|---|---|
| `-dir` | `../kaggle` | Folder with the `All_Data_*.csv` files |
| `-date` | latest day in each file | Snapshot day to import, as `YYYYMMDD` (e.g. `20240301`) |
| `-convenience-markup` | `0` | Percent added to prices for stores with "Local", "Express" or "Metro" in their name |
| `-batch` | `1000` | Rows per `INSERT` |
| `-dry-run` | `false` | Read the files and print how many products each has, without touching the database |

You can run it again at any time. Each run deletes the products it imported before (`source = 'kaggle'`) for the stores it updates, then inserts the new snapshot. Seed products and other stores are never changed, and saved shopping sessions keep their own snapshot of each option.

### How the data is mapped

The dataset has one price per **chain**, but GoodBuy has prices per **store**. So each chain's prices are copied to every store in the `stores` table whose name starts with that chain's name:

| CSV file | Applied to stores named… | Seed stores that match |
|---|---|---|
| `All_Data_Sains.csv` | `Sainsbury's…` | Sainsbury's Cameron Toll, Sainsbury's Local Princes Street |
| `All_Data_Tesco.csv` | `Tesco…` | Tesco Express Morningside |
| `All_Data_Aldi.csv` | `Aldi…` | none |
| `All_Data_ASDA.csv` | `ASDA…` | none |
| `All_Data_Morrisons.csv` | `Morrisons…` | none |

A chain with no matching store is skipped with a message. To use it, add a store first (for example `Aldi Leith Walk`, with its real latitude and longitude) and run the import again.

The prices are the chains' main online prices. Convenience stores (Sainsbury's Local, Tesco Express) often charge more, which `-convenience-markup` roughly allows for.

Each CSV row becomes a product like this:

| CSV column | `products` column | Notes |
|---|---|---|
| `names` | `name` | Extra spaces removed; cut to 255 characters |
| `prices_(£)` | `price_gbp` | Plus the markup for convenience stores |
| `prices_unit_(£)` | `unit_price_gbp` | `NULL` when the file has `0` (Tesco often does) |
| `unit` | `unit` | e.g. `kg`, `l`, `unit` |
| `category` | `keywords` | e.g. `fresh_food` becomes "fresh food", so the matcher can find it |
| `own_brand` | `is_own_brand` | |
| `date` | `price_date` | |
| none | `is_organic` | True when the name contains "organic" |
| none | `is_recyclable` | Always false, because the dataset doesn't record packaging |
| none | `source` | Always `kaggle`; seed products have an empty `source` |

Only one day is imported. Each file has a row for every product on every day, so the import keeps the rows for the chosen day (by default the latest) and ignores the rest. When a product is listed more than once that day under different categories, it keeps one row and adds every category to `keywords`.

### Known limitation: search speed

Product search (`ProductRepository.FindByAnyToken`) uses `LIKE '%word%'` across every product and loads all matches. With the full import, a four-item shopping list takes about 1 second to submit. A MySQL `FULLTEXT` index on `name` and `keywords`, or a limit on matches per item, would speed this up.

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
- `products`: name, keywords, price_gbp, store_id, is_recyclable, is_organic, plus unit_price_gbp, unit, is_own_brand, price_date and source (filled by the Kaggle import)
- `shopping_sessions`, `session_items`, `session_item_options`: one shopping flow and its suggestions. These are not user profiles and hold no personal details besides the submitted location.

## Layout

```
cmd/server         entry point
cmd/import-kaggle  one-off import of the Kaggle supermarket prices
internal/config    environment config
internal/database  connection and migrations
internal/models    GORM models
internal/dto       request and response JSON
internal/repository database access
internal/service   matching, ranking and session logic
internal/handler   Fiber routes
internal/seed      sample data
```

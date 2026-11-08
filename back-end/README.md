Requirement:
- It is a website which user input their preference and their shopping list items. then, the back-end must suggest options for each item on where should the user buy based on their preference and data of the related items in the database.
- Data: product (product item, product price, and product store) [in the next phase, the database might also store user profile. but as of now (for simplicity) we will not store it)
- API Spec: https://docs.google.com/spreadsheets/d/1TMo_HwhqbyV9DZ2_VAMGc-wQZoiK4hf-CWIuBg7jkY0/edit?gid=0#gid=0
Tech Stack: Go, Go Fiber, GORM, MySQL
---

# GoodBuy backend

Go Fiber + GORM + MySQL API that suggests where to buy each item on a shopping list, based on the user's location, priorities, sustainability preferences and budget.

## Run

With Docker (starts MySQL and the API, and seeds Edinburgh retailers and their stores; products come from the Kaggle import below):

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
| `SEED_DATA` | `true` | Seed retailers and Edinburgh stores when the `retailers` table is empty |
| `CORS_ORIGINS` | `*` | Allowed browser origins for the website |

## Importing real supermarket prices (Kaggle)

The seed creates retailers and stores but no products. `cmd/import-kaggle` loads real prices from the Kaggle dataset [Time Series UK Supermarket Data](https://www.kaggle.com/datasets/declanmcalinden/time-series-uk-supermarket-data) (Aldi, ASDA, Morrisons, Sainsbury's and Tesco).

> **The prices are from 2024, not today.** The dataset has daily snapshots from 9 Jan to 13 Apr 2024 and is not updated. Each imported price stores its `price_date`, so the app can show how old a price is. Lidl, Sainsbury's Local, Real Foods and the farmers' market are not imported, so they have no products unless you add some.

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

Start the database and API first, so the tables exist and the retailers and stores are seeded (`docker compose up --build`). Then, from `back-end/`:

```bash
go run ./cmd/import-kaggle -dir ../kaggle
```

The command reads the same `DB_*` variables as the server, and their defaults match `docker-compose.yml`, which publishes MySQL on `localhost:3306`. A full import takes under a minute and adds about 103,000 products, each with one row in `product_prices`.

| Flag | Default | Meaning |
|---|---|---|
| `-dir` | `../kaggle` | Folder with the `All_Data_*.csv` files |
| `-date` | latest day in each file | Snapshot day to import, as `YYYYMMDD` (e.g. `20240301`) |
| `-batch` | `1000` | Rows per `INSERT` |
| `-dry-run` | `false` | Read the files and print how many products each has, without touching the database |

You can run it again at any time. For each retailer it imports, it deletes the products it imported before (`source = 'kaggle'`) and their prices, then inserts the new snapshot, all in one transaction. Hand-entered products and other retailers are never changed, and saved shopping sessions keep their own snapshot of each option.

### How the data is mapped

Each CSV file belongs to one retailer, matched by its exact name in the `retailers` table:

| CSV file | Retailer | Seeded stores |
|---|---|---|
| `All_Data_Sains.csv` | `Sainsbury's` | 5 |
| `All_Data_Tesco.csv` | `Tesco` (incl. Express and Extra) | 31 |
| `All_Data_Aldi.csv` | `Aldi` | 8 |
| `All_Data_ASDA.csv` | `ASDA` | 3 |
| `All_Data_Morrisons.csv` | `Morrisons` (incl. Daily) | 13 |

`Sainsbury's Local` is a separate retailer and is deliberately not imported. A retailer with no stores is skipped with a message.

The dataset has one price per **chain**, and so does GoodBuy: each product is stored once in `products` for the retailer, with its price in `product_prices`, and that price applies at every one of the retailer's stores. The prices are the chains' main online prices; convenience formats (Express, Daily) may charge more in reality.

Each CSV row becomes a product and its price like this:

| CSV column | Column | Notes |
|---|---|---|
| `names` | `products.name` | Extra spaces removed; cut to 255 characters |
| `prices_(£)` | `product_prices.price_gbp` | |
| `prices_unit_(£)` | `product_prices.unit_price_gbp` | `NULL` when the file has `0` (Tesco often does) |
| `unit` | `product_prices.unit` | e.g. `kg`, `l`, `unit` |
| `category` | `products.keywords` | e.g. `fresh_food` becomes "fresh food", so the matcher can find it |
| `own_brand` | `products.is_own_brand` | |
| `date` | `product_prices.price_date` | |
| none | `products.is_organic` | True when the name contains "organic" |
| none | `products.is_recyclable` | Always false, because the dataset doesn't record packaging |
| none | `products.source` | Always `kaggle`; hand-entered products have an empty `source` |

A few names count as the same under MySQL's case- and accent-insensitive collation (e.g. "Café" and "Cafe"); only the first one is kept.

Only one day is imported. Each file has a row for every product on every day, so the import keeps the rows for the chosen day (by default the latest) and ignores the rest. When a product is listed more than once that day under different categories, it keeps one row and adds every category to `keywords`.

### Known limitation: search speed

Product search (`ProductRepository.FindProducts`) uses `LIKE '%word%'` on product names and keywords and loads every match. It only searches retailers with a store inside a latitude/longitude box around the user's maximum distance (`FindStores`), but a large chain nearby still means scanning all of its products. A MySQL `FULLTEXT` index on `name` and `keywords`, or a limit on matches per item, would speed this up.

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
2. **Filter:** drop products whose `price × quantity` is over `budget_gbp`. Pair each remaining product with up to 3 of its retailer's nearest stores within `maximum_distance_from_user_location_miles` (straight-line distance); each pair is one candidate, so the same product can be suggested at more than one branch.
3. **Score:** match quality + the priority ranks (`best_price`, `closest_location`: rank 1 counts more than rank 2) + a smaller bonus for sustainability ranks (`local_business`, `recycleable`, `organic`: rank 1 to 3). The bonus is capped below the top priority's weight, so it breaks ties but never overrides the user's first priority. A rank of 0 means "don't care".
4. **Return** the top 3, best first.

Suggestions are saved with the session as a snapshot, so later price changes don't alter a session's results or summary.

## Data model

- `retailers`: the brand or chain (e.g. Sainsbury's), or one independent business: name, is_local_business
- `stores`: one branch of a retailer: retailer_id, branch_name, format, address, postcode, latitude, longitude
- `products`: a catalogue item sold at all of a retailer's stores: retailer_id, name, keywords, is_recyclable, is_organic, plus is_own_brand and source (filled by the Kaggle import)
- `product_prices`: a product's current chain-wide price, one row per product: product_id, price_gbp, unit_price_gbp, unit, price_date

On startup, catalogue tables from earlier designs are dropped (old `stores`/`products` with a brand name or `store_id`, and `store_products`); products then need importing again.
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
internal/seed      retailers and Edinburgh stores
```

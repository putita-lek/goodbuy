# GoodBuy

_Say goodbye to bad buys._

Front end for the "Plan a shop" grocery planner (step 1: Your preferences).

**Stack:** React 18 + Vite + Tailwind CSS v4 + lucide-react icons.

## Run

```bash
npm install
npm run dev
```

Then open http://localhost:5173.

## Structure

```
src/
  App.jsx                    page layout, step state, API calls between steps
  index.css                  Tailwind + design tokens (colours, font)
  api/submitPreference.js    calls to the Go API (axios)
  geocode.js                 typed location -> lat/long via postcodes.io
  components/
    Header.jsx               logo, nav, avatar
    Stepper.jsx              3-step progress bar
    PreferencesCard.jsx      1. priority options + 2. location / distance / budget
    ShoppingList.jsx         3. shopping list (add / remove items)
    SidePanel.jsx            photo card + "Your shop, your way"
    ui.jsx                   shared Card, Field, OptionCard, RankBadge, ...
    ReviewPage.jsx           step 2: ranked options per item (Best match / Runner up)
    SummaryPage.jsx          step 3: totals, shops grouped by store, budget, route
    RouteMap.jsx             Leaflet map of the route
  data/shops.js              adapts API options to the UI (quantity totals, tags, labels)
public/basket.jpg            placeholder photo (replace with a full-res image)
```

Priority cards are ranked in the order you click them (1 = highest).

## Backend

The app talks to the Go API in `../back-end` (start it with `docker compose up --build`
there). Set `VITE_API_URL` when the API is not on `http://localhost:8080`.

1. **Review my list** sends `POST /submit-preference`, then loads
   `GET /shopping-item/option-result`. The API ranks the options; the page keeps its order.
2. **See summary** saves the chosen option per item with `POST /shopping-item/submit-option`.

Prices from the API are per item; the pages multiply them by the item quantity.

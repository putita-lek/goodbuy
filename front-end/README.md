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
  App.jsx                    page layout, step state, initial data
  index.css                  Tailwind + design tokens (colours, font)
  components/
    Header.jsx               logo, nav, avatar
    Stepper.jsx              3-step progress bar
    PreferencesCard.jsx      1. priority options + 2. location / distance / budget
    ShoppingList.jsx         3. shopping list (add / remove items)
    SidePanel.jsx            photo card + "Your shop, your way"
    ui.jsx                   shared Card, Field, OptionCard, RankBadge, ...
    ReviewPage.jsx           step 2: 3 shop options per item (Best match / Runner up)
    SummaryPage.jsx          step 3: totals, shops grouped by store, budget, route
  data/shops.js              mock shops + prices and the scoring/ranking logic
public/basket.jpg            placeholder photo (replace with a full-res image)
```

Priority cards are ranked in the order you click them (1 = highest).
Shop data and prices are mock data in `src/data/shops.js` — swap in a real API later.

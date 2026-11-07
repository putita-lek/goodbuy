// Mock shop + price data and the ranking logic used by pages 2 and 3.
// Replace SHOPS / BASE_PRICES with a real API later — the rest keeps working.

export const SHOPS = [
  { id: 'freshmart', name: 'FreshMart', type: 'Supermarket', distance: 0.7, priceFactor: 1.0, tags: ['Recycled bags'] },
  { id: 'greenway', name: 'Greenway Market', type: 'Local market', distance: 0.5, priceFactor: 1.15, tags: ['Local', 'Organic'] },
  { id: 'oak', name: 'Oak & Earth', type: 'Organic store', distance: 1.1, priceFactor: 1.4, tags: ['Organic', 'Local', 'Recycled bags'] },
  { id: 'farmers', name: 'Local Farmers Market', type: 'Local market', distance: 1.3, priceFactor: 1.2, tags: ['Local', 'Organic'] },
]

// Which shop tag satisfies which "What matters to you?" value.
export const VALUE_TAG = { local: 'Local', organic: 'Organic', recycle: 'Recycled bags' }
export const VALUE_LABEL = { local: 'Local', organic: 'Organic', recycle: 'Recycle' }
export const PRIORITY_LABEL = { cheapest: 'Cheapest', closest: 'Closest' }

// Typical UK prices (GBP) for the item as written on the list.
const BASE_PRICES = [
  [/tomato/i, 1.6],
  [/bread|sourdough|loaf/i, 3.2],
  [/egg/i, 2.4],
  [/milk/i, 1.85],
  [/spinach|salad|lettuce/i, 1.5],
  [/apple/i, 2.2],
  [/yog(h)?urt/i, 1.9],
  [/cheese/i, 3.5],
  [/banana/i, 1.1],
]

// Small deterministic hash so mock prices stay stable between renders.
function hash(str) {
  let h = 0
  for (const ch of str) h = (h * 31 + ch.charCodeAt(0)) >>> 0
  return h
}

function basePrice(description) {
  const hit = BASE_PRICES.find(([re]) => re.test(description))
  return hit ? hit[1] : 1.5 + (hash(description) % 250) / 100 // £1.50–£4.00 for unknown items
}

export const driveMinutes = (miles) => Math.max(2, Math.round(miles * 5 + 1))

/**
 * Options for one list item, scored against the user's preferences.
 * Returns options sorted best-first, each with { shop, price, score, matchedTags, label }.
 */
export function optionsForItem(item, prefs) {
  const maxDistance = parseFloat(prefs.maxDistance) || Infinity
  const nearby = SHOPS.filter((s) => s.distance <= maxDistance)
  // Each item is stocked by 3 of the nearby shops (stable per item).
  const skip = nearby.length > 3 ? hash(item.description) % nearby.length : -1
  const stocked = nearby.filter((_, i) => i !== skip)

  const base = basePrice(item.description)
  const options = stocked.map((shop) => {
    const wobble = 1 + ((hash(item.description + shop.id) % 11) - 5) / 100 // ±5%
    const price = Math.round(base * shop.priceFactor * wobble * 20) / 20 // nearest 5p
    return { shop, price }
  })
  if (options.length === 0) return []

  const cheapest = Math.min(...options.map((o) => o.price))
  const closest = Math.min(...options.map((o) => o.shop.distance))
  const priorityWeights = [3, 2]
  const valueWeights = [1, 0.6, 0.3] // values break ties; priorities lead

  for (const o of options) {
    let score = 0
    prefs.priorities.forEach((p, i) => {
      const fit = p === 'cheapest' ? cheapest / o.price : closest / o.shop.distance
      score += priorityWeights[i] * fit
    })
    o.matchedTags = prefs.values.map((v) => VALUE_TAG[v]).filter((t) => o.shop.tags.includes(t))
    prefs.values.forEach((v, i) => {
      if (o.shop.tags.includes(VALUE_TAG[v])) score += valueWeights[i]
    })
    o.score = score
    o.isCheapest = o.price === cheapest
    o.isClosest = o.shop.distance === closest
  }

  options.sort((a, b) => b.score - a.score)
  options.forEach((o, i) => {
    o.letter = 'ABC'[i]
    o.rank = i === 0 ? 'best' : i === 1 ? 'runner' : 'other'
    o.reason = o.isCheapest ? 'Cheapest' : o.isClosest ? 'Closest' : o.matchedTags[0] || null
  })
  return options
}

export const formatGBP = (n) => `£${n.toFixed(2)}`

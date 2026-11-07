// Shop options and the ranking logic used by pages 2 and 3.
// Shops and prices now come from GET /shopping-item/option-result (see
// getOptionResult in api/submitPreference.js); this file only adapts that data
// to the UI and scores it against the user's preferences.

import { toApiItem } from '../api/submitPreference.js'

// Which shop tag satisfies which "What matters to you?" value.
export const VALUE_TAG = { local: 'Local', organic: 'Organic', recycle: 'Recycled bags' }
export const VALUE_LABEL = { local: 'Local', organic: 'Organic', recycle: 'Recycle' }
export const PRIORITY_LABEL = { cheapest: 'Cheapest', closest: 'Closest' }

export const driveMinutes = (miles) =>
  Number.isFinite(miles) ? Math.max(2, Math.round(miles * 5 + 1)) : null

const norm = (s) => String(s ?? '').trim().toLowerCase()

/**
 * Find the API result for a list item and return its options.
 * Matches on the item name we sent to the API; if the API renamed the items,
 * falls back to list position when both lists have the same length.
 */
export function apiOptionsForItem(optionResult, items, item) {
  if (!optionResult?.length) return []

  const wanted = norm(toApiItem(item).item_name)
  const byName = optionResult.find((r) => norm(r.itemName) === wanted)
  if (byName) return byName.options

  if (optionResult.length === items.length) {
    return optionResult[items.indexOf(item)]?.options ?? []
  }
  return []
}

// Straight-line distance in miles between two { latitude, longitude } points.
// Returns null when either point has no coordinates.
export function haversineMiles(from, to) {
  const lat1 = from?.latitude
  const lon1 = from?.longitude
  const lat2 = to?.latitude
  const lon2 = to?.longitude
  if (![lat1, lon1, lat2, lon2].every(Number.isFinite)) return null

  const toRad = (d) => (d * Math.PI) / 180
  const dLat = toRad(lat2 - lat1)
  const dLon = toRad(lon2 - lon1)
  const a =
    Math.sin(dLat / 2) ** 2 + Math.cos(toRad(lat1)) * Math.cos(toRad(lat2)) * Math.sin(dLon / 2) ** 2
  const miles = 3958.8 * 2 * Math.asin(Math.sqrt(Math.min(1, Math.max(0, a)))) // clamp: guards bad coords
  return Math.round(miles * 10) / 10
}

// The API's boolean flags -> the tags the scoring and UI already use.
function tagsFor(o) {
  const tags = []
  if (o.isLocalBusiness) tags.push('Local')
  if (o.isOrganic) tags.push('Organic')
  if (o.isRecyclable) tags.push('Recycled bags')
  return tags
}

// API option -> the `shop` object the pages expect.
// The id is stable across items, so the same shop groups together.
function toShop(o, prefs) {
  const { name, latitude, longitude } = o.location
  return {
    id: `${name}|${latitude},${longitude}`.toLowerCase(),
    name,
    type: o.isLocalBusiness ? 'Local business' : 'Shop', // the API doesn't send a shop type
    distance: haversineMiles(prefs, o.location), // miles from the user, or null if unknown
    tags: tagsFor(o),
  }
}

/**
 * Options for one list item, scored against the user's preferences.
 * `apiOptions` are the normalised options from getOptionResult for this item.
 * Returns options sorted best-first, each with { shop, price, score, matchedTags, label }.
 *
 * The API already applies the maximum distance and the quantity, so no
 * distance filtering or quantity maths happens here; `price` is price_gbp as sent.
 */
export function optionsForItem(item, prefs, apiOptions = []) {
  const options = apiOptions.map((o) => ({
    key: o.key,
    optionId: o.optionId,
    productName: o.name, // e.g. "Diet coke - 330ml"
    shop: toShop(o, prefs),
    unitPrice: o.priceGbp,
    quantity: 1,
    price: o.priceGbp,
  }))
  if (options.length === 0) return []

  const cheapest = Math.min(...options.map((o) => o.price))
  const distances = options.map((o) => o.shop.distance).filter(Number.isFinite)
  const closest = distances.length ? Math.min(...distances) : null
  const priorityWeights = [3, 2]
  const valueWeights = [1, 0.6, 0.3] // values break ties; priorities lead

  for (const o of options) {
    let score = 0
    prefs.priorities.forEach((p, i) => {
      let fit
      if (p === 'cheapest') {
        fit = o.price > 0 ? cheapest / o.price : 1
      } else {
        // closest: no fit if we couldn't work out this shop's distance
        const d = o.shop.distance
        fit = Number.isFinite(d) && closest !== null ? (d === 0 ? 1 : closest / d) : 0
      }
      score += priorityWeights[i] * fit
    })
    o.matchedTags = prefs.values.map((v) => VALUE_TAG[v]).filter((t) => o.shop.tags.includes(t))
    prefs.values.forEach((v, i) => {
      if (o.shop.tags.includes(VALUE_TAG[v])) score += valueWeights[i]
    })
    o.score = score
    o.isCheapest = o.price === cheapest
    o.isClosest = closest !== null && o.shop.distance === closest
  }

  options.sort((a, b) => b.score - a.score)
  options.forEach((o, i) => {
    o.letter = String.fromCharCode(65 + i) // A, B, C...
    o.rank = i === 0 ? 'best' : i === 1 ? 'runner' : 'other'
    o.reason = o.isCheapest ? 'Cheapest' : o.isClosest ? 'Closest' : o.matchedTags[0] || null
  })
  return options
}

export const formatGBP = (n) => `£${n.toFixed(2)}`
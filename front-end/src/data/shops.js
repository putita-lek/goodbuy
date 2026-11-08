// Adapts the options from GET /shopping-item/option-result (see getOptionResult
// in api/submitPreference.js) to the shape pages 2 and 3 use.
//
// The API is the source of truth for prices, shops, distances and ranking: it
// already scores every option against the user's preferences (and how well the
// product matches the item), so the order it sends is kept here.

// Which shop tag satisfies which "What matters to you?" value.
export const VALUE_TAG = { local: 'Local', organic: 'Organic', recycle: 'Recyclable' }
export const VALUE_LABEL = { local: 'Local', organic: 'Organic', recycle: 'Recycle' }
export const PRIORITY_LABEL = { cheapest: 'Cheapest', closest: 'Closest' }

export const driveMinutes = (miles) =>
  Number.isFinite(miles) ? Math.max(2, Math.round(miles * 5 + 1)) : null

// "1.3 mi" or "— mi" when the distance is unknown.
export const formatMiles = (miles) =>
  Number.isFinite(miles) ? `${Math.round(miles * 10) / 10} mi` : '— mi'

/**
 * The API result for the list item at `index`. Results are numbered
 * result_id = 1, 2, 3… in the order the items were submitted, so the position
 * is used rather than the name (two items can share a name).
 */
export function apiOptionsForItem(optionResult, index) {
  if (!optionResult?.length) return []
  const result = optionResult.find((r) => r.resultId === index + 1) ?? optionResult[index]
  return result?.options ?? []
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

// The API's boolean flags -> the tags the UI shows.
function tagsFor(o) {
  const tags = []
  if (o.isLocalBusiness) tags.push(VALUE_TAG.local)
  if (o.isOrganic) tags.push(VALUE_TAG.organic)
  if (o.isRecyclable) tags.push(VALUE_TAG.recycle)
  return tags
}

// API option -> the `shop` object the pages expect.
// The id is stable across items, so the same shop groups together on page 3.
function toShop(o, userLocation) {
  const { name, latitude, longitude } = o.location
  return {
    id: `${name}|${latitude},${longitude}`.toLowerCase(),
    name,
    type: o.isLocalBusiness ? 'Local business' : 'Shop', // the API doesn't send a shop type
    latitude,
    longitude,
    // The API's distance is measured from the point it searched around; fall
    // back to our own calculation only if it is missing.
    distance: o.distanceMiles ?? haversineMiles(userLocation, o.location),
    tags: tagsFor(o),
  }
}

/**
 * Options for one list item, in the API's order (best first).
 * Each option is { key, optionId, productName, shop, unitPrice, quantity, price,
 * matchedTags, isCheapest, isClosest, letter, rank, reason }, where `price` is
 * the line total: the API's per-item price_gbp times the item quantity.
 */
export function optionsForItem(item, prefs, apiOptions = [], userLocation = null) {
  const quantity = Number(item.quantity) > 0 ? Number(item.quantity) : 1
  const options = apiOptions.map((o) => {
    const shop = toShop(o, userLocation)
    return {
      key: o.key,
      optionId: o.optionId,
      productName: o.name, // e.g. "Diet coke - 330ml"
      shop,
      unitPrice: o.priceGbp,
      quantity,
      price: Math.round(o.priceGbp * quantity * 100) / 100,
      matchedTags: prefs.values.map((v) => VALUE_TAG[v]).filter((t) => shop.tags.includes(t)),
    }
  })
  if (options.length === 0) return []

  const cheapest = Math.min(...options.map((o) => o.unitPrice))
  const distances = options.map((o) => o.shop.distance).filter(Number.isFinite)
  const closest = distances.length ? Math.min(...distances) : null

  options.forEach((o, i) => {
    o.isCheapest = o.unitPrice === cheapest
    o.isClosest = closest !== null && o.shop.distance === closest
    o.letter = String.fromCharCode(65 + i) // A, B, C...
    o.rank = i === 0 ? 'best' : i === 1 ? 'runner' : 'other'
    o.reason = o.isCheapest ? 'Cheapest' : o.isClosest ? 'Closest' : o.matchedTags[0] || null
  })
  return options
}

export const formatGBP = (n) => `£${(Number.isFinite(n) ? n : 0).toFixed(2)}`

import axios from 'axios'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL ?? 'http://localhost:8080',
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' },
})

// Map the ids used in the UI -> the keys the API expects.
// Adjust the left-hand side if your PreferencesCard uses different ids.
const PRIORITY_KEYS = {
  cheapest: 'best_price',
  closest: 'closest_location',
}
const SUSTAINABILITY_KEYS = {
  local: 'local_business',
  recycle: 'recycleable', // UI id (PreferencesCard) -> API spelling
  recyclable: 'recycleable',
  recycleable: 'recycleable',
  organic: 'organic',
}

// ['cheapest', 'closest'] -> { best_price: 1, closest_location: 2 }
function toRanking(ids, keyMap) {
  return ids
    .filter((id) => {
      if (keyMap[id]) return true
      console.warn(`submitPreference: unknown preference id "${id}" skipped`)
      return false
    })
    .reduce((acc, id, i) => ({ ...acc, [keyMap[id]]: i + 1 }), {})
}

// { description: 'Vine tomatoes', quantity: '500 g' }
//   -> { item_name: 'Vine tomatoes in g', item_quantity: 500 }
// The API wants a numeric quantity, so the unit is folded into the name.
export function toApiItem({ description, quantity }) {
  const name = description.trim()
  const match = String(quantity).trim().match(/^(\d+(?:\.\d+)?)\s*(.*)$/)

  if (!match) {
    // No leading number (e.g. "a few"): keep the text, default to 1
    const extra = String(quantity).trim()
    return { item_name: extra ? `${name} (${extra})` : name, item_quantity: 1 }
  }

  const [, amount, unit] = match
  return {
    item_name: unit ? `${name} in ${unit}` : name,
    item_quantity: parseFloat(amount),
  }
}

// Free-text location -> coordinates, using OpenStreetMap Nominatim.
// Fine for low volume; for production consider Google/Mapbox geocoding, or
// store lat/lng in `prefs` directly when the user picks a location.
export async function geocodeLocation(query) {
  const { data } = await axios.get('https://nominatim.openstreetmap.org/search', {
    params: { q: query, format: 'jsonv2', limit: 1, countrycodes: 'gb' },
    timeout: 10000,
  })

  if (!data.length) {
    throw new Error(`We couldn't find "${query}". Try a nearby postcode or area name.`)
  }
  return { latitude: parseFloat(data[0].lat), longitude: parseFloat(data[0].lon) }
}

export function buildPayload(prefs, items, userLocation) {
  const budget = parseFloat(prefs.budget)
  const distance = parseFloat(prefs.maxDistance)

  if (Number.isNaN(budget) || budget <= 0) throw new Error('Please enter a valid budget.')
  if (Number.isNaN(distance) || distance <= 0) throw new Error('Please enter a valid maximum distance.')

  return {
    user_location: userLocation,
    maximum_distance_from_user_location_miles: distance, // assumes the UI value is in miles
    preference_priority: toRanking(prefs.priorities, PRIORITY_KEYS),
    preference_sustainability: toRanking(prefs.values, SUSTAINABILITY_KEYS),
    budget_gbp: budget,
    shopping_item_list: items.map(toApiItem),
  }
}

// Turn any failure into a message that is safe to show the user.
function toFriendlyError(err) {
  if (!axios.isAxiosError(err)) return err.message || 'Something went wrong.'

  if (err.code === 'ECONNABORTED') return 'The request timed out. Please try again.'

  if (err.response) {
    const { status, data } = err.response
    // Use the server's message when it sends a string one
    const serverMessage = [data?.message, data?.detail, data?.error].find(
      (m) => typeof m === 'string',
    )
    if (serverMessage) return serverMessage
    if (status === 400 || status === 422) return 'Some of your details look invalid. Please check them and try again.'
    if (status === 404) return "We couldn't find results for this session. Please go back and start again."
    if (status === 429) return 'Too many requests. Please wait a moment and try again.'
    if (status >= 500) return 'Our server had a problem. Please try again shortly.'
    return `Request failed (${status}).`
  }

  // Request was sent but no response came back
  return 'Unable to reach the server. Check your connection and try again.'
}

// Normalise one option from the API into the shape the UI uses.
// Note: the API spells the local flag "is_local_businesss" (three s's); both
// spellings are accepted so a future fix on the server won't break this.
function toUiOption(option, resultId, index) {
  return {
    key: `${resultId}-${index}`, // option_id repeats within an item, so don't use it as a React key
    optionId: option.option_id,
    name: option.option_name,
    priceGbp: option.price_gbp,
    location: {
      name: option.location?.name ?? 'Unknown shop',
      latitude: option.location?.latitude ?? null,
      longitude: option.location?.longitude ?? null,
    },
    distanceMiles: Number.isFinite(option.distance_miles) ? option.distance_miles : null,
    isRecyclable: Boolean(option.is_recycleable),
    isLocalBusiness: Boolean(option.is_local_businesss ?? option.is_local_business),
    isOrganic: Boolean(option.is_organic),
  }
}

// GET /shopping-item/option-result?session_id=... -> resolves with
// [{ resultId, itemName, itemQuantity, options: [...] }]
export async function getOptionResult(sessionId) {
  try {
    if (sessionId == null) throw new Error('Missing session. Please go back and submit your list again.')

    const { data } = await api.get('/shopping-item/option-result', {
      params: { session_id: sessionId },
    })

    if (!Array.isArray(data?.option_result)) throw new Error('Unexpected response from the server.')
    if (data.option_result.length === 0) {
      throw new Error('No matching shops found. Try a larger distance or a different list.')
    }

    return data.option_result.map((r) => ({
      resultId: r.result_id,
      itemName: r.item_name,
      itemQuantity: r.item_quantity,
      options: (r.options ?? []).map((o, i) => toUiOption(o, r.result_id, i)),
    }))
  } catch (err) {
    throw new Error(toFriendlyError(err))
  }
}

// POST /submit-preference -> resolves with { sessionId, userLocation }.
// userLocation is the point the search used (typed place, geocoded, or GPS), so
// the review and summary pages can show distances and the route from it.
export async function submitPreference(prefs, items) {
  try {
    if (!items.length) throw new Error('Add at least one item to your list.')

    let userLocation
    if (prefs.latitude != null && prefs.longitude != null) {
      userLocation = { latitude: Number(prefs.latitude), longitude: Number(prefs.longitude) }
    } else if (prefs.location.trim()) {
      userLocation = await geocodeLocation(prefs.location.trim())
    } else {
      throw new Error('Please enter your location, or use your current location.')
    }
    const payload = buildPayload(prefs, items, userLocation)

    const { data } = await api.post('/submit-preference', payload)

    if (data?.session_id == null) throw new Error('Unexpected response from the server.')
    return { sessionId: data.session_id, userLocation }
  } catch (err) {
    throw new Error(toFriendlyError(err))
  }
}

// POST /shopping-item/submit-option: saves the option chosen for each item.
// choices: [{ resultId, optionId }]
export async function submitOption(sessionId, choices) {
  try {
    if (sessionId == null) throw new Error('Missing session. Please go back and submit your list again.')
    if (!choices.length) throw new Error('Choose at least one option first.')

    const { data } = await api.post('/shopping-item/submit-option', {
      session_id: sessionId,
      submitted_option: choices.map((c) => ({ result_id: c.resultId, selected_option_id: c.optionId })),
    })
    if (!data?.is_success) throw new Error(data?.error || 'Could not save your choices.')
  } catch (err) {
    throw new Error(toFriendlyError(err))
  }
}
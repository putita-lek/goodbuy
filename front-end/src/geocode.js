// Turn what the user typed in "Your location" into latitude/longitude using
// postcodes.io (free, no API key, UK only). Tries, in order:
//   1. full postcode   "EH1 1YZ"          -> /postcodes/EH11YZ
//   2. postcode district "EH6"            -> /outcodes/EH6
//   3. place name      "Leith, Edinburgh" -> /places?q=Leith
// Returns { latitude, longitude, label } or null when nothing matches.

const BASE = 'https://api.postcodes.io'

const FULL_POSTCODE = /^[A-Z]{1,2}\d[A-Z\d]?\s*\d[A-Z]{2}$/i
const OUTCODE = /^[A-Z]{1,2}\d[A-Z\d]?$/i

async function get(path, signal) {
  const res = await fetch(`${BASE}${path}`, { signal })
  if (res.status === 404) return null
  if (!res.ok) throw new Error(`postcodes.io error ${res.status}`)
  return (await res.json()).result
}

export async function geocodeUK(text, signal) {
  const query = text.trim()
  if (query.length < 2) return null

  if (FULL_POSTCODE.test(query)) {
    const r = await get(`/postcodes/${encodeURIComponent(query.replace(/\s+/g, ''))}`, signal)
    return r && { latitude: r.latitude, longitude: r.longitude, label: r.postcode }
  }

  if (OUTCODE.test(query)) {
    const r = await get(`/outcodes/${encodeURIComponent(query)}`, signal)
    if (r) return { latitude: r.latitude, longitude: r.longitude, label: r.outcode }
  }

  // Place names: only the part before the first comma ("Leith, Edinburgh" -> "Leith").
  const place = query.split(',')[0].trim()
  const results = await get(`/places?q=${encodeURIComponent(place)}&limit=1`, signal)
  const r = results?.[0]
  return r ? { latitude: r.latitude, longitude: r.longitude, label: `${r.name_1}, ${r.county_unitary}` } : null
}

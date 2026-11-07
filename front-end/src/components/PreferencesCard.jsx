import { useEffect, useState } from 'react'
import { Leaf, LocateFixed, MapPin, Recycle, Store, Wallet } from 'lucide-react'
import { Card, Field, Label, OptionCard, SectionTitle } from './ui.jsx'
import { geocodeUK } from '../geocode.js'

const PRIORITIES = [
  { id: 'cheapest', icon: Wallet, title: 'Cheapest', description: 'Lower total basket cost' },
  { id: 'closest', icon: MapPin, title: 'Closest', description: 'Shortest travel distance' },
]

const VALUES = [
  { id: 'local', icon: Store, title: 'Local market', description: 'Independent shops & local growers' },
  { id: 'organic', icon: Leaf, title: 'Organic', description: 'More organically grown choices' },
  { id: 'recycle', icon: Recycle, title: 'Recycle', description: 'Less waste, recyclable packaging' },
]

// Toggle an id in an ordered list. Position in the list = rank (1 first).
export function toggleRanked(list, id) {
  return list.includes(id) ? list.filter((x) => x !== id) : [...list, id]
}

const rankOf = (list, id) => (list.includes(id) ? list.indexOf(id) + 1 : null)

const CURRENT_LOCATION = 'Current location'

export default function PreferencesCard({ prefs, onChange }) {
  // idle | loading (GPS) | denied | searching (postcodes.io) | found | notfound | error
  const [geoStatus, setGeoStatus] = useState('idle')
  const hasCoords = prefs.latitude != null && prefs.longitude != null

  // Typed location -> lat/long via postcodes.io, 600 ms after the user stops typing.
  useEffect(() => {
    const text = prefs.location.trim()
    if (!text || text === CURRENT_LOCATION || hasCoords) return
    const controller = new AbortController()
    const timer = setTimeout(async () => {
      setGeoStatus('searching')
      try {
        const hit = await geocodeUK(text, controller.signal)
        if (!hit) return setGeoStatus('notfound')
        set({ latitude: hit.latitude, longitude: hit.longitude })
        setGeoStatus('found')
      } catch (e) {
        if (e.name !== 'AbortError') setGeoStatus('error')
      }
    }, 600)
    return () => {
      clearTimeout(timer)
      controller.abort()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [prefs.location])

  // Functional update so rapid clicks never work from stale state.
  const set = (patch) =>
    onChange((p) => ({ ...p, ...(typeof patch === 'function' ? patch(p) : patch) }))

  // Browser geolocation: asks the user for permission, then stores lat/long
  // (sent to the backend as user_location). The text field just shows a label.
  const useMyLocation = () => {
    if (!navigator.geolocation) return setGeoStatus('denied')
    setGeoStatus('loading')
    navigator.geolocation.getCurrentPosition(
      ({ coords }) => {
        set({ location: CURRENT_LOCATION, latitude: coords.latitude, longitude: coords.longitude })
        setGeoStatus('idle')
      },
      () => setGeoStatus('denied'),
      { enableHighAccuracy: false, timeout: 10000, maximumAge: 60000 },
    )
  }

  return (
    <Card className="p-7">
      {/* 1. Priority options */}
      <SectionTitle number={1} title="Choose your priority options" />
      <p className="mt-6 text-[15px] text-muted">Select one or both</p>
      <div className="mt-3 grid gap-4 sm:grid-cols-2">
        {PRIORITIES.map((o) => (
          <OptionCard
            key={o.id}
            {...o}
            rank={rankOf(prefs.priorities, o.id)}
            onClick={() => set((p) => ({ priorities: toggleRanked(p.priorities, o.id) }))}
          />
        ))}
      </div>

      <div className="mt-7 flex flex-wrap items-baseline justify-between gap-2">
        <h3 className="text-[15px] font-semibold">What matters to you?</h3>
        <span className="text-sm text-muted">
          1 is highest priority · {prefs.values.length} selected
        </span>
      </div>
      <div className="mt-3 grid gap-4 sm:grid-cols-3">
        {VALUES.map((o) => (
          <OptionCard
            key={o.id}
            {...o}
            rank={rankOf(prefs.values, o.id)}
            onClick={() => set((p) => ({ values: toggleRanked(p.values, o.id) }))}
          />
        ))}
      </div>
      <p className="mt-4 text-sm text-muted">
        Higher numbers rank lower. We’ll favour the closest match within your limits,
        prioritising 1 first, then 2, then 3.
      </p>

      {/* 2. Shopping preferences */}
      <div className="mt-8">
        <SectionTitle number={2} title="Set your shopping preferences" />
        <p className="mt-2 text-[15px] text-muted">
          Choose your starting point and what you’d like us to look for.
        </p>
      </div>

      <div className="mt-8 grid gap-5 sm:grid-cols-[1fr_240px]">
        <div>
          <Label htmlFor="location">Your location</Label>
          <Field
            id="location"
            leading={<MapPin className="size-5" strokeWidth={1.6} />}
            value={prefs.location}
            // Typing a new place clears coordinates from an earlier "current location".
            onChange={(e) => {
              set({ location: e.target.value, latitude: null, longitude: null })
              setGeoStatus('idle')
            }}
            placeholder="Town or postcode"
          />
          <div className="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-sm">
            <button
              type="button"
              onClick={useMyLocation}
              disabled={geoStatus === 'loading'}
              className="flex items-center gap-1.5 font-semibold text-brand-600 hover:underline disabled:opacity-60"
            >
              <LocateFixed className="size-4" strokeWidth={1.8} />
              {geoStatus === 'loading' ? 'Finding your location…' : 'Use my current location'}
            </button>
            {geoStatus === 'denied' ? (
              <span className="text-red-700">Location access was blocked. Allow it in your browser settings.</span>
            ) : geoStatus === 'searching' ? (
              <span className="text-muted">Looking up “{prefs.location.trim()}”…</span>
            ) : geoStatus === 'notfound' ? (
              <span className="text-red-700">Couldn’t find that place. Try a UK postcode like EH1 1YZ.</span>
            ) : geoStatus === 'error' ? (
              <span className="text-red-700">Location lookup is unavailable right now. Use your current location.</span>
            ) : null}
          </div>
        </div>
        <div>
          <Label htmlFor="distance">Maximum distance</Label>
          <Field
            id="distance"
            type="number"
            min="0"
            step="0.5"
            suffix="miles"
            value={prefs.maxDistance}
            onChange={(e) => set({ maxDistance: e.target.value })}
          />
        </div>
      </div>

      <hr className="my-7 border-line" />

      <div className="grid items-end gap-5 sm:grid-cols-[1fr_240px]">
        <div className="pb-2">
          <h3 className="text-[17px] font-semibold">Keep your basket within budget</h3>
          <p className="mt-1 text-[15px] text-muted">
            The maximum you want to spend on this shopping trip.
          </p>
        </div>
        <div>
          <Label htmlFor="budget">Maximum budget</Label>
          <Field
            id="budget"
            inputMode="decimal"
            leading={<span className="text-xl text-ink">£</span>}
            value={prefs.budget}
            onChange={(e) => set({ budget: e.target.value })}
          />
        </div>
      </div>
    </Card>
  )
}

import { useEffect, useRef } from 'react'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'

// Real map (Leaflet + OpenStreetMap tiles — free, no API key).
//
// Props:
//   user:  { latitude, longitude } or null  -> blue "You" dot
//   stops: [{ name, latitude, longitude }]  -> numbered green pins, in visiting order
// Draws a dashed route You -> 1 -> 2 -> … and zooms to fit everything.

const pinIcon = (label, n) =>
  L.divIcon({
    className: '',
    iconSize: null,
    iconAnchor: [14, 14],
    html: `<div style="display:flex;align-items:center;gap:6px;transform:translateY(-2px)">
      <span style="display:grid;place-items:center;width:28px;height:28px;border-radius:999px;background:#2e6b45;color:#fff;font:600 13px Inter,sans-serif;border:2px solid #fff;box-shadow:0 1px 4px rgba(0,0,0,.3)">${n}</span>
      <span style="background:#fff;border-radius:6px;padding:2px 8px;font:600 12px Inter,sans-serif;color:#1f2a24;white-space:nowrap;box-shadow:0 1px 4px rgba(0,0,0,.2)">${label}</span>
    </div>`,
  })

const youIcon = L.divIcon({
  className: '',
  iconSize: [18, 18],
  iconAnchor: [9, 9],
  html: '<span style="display:block;width:18px;height:18px;border-radius:999px;background:#1565c0;border:3px solid #fff;box-shadow:0 0 0 4px rgba(21,101,192,.25)"></span>',
})

export default function RouteMap({ user, stops, className = 'h-72' }) {
  const el = useRef(null)
  // Rebuild the map only when the coordinates actually change.
  const key = JSON.stringify({ user, stops })

  useEffect(() => {
    if (!el.current) return
    const map = L.map(el.current, { scrollWheelZoom: false, attributionControl: true })
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      maxZoom: 19,
      attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>',
    }).addTo(map)

    const route = []
    if (user) {
      const p = [user.latitude, user.longitude]
      L.marker(p, { icon: youIcon, title: 'You' }).addTo(map).bindTooltip('You', { direction: 'top' })
      route.push(p)
    }
    stops.forEach((s, i) => {
      const p = [s.latitude, s.longitude]
      L.marker(p, { icon: pinIcon(s.name, i + 1), title: s.name }).addTo(map)
      route.push(p)
    })
    if (route.length > 1) {
      L.polyline(route, { color: '#2e6b45', weight: 3, dashArray: '6 8', opacity: 0.8 }).addTo(map)
    }

    if (route.length === 1) map.setView(route[0], 15)
    else if (route.length > 1) map.fitBounds(route, { padding: [40, 40], maxZoom: 16 })
    else map.setView([55.9533, -3.1883], 12) // nothing to show: default to Edinburgh

    return () => map.remove()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [key])

  return <div ref={el} className={`w-full overflow-hidden rounded-xl ${className}`} />
}

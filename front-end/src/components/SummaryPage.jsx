import { useState } from 'react'
import { ArrowLeft, Check, Download, Share2 } from 'lucide-react'
import PageIntro from './PageIntro.jsx'
import ItemIcon from './ItemIcon.jsx'
import { Card } from './ui.jsx'
import RouteMap from './RouteMap.jsx'
import { Tag } from './ReviewPage.jsx'
import { VALUE_LABEL, VALUE_TAG, formatGBP, formatMiles } from '../data/shops.js'

// Group the chosen options by shop; the shop covering most items comes first.
function groupByShop(plan) {
  const groups = new Map()
  for (const { item, chosen } of plan) {
    if (!chosen) continue
    const g = groups.get(chosen.shop.id) ?? { shop: chosen.shop, lines: [], subtotal: 0 }
    g.lines.push({ item, option: chosen })
    g.subtotal = Math.round((g.subtotal + chosen.price) * 100) / 100
    groups.set(chosen.shop.id, g)
  }
  const dist = (g) => (Number.isFinite(g.shop.distance) ? g.shop.distance : Infinity)
  return [...groups.values()].sort((a, b) => b.lines.length - a.lines.length || dist(a) - dist(b))
}

// 0–5: share of chosen options that carry the tags the user cares about.
function sustainabilityScore(plan, prefs) {
  const tags = prefs.values.map((v) => VALUE_TAG[v])
  const chosen = plan.filter((p) => p.chosen)
  if (!tags.length || !chosen.length) return null
  const hits = chosen.reduce((n, p) => n + tags.filter((t) => p.chosen.shop.tags.includes(t)).length, 0)
  return Math.round((5 * hits) / (chosen.length * tags.length))
}

function listAsText(groups, total) {
  const lines = ['GoodBuy – shopping plan', '']
  for (const g of groups) {
    lines.push(`${g.shop.name} (${formatMiles(g.shop.distance)})`)
    for (const { item, option } of g.lines)
      lines.push(`  [ ] ${item.description} ×${item.quantity} – ${formatGBP(option.price)}`)
    lines.push(`  Subtotal: ${formatGBP(g.subtotal)}`, '')
  }
  lines.push(`Total: ${formatGBP(total)}`)
  return lines.join('\n')
}

const join = (names) =>
  names.length <= 2 ? names.join(' + ') : `${names.slice(0, -1).join(', ')} + ${names.at(-1)}`

function Stat({ label, value, sub }) {
  return (
    <Card className="p-6">
      <p className="text-xs font-medium tracking-[0.12em] text-muted uppercase">{label}</p>
      <p className="mt-3 text-[28px] font-bold tracking-tight">{value}</p>
      <p className="mt-2 text-sm text-muted">{sub}</p>
    </Card>
  )
}

function ShopCard({ group, isMain }) {
  return (
    <Card className="flex flex-col p-6">
      <div className="flex items-start justify-between gap-3">
        <h2 className="text-xl font-semibold tracking-tight">
          You go to {group.shop.name} ({formatMiles(group.shop.distance)}) for:
        </h2>
        {isMain && (
          <span className="shrink-0 rounded-full bg-brand-600 px-2.5 py-1 text-xs font-semibold text-white">
            Recommended
          </span>
        )}
      </div>
      <p className="mt-1 text-sm text-muted">
        {group.lines.length} {group.lines.length === 1 ? 'item' : 'items'} · {group.shop.type}
      </p>

      <table className="mt-6 w-full text-left">
        <thead>
          <tr className="text-[11px] font-medium tracking-wider text-muted uppercase">
            <th className="pb-3 font-medium">Item</th>
            <th className="pb-3 text-center font-medium">Qty</th>
            <th className="pb-3 text-right font-medium">Total</th>
          </tr>
        </thead>
        <tbody>
          {group.lines.map(({ item, option }) => (
            <tr key={item.id} className="border-b border-line last:border-b-0">
              <td className="py-3">
                <div className="flex items-center gap-3">
                  <span className="grid size-9 shrink-0 place-items-center rounded-lg bg-page text-brand-600">
                    <ItemIcon description={item.description} />
                  </span>
                  <div>
                    <p className="text-[15px] font-medium">{item.description}</p>
                    {option.matchedTags.length > 0 && (
                      <div className="mt-1.5 flex flex-wrap gap-1">
                        {option.matchedTags.map((t) => (
                          <Tag key={t}>{t}</Tag>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
              </td>
              <td className="py-3 text-center text-[15px] whitespace-nowrap">{item.quantity}</td>
              <td className="py-3 text-right text-[15px] font-semibold">{formatGBP(option.price)}</td>
            </tr>
          ))}
        </tbody>
      </table>

      <p className="mt-auto rounded-lg bg-brand-50 px-3 py-2 text-sm font-semibold text-brand-700">
        Subtotal {formatGBP(group.subtotal)}
      </p>
    </Card>
  )
}

// Real map of the trip: You -> shop 1 -> shop 2 …, using the coordinates the
// API sent for each store. Stores without coordinates are left off the map.
function RoutePreview({ groups, userLocation }) {
  const user =
    Number.isFinite(userLocation?.latitude) && Number.isFinite(userLocation?.longitude)
      ? { latitude: userLocation.latitude, longitude: userLocation.longitude }
      : null
  const stops = groups
    .filter((g) => Number.isFinite(g.shop.latitude) && Number.isFinite(g.shop.longitude))
    .map((g) => ({ name: g.shop.name, latitude: g.shop.latitude, longitude: g.shop.longitude }))

  return (
    <Card className="p-6">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold">Route preview</h2>
        <span className="text-sm text-muted">
          {groups.length} {groups.length === 1 ? 'stop' : 'stops'}
        </span>
      </div>
      <RouteMap user={user} stops={stops} className="mt-5 h-72" />
      {!user && (
        <p className="mt-3 text-[13px] text-muted">
          Add your location on step 1 to see the route from where you are.
        </p>
      )}
    </Card>
  )
}

export default function SummaryPage({ plan, prefs, userLocation, total, onBack }) {
  const [copied, setCopied] = useState(false)
  const groups = groupByShop(plan)
  const budget = parseFloat(prefs.budget) || 0
  const remaining = budget - total
  const usedPct = budget ? (total / budget) * 100 : 0
  const score = sustainabilityScore(plan, prefs)
  const itemCount = plan.filter((p) => p.chosen).length

  const download = () => {
    const blob = new Blob([listAsText(groups, total)], { type: 'text/plain' })
    const a = Object.assign(document.createElement('a'), {
      href: URL.createObjectURL(blob),
      download: 'goodbuy-list.txt',
    })
    a.click()
    // Revoking straight away can cancel the download in some browsers.
    setTimeout(() => URL.revokeObjectURL(a.href), 1000)
  }

  const share = async () => {
    const text = listAsText(groups, total)
    try {
      if (navigator.share) await navigator.share({ title: 'My GoodBuy plan', text })
      else {
        await navigator.clipboard.writeText(text)
        setCopied(true)
        setTimeout(() => setCopied(false), 2000)
      }
    } catch {
      /* share cancelled */
    }
  }

  if (!groups.length)
    return (
      <Card className="mt-14 p-10">
        <h1 className="text-2xl font-semibold">No items selected yet</h1>
        <button type="button" onClick={onBack} className="mt-6 font-semibold text-brand-600">
          ← Back to options
        </button>
      </Card>
    )

  const stops = groups.length === 1 ? 'One stop' : `${groups.length} stops`
  const budgetNote = remaining >= 0 ? 'within your budget' : 'but a little over your budget'
  const missing = plan.length - itemCount
  const coverage = missing === 0 ? 'everything on your list' : `${itemCount} of ${plan.length} items`

  return (
    <>
      <PageIntro
        eyebrow="Step 3 of 3 · Ready to shop"
        title="Your shopping plan"
        subtitle={`${stops}, ${coverage}, ${budgetNote}.`}
      />

      <div className="mt-10 grid gap-5 md:grid-cols-3">
        <Stat label="Total cost" value={formatGBP(total)} sub="Estimated basket total" />
        <Stat label="Stores to visit" value={groups.length} sub={join(groups.map((g) => g.shop.name))} />
        <Stat
          label="Sustainability score"
          value={score === null ? '—' : `${score}/5 leaves`}
          sub={
            prefs.values.length
              ? `${prefs.values.map((v) => VALUE_LABEL[v]).join(', ')} choices`
              : 'No sustainability preferences set'
          }
        />
      </div>

      <div className="mt-6 grid items-start gap-5 md:grid-cols-2">
        {groups.map((g, i) => (
          <ShopCard key={g.shop.id} group={g} isMain={i === 0 && groups.length > 1} />
        ))}
      </div>

      <Card className="mt-6 p-6">
        <div className="flex items-center justify-between">
          <h2 className="text-xl font-semibold">Budget</h2>
          <span className={`font-semibold ${remaining >= 0 ? 'text-brand-600' : 'text-red-600'}`}>
            {remaining >= 0 ? `${formatGBP(remaining)} remaining` : `${formatGBP(-remaining)} over budget`}
          </span>
        </div>
        <div className="mt-4 h-2 overflow-hidden rounded-full bg-page">
          <div
            className={`h-full rounded-full ${remaining >= 0 ? 'bg-brand-600' : 'bg-red-500'}`}
            style={{ width: `${Math.min(100, usedPct)}%` }}
          />
        </div>
        <div className="mt-3 flex justify-between text-[13px] text-muted">
          <span>{usedPct.toFixed(1)}% of budget used</span>
          <span>{formatGBP(budget)}</span>
        </div>
      </Card>

      <div className="mt-6 grid gap-4 sm:grid-cols-[200px_1fr_1fr]">
        <button
          type="button"
          onClick={onBack}
          className="flex items-center justify-center gap-2 rounded-lg border border-line bg-white px-5 py-3 font-semibold hover:border-brand-500"
        >
          <ArrowLeft className="size-4" /> Edit choices
        </button>
        <button
          type="button"
          onClick={download}
          className="flex items-center justify-center gap-2 rounded-lg bg-brand-600 px-5 py-3 font-semibold text-white hover:bg-brand-700"
        >
          <Download className="size-4" /> Download list
        </button>
        <button
          type="button"
          onClick={share}
          className="flex items-center justify-center gap-2 rounded-lg bg-brand-600 px-5 py-3 font-semibold text-white hover:bg-brand-700"
        >
          {copied ? <Check className="size-4" /> : <Share2 className="size-4" />}
          {copied ? 'Copied to clipboard' : 'Share'}
        </button>
      </div>

      <div className="mt-6">
        <RoutePreview groups={groups} userLocation={userLocation} />
      </div>

      <p className="mt-4 text-[13px] text-muted">
        {itemCount} {itemCount === 1 ? 'item' : 'items'} · prices are estimates — check availability in store before you travel.
      </p>
    </>
  )
}

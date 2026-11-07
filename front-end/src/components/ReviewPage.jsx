import { ArrowLeft, ArrowRight, CircleCheck, Info, Leaf, MapPin, SlidersHorizontal, Sparkles } from 'lucide-react'
import PageIntro from './PageIntro.jsx'
import ItemIcon from './ItemIcon.jsx'
import { PRIORITY_LABEL, VALUE_LABEL, driveMinutes, formatGBP } from '../data/shops.js'

function Chip({ icon: Icon, children }) {
  return (
    <span className="inline-flex items-center gap-1.5 rounded-full bg-brand-50 px-3 py-1.5 text-sm font-medium text-brand-700">
      {Icon && <Icon className="size-4" strokeWidth={1.8} />}
      {children}
    </span>
  )
}

export function Tag({ children }) {
  return (
    <span className="rounded border border-brand-100 bg-brand-50 px-2 py-0.5 text-xs font-medium text-brand-700">
      {children}
    </span>
  )
}

// Label strip on top of each option card: Best match / Runner up / shop type.
function CardLabel({ option }) {
  if (option.rank === 'best')
    return (
      <div className="flex items-center gap-1.5 bg-brand-600 px-3.5 py-1.5 text-xs font-semibold text-white">
        <Sparkles className="size-3.5" /> Best match
      </div>
    )
  if (option.rank === 'runner')
    return (
      <div className="flex items-center gap-1.5 bg-brand-100 px-3.5 py-1.5 text-xs font-semibold text-brand-700">
        Runner up{option.reason && <span className="font-medium opacity-80">· {option.reason}</span>}
      </div>
    )
  return <div className="bg-page px-3.5 py-1.5 text-xs font-medium text-muted">{option.shop.type}</div>
}

function OptionCard({ option, selected, onSelect }) {
  const { shop } = option
  return (
    <button
      type="button"
      onClick={onSelect}
      aria-pressed={selected}
      className={`overflow-hidden rounded-xl border-2 bg-white text-left transition ${
        selected ? 'border-brand-600 shadow-sm' : 'border-line hover:border-brand-500/50'
      }`}
    >
      <CardLabel option={option} />
      <div className="p-4">
        <div className="flex items-center gap-2.5">
          <span className="grid size-5 place-items-center rounded-full bg-page text-[10px] font-semibold text-muted">
            {option.letter}
          </span>
          <span className="flex-1 text-[15px] font-semibold">{shop.name}</span>
          {selected ? (
            <CircleCheck className="size-5 text-brand-600" strokeWidth={1.8} />
          ) : (
            <span className="size-5 rounded-full border-[1.5px] border-muted/50" />
          )}
        </div>
        <p className="mt-3 text-[28px] font-bold tracking-tight">{formatGBP(option.price)}</p>
        <p className="mt-2 flex items-center gap-1.5 text-[13px] text-muted">
          <MapPin className="size-3.5" strokeWidth={1.8} />
          {shop.distance} mi · {driveMinutes(shop.distance)} min drive
        </p>
        <div className="mt-4 flex flex-wrap gap-1.5">
          {shop.tags.map((t) => (
            <Tag key={t}>{t}</Tag>
          ))}
        </div>
      </div>
    </button>
  )
}

export default function ReviewPage({ plan, prefs, total, onSelect, onBack, onNext }) {
  const budget = parseFloat(prefs.budget) || 0
  const remaining = budget - total
  const usedPct = budget ? Math.min(100, (total / budget) * 100) : 0
  const missing = plan.filter((p) => !p.chosen).length

  return (
    <>
      <PageIntro
        eyebrow="Step 2 of 3 · Find your fit"
        title="We found options for your items"
        subtitle="Choose the best option for each item. We’ll build your basket from your selections."
      />

      <div className="mt-8 flex flex-wrap items-center gap-3 rounded-xl border border-line bg-white px-4 py-3">
        <span className="mr-1 text-xs font-medium tracking-[0.12em] text-muted uppercase">Your preferences</span>
        {prefs.priorities.length > 0 && (
          <Chip icon={SlidersHorizontal}>{prefs.priorities.map((p) => PRIORITY_LABEL[p]).join(' › ')}</Chip>
        )}
        {prefs.values.length > 0 && (
          <Chip icon={Leaf}>{prefs.values.map((v) => VALUE_LABEL[v]).join(' › ')}</Chip>
        )}
        <Chip>Within {prefs.maxDistance} mi</Chip>
        <Chip>Budget {formatGBP(budget)}</Chip>
        <button type="button" onClick={onBack} className="ml-auto text-sm font-semibold text-brand-600 hover:underline">
          Edit
        </button>
      </div>

      <div className="mt-8 space-y-9">
        {plan.map(({ item, options, chosen }) => (
          <section key={item.id}>
            <div className="mb-3 flex items-center gap-3">
              <span className="grid size-10 place-items-center rounded-lg bg-brand-50 text-brand-600">
                <ItemIcon description={item.description} />
              </span>
              <div>
                <h2 className="text-lg font-semibold">{item.description}</h2>
                <p className="text-[13px] text-muted">
                  {item.quantity}
                  {options.length > 0 && ` · from ${formatGBP(Math.min(...options.map((o) => o.price)))}`}
                </p>
              </div>
            </div>
            {options.length === 0 ? (
              <p className="rounded-xl border border-dashed border-line bg-white p-5 text-sm text-muted">
                No shops stock this within {prefs.maxDistance} miles. Try increasing your maximum distance.
              </p>
            ) : (
              <div className="grid gap-4 md:grid-cols-3">
                {options.map((o) => (
                  <OptionCard
                    key={o.shop.id}
                    option={o}
                    selected={chosen?.shop.id === o.shop.id}
                    onSelect={() => onSelect(item.id, o.shop.id)}
                  />
                ))}
              </div>
            )}
          </section>
        ))}
      </div>

      <p className="mt-8 flex items-center gap-2 text-[13px] text-muted">
        <Info className="size-4" strokeWidth={1.6} />
        Illustrative prices for your quantities. Availability and final in-store prices may vary.
      </p>

      {/* Sticky total bar */}
      <div className="sticky bottom-0 z-10 mx-[calc(50%-50vw)] mt-10 border-t border-line bg-white/95 py-4 backdrop-blur">
        <div className="mx-auto flex max-w-[1508px] flex-wrap items-center justify-between gap-4 px-6">
          <button
            type="button"
            onClick={onBack}
            className="flex items-center gap-2 rounded-lg border border-line bg-white px-5 py-3 font-semibold hover:border-brand-500"
          >
            <ArrowLeft className="size-5" /> Back
          </button>

          <div className="min-w-[280px] text-center">
            <p className="flex flex-wrap items-center justify-center gap-2">
              <span className="text-lg font-semibold">Total: {formatGBP(total)}</span>
              <span className="text-muted">/ {formatGBP(budget)} budget</span>
              <span
                className={`rounded-full px-2.5 py-0.5 text-xs font-semibold ${
                  remaining >= 0 ? 'bg-brand-50 text-brand-700' : 'bg-red-50 text-red-700'
                }`}
              >
                {remaining >= 0 ? `${formatGBP(remaining)} remaining` : `${formatGBP(-remaining)} over budget`}
              </span>
            </p>
            <div className="mx-auto mt-2 h-1.5 w-72 overflow-hidden rounded-full bg-line">
              <div
                className={`h-full rounded-full ${remaining >= 0 ? 'bg-brand-600' : 'bg-red-500'}`}
                style={{ width: `${usedPct}%` }}
              />
            </div>
          </div>

          <button
            type="button"
            onClick={onNext}
            disabled={missing === plan.length}
            className="flex items-center gap-2 rounded-lg bg-brand-600 px-6 py-3 font-semibold text-white hover:bg-brand-700 disabled:opacity-50"
          >
            See summary <ArrowRight className="size-5" />
          </button>
        </div>
      </div>
    </>
  )
}

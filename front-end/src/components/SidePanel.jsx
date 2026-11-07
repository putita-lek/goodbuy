import { ListChecks, Lock, SlidersHorizontal, Wallet } from 'lucide-react'
import { Card } from './ui.jsx'

const POINTS = [
  { icon: Wallet, title: 'Your limits come first', text: 'We stay within your budget and the distance you choose.' },
  { icon: SlidersHorizontal, title: 'Preferences guide the match', text: 'We rank nearby shops by your priority order: 1 first, then 2, then 3.' },
  { icon: ListChecks, title: 'You’re always in control', text: 'Review every item before we recommend a shop.' },
]

export default function SidePanel() {
  return (
    <aside className="flex flex-col gap-6">
      <div className="overflow-hidden rounded-2xl bg-sand">
        <img
          src="/basket.jpg"
          alt="Wicker basket of fresh vegetables in a local grocery shop"
          className="aspect-[442/298] w-full object-cover"
        />
        <div className="p-7">
          <p className="text-xs font-semibold tracking-[0.15em] text-brand-600 uppercase">
            Small choices. Good things.
          </p>
          <h2 className="mt-3 text-[28px] leading-tight tracking-tight">
            Better for your basket. Kinder to your neighbourhood.
          </h2>
          <p className="mt-4 text-[15px] leading-relaxed text-muted">
            Discover fresh essentials from the independent shops around you.
          </p>
        </div>
      </div>

      <Card className="p-7">
        <h2 className="text-xl font-semibold">Your shop, your way</h2>
        <ul className="mt-5 space-y-6">
          {POINTS.map(({ icon: Icon, title, text }) => (
            <li key={title} className="flex gap-4">
              <Icon className="mt-0.5 size-5 shrink-0 text-brand-600" strokeWidth={1.6} />
              <div>
                <p className="text-[15px] font-semibold">{title}</p>
                <p className="mt-1 text-sm leading-relaxed text-muted">{text}</p>
              </div>
            </li>
          ))}
        </ul>
      </Card>

      <p className="flex items-center gap-3 px-3 text-sm text-muted">
        <Lock className="size-4" strokeWidth={1.6} />
        Your location is only used to find shops nearby.
      </p>
    </aside>
  )
}

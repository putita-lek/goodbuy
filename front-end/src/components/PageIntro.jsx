export default function PageIntro({ eyebrow, title, subtitle, aside }) {
  return (
    <div className="mt-14 flex flex-wrap items-end justify-between gap-4">
      <div>
        <p className="text-sm font-medium tracking-[0.15em] text-brand-600 uppercase">{eyebrow}</p>
        <h1 className="mt-3 text-4xl font-semibold tracking-tight sm:text-[44px]">{title}</h1>
        <p className="mt-4 text-lg text-muted">{subtitle}</p>
      </div>
      {aside}
    </div>
  )
}

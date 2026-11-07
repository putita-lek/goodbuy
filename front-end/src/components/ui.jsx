// Small shared building blocks used across the page.

export function Card({ className = '', children }) {
  return (
    <section className={`rounded-2xl border border-line bg-white ${className}`}>
      {children}
    </section>
  )
}

export function SectionTitle({ number, title, aside }) {
  return (
    <div className="flex items-center justify-between gap-4">
      <div className="flex items-center gap-3">
        <span className="grid size-7 place-items-center rounded-full bg-brand-100 text-[13px] font-medium text-brand-700">
          {number}
        </span>
        <h2 className="text-2xl font-semibold tracking-tight">{title}</h2>
      </div>
      {aside}
    </div>
  )
}

export function Label({ htmlFor, children }) {
  return (
    <label htmlFor={htmlFor} className="mb-2 block text-[15px] font-medium">
      {children}
    </label>
  )
}

// Text input with optional leading icon/prefix and trailing suffix.
export function Field({ id, leading, suffix, className = '', ...props }) {
  return (
    <div
      className={`flex h-14 items-center gap-3 rounded-lg border border-line bg-white px-4 transition focus-within:border-brand-500 focus-within:ring-2 focus-within:ring-brand-100 ${className}`}
    >
      {leading && <span className="text-muted">{leading}</span>}
      <input
        id={id}
        className="w-full min-w-0 bg-transparent text-[17px] outline-none placeholder:text-muted/80"
        {...props}
      />
      {suffix && <span className="text-[15px] text-muted">{suffix}</span>}
    </div>
  )
}

// Rank badge: filled when selected, outlined when not.
export function RankBadge({ rank }) {
  return rank ? (
    <span className="grid size-[22px] place-items-center rounded bg-brand-600 text-xs font-semibold text-white">
      {rank}
    </span>
  ) : (
    <span className="size-[22px] rounded border border-line" />
  )
}

// Selectable card used for both priority option groups.
export function OptionCard({ icon: Icon, title, description, rank, onClick }) {
  const selected = Boolean(rank)
  return (
    <button
      type="button"
      onClick={onClick}
      aria-pressed={selected}
      className={`flex w-full flex-col items-start rounded-xl border p-5 text-left transition ${
        selected
          ? 'border-brand-500 bg-brand-50'
          : 'border-line bg-white hover:border-brand-500/50'
      }`}
    >
      <div className="flex w-full items-start justify-between">
        <Icon className="size-6 text-brand-600" strokeWidth={1.6} />
        <RankBadge rank={rank} />
      </div>
      <span className="mt-5 text-[17px] font-semibold">{title}</span>
      <span className="mt-1 text-sm text-muted">{description}</span>
    </button>
  )
}

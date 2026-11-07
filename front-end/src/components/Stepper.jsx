const STEPS = ['Your preferences', 'Review your list', 'Your shop recommendation']

export default function Stepper({ current }) {
  return (
    <ol className="flex items-center gap-4">
      {STEPS.map((label, i) => {
        const n = i + 1
        const active = n === current
        const done = n < current
        return (
          <li key={label} className="flex flex-1 items-center gap-4 last:flex-none">
            <span className="flex items-center gap-3 whitespace-nowrap">
              <span
                className={`grid size-9 place-items-center rounded-full text-sm font-medium ${
                  active
                    ? 'bg-brand-600 text-white'
                    : done
                      ? 'bg-brand-100 text-brand-700'
                      : 'border border-line text-muted'
                }`}
              >
                {n}
              </span>
              <span
                className={`hidden text-[17px] sm:inline ${
                  active ? 'font-semibold' : 'text-muted'
                }`}
              >
                {label}
              </span>
            </span>
            {n < STEPS.length && <span className="h-px flex-1 bg-line" />}
          </li>
        )
      })}
    </ol>
  )
}

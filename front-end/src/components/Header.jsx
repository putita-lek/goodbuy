import { CircleHelp, Clover } from 'lucide-react'

export default function Header() {
  return (
    <header className="border-b border-line bg-white">
      <div className="mx-auto flex h-[104px] max-w-[1640px] items-center gap-10 px-6 lg:px-[77px]">
        <a href="#" className="flex items-center gap-3" aria-label="GoodBuy home">
          <span className="grid size-11 place-items-center rounded-xl bg-brand-600 text-white">
            <Clover className="size-6" strokeWidth={2} />
          </span>
          <span className="flex flex-col leading-none">
            <span className="text-[28px] tracking-tight">
              Good<span className="font-semibold text-brand-600">Buy</span>
            </span>
            <span className="mt-1.5 hidden text-xs text-muted sm:block">Say goodbye to bad buys</span>
          </span>
        </a>

        <nav className="hidden gap-10 text-[17px] md:flex">
          <a href="#" className="font-semibold text-brand-600">Plan a shop</a>
          <a href="#" className="text-muted hover:text-ink">Saved lists</a>
        </nav>

        <div className="ml-auto flex items-center gap-8">
          <a href="#" className="hidden items-center gap-2 text-[16px] text-muted hover:text-ink sm:flex">
            <CircleHelp className="size-5" strokeWidth={1.6} />
            How it works
          </a>
          <span className="grid size-11 place-items-center rounded-full bg-sand text-sm font-semibold">
            JD
          </span>
        </div>
      </div>
    </header>
  )
}

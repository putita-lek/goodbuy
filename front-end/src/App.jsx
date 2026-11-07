import { useEffect, useMemo, useState } from 'react'
import { ArrowRight, Clover, Loader2 } from 'lucide-react'
import Header from './components/Header.jsx'
import Stepper from './components/Stepper.jsx'
import PageIntro from './components/PageIntro.jsx'
import PreferencesCard from './components/PreferencesCard.jsx'
import ShoppingList from './components/ShoppingList.jsx'
import SidePanel from './components/SidePanel.jsx'
import ReviewPage from './components/ReviewPage.jsx'
import SummaryPage from './components/SummaryPage.jsx'
import { optionsForItem } from './data/shops.js'
import { submitPreference } from './api/submitPreference.js'

const INITIAL_PREFS = {
  priorities: [], // order = rank
  values: [], // order = rank
  location: '', // free text the user types (or "Current location")
  latitude: null, // set by "Use my current location"; sent to the API as user_location
  longitude: null,
  maxDistance: '',
  budget: '',
}

const INITIAL_ITEMS = []

export default function App() {
  const [step, setStep] = useState(1)
  const [prefs, setPrefs] = useState(INITIAL_PREFS)
  const [items, setItems] = useState(INITIAL_ITEMS)
  const [selections, setSelections] = useState({}) // itemId -> shopId the user picked
  const [sessionId, setSessionId] = useState(null) // returned by POST /submit-preference
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [submitError, setSubmitError] = useState(null)

  const handleReviewMyList = async () => {
    if (isSubmitting) return
    setIsSubmitting(true)
    setSubmitError(null)

    try {
      const id = await submitPreference(prefs, items)
      setSessionId(id)
      setStep(2) // only move on once the API has accepted the preferences
    } catch (err) {
      setSubmitError(err.message)
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleSeeSummaryClicked = () => {
    // call api
    setStep(3)
  }

  // Options per item, ranked by the current preferences. A manual pick wins
  // if that shop is still an option; otherwise the best match is chosen.
  const plan = useMemo(
    () =>
      items.map((item) => {
        const options = optionsForItem(item, prefs)
        const chosen = options.find((o) => o.shop.id === selections[item.id]) ?? options[0] ?? null
        return { item, options, chosen }
      }),
    [items, prefs, selections],
  )
  const total = plan.reduce((sum, p) => sum + (p.chosen?.price ?? 0), 0)

  useEffect(() => {
    window.scrollTo({ top: 0 })
  }, [step])

  const canContinue = items.length > 0 && prefs.priorities.length > 0

  return (
    <div className="flex min-h-screen flex-col">
      <Header />

      <main className="mx-auto w-full max-w-[1508px] flex-1 px-6 pt-10 pb-10">
        <Stepper current={step} />

        {step === 1 && (
          <>
            <PageIntro
              eyebrow="Step 1 of 3 · Make it yours"
              title="Good groceries start with a good plan."
              subtitle="Tell us what matters to you. We’ll help you find a nearby shop that fits your list and budget."
              aside={<span className="text-[16px] text-muted">Weekly essentials</span>}
            />

            <div className="mt-11 grid gap-7 lg:grid-cols-[1fr_442px]">
              <div className="flex flex-col gap-7">
                <PreferencesCard prefs={prefs} onChange={setPrefs} />
                <ShoppingList items={items} onChange={setItems} />

                <div className="flex flex-wrap items-center justify-between gap-4 pt-2">
                  <p className="text-[15px] text-muted">Next: check your items and estimated prices</p>
                  <button
                    type="button"
                    disabled={!canContinue || isSubmitting}
                    onClick={handleReviewMyList}
                    className="flex items-center gap-3 rounded-lg bg-brand-600 px-6 py-4 text-[17px] font-semibold text-white shadow-sm transition hover:bg-brand-700 disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    {isSubmitting ? (
                      <>
                        Sending… <Loader2 className="size-5 animate-spin" />
                      </>
                    ) : (
                      <>
                        Review my list <ArrowRight className="size-5" />
                      </>
                    )}
                  </button>
                </div>

                {submitError && (
                  <p
                    role="alert"
                    className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-[15px] text-red-700"
                  >
                    {submitError}
                  </p>
                )}
              </div>
              <SidePanel />
            </div>
          </>
        )}

        {step === 2 && (
          <ReviewPage
            plan={plan}
            prefs={prefs}
            total={total}
            sessionId={sessionId}
            onSelect={(itemId, shopId) => setSelections((s) => ({ ...s, [itemId]: shopId }))}
            onBack={() => setStep(1)}
            onNext={() => setStep(3)}
          />
        )}

        {step === 3 && (
          <SummaryPage plan={plan} prefs={prefs} total={total} onBack={() => setStep(2)} />
        )}
      </main>

      <footer className="border-t border-line">
        <div className="mx-auto flex max-w-[1508px] flex-wrap items-center justify-between gap-4 px-6 py-8 text-[15px] text-muted">
          <span className="flex items-center gap-2">
            <Clover className="size-4 text-brand-600" strokeWidth={1.8} />
            GoodBuy · Say goodbye to bad buys.
          </span>
          <nav className="flex gap-8">
            <a href="#" className="hover:text-ink">Privacy</a>
            <a href="#" className="hover:text-ink">Help &amp; support</a>
            <span>United Kingdom · GBP £</span>
          </nav>
        </div>
      </footer>
    </div>
  )
}

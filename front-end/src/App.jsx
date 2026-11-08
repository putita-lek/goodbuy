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
import { optionsForItem, apiOptionsForItem } from './data/shops.js'
import { submitPreference, getOptionResult, submitOption } from './api/submitPreference.js'

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

function ErrorBanner({ message }) {
  if (!message) return null
  return (
    <p
      role="alert"
      className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-[15px] text-red-700"
    >
      {message}
    </p>
  )
}

export default function App() {
  const [step, setStep] = useState(1)
  const [prefs, setPrefs] = useState(INITIAL_PREFS)
  const [items, setItems] = useState(INITIAL_ITEMS)
  const [selections, setSelections] = useState({}) // itemId -> option key the user picked
  const [sessionId, setSessionId] = useState(null) // returned by POST /submit-preference
  const [userLocation, setUserLocation] = useState(null) // the point the API searched around
  const [optionResult, setOptionResult] = useState(null) // returned by GET /shopping-item/option-result
  // The list and preferences that produced optionResult. Pages 2 and 3 use these,
  // so editing step 1 never mixes new items with old results.
  const [submitted, setSubmitted] = useState({ prefs: INITIAL_PREFS, items: INITIAL_ITEMS })
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [submitError, setSubmitError] = useState(null)
  const [isSaving, setIsSaving] = useState(false)
  const [saveError, setSaveError] = useState(null)

  // Step 1 -> 2: submit the preferences, then fetch the options for that session.
  // Both calls must succeed before moving on; any error keeps the user on step 1
  // with a message, and they can simply click the button again.
  const handleReviewMyList = async () => {
    if (isSubmitting) return
    setIsSubmitting(true)
    setSubmitError(null)
    setOptionResult(null) // a new submission means any earlier results are stale
    setSelections({}) // ...and so are manual shop picks

    try {
      const { sessionId: id, userLocation: located } = await submitPreference(prefs, items)
      const result = await getOptionResult(id)
      setSessionId(id)
      setUserLocation(located)
      setOptionResult(result)
      setSubmitted({ prefs, items })
      setSaveError(null)
      setStep(2)
    } catch (err) {
      setSubmitError(err.message)
    } finally {
      setIsSubmitting(false)
    }
  }

  // Options per submitted item, in the API's ranked order. A manual pick wins;
  // otherwise the best match is chosen.
  const plan = useMemo(
    () =>
      submitted.items.map((item, index) => {
        const apiOptions = apiOptionsForItem(optionResult, index)
        const options = optionsForItem(item, submitted.prefs, apiOptions, userLocation)
        const chosen = options.find((o) => o.key === selections[item.id]) ?? options[0] ?? null
        return { item, resultId: index + 1, options, chosen }
      }),
    [submitted, selections, optionResult, userLocation],
  )
  const total = Math.round(plan.reduce((sum, p) => sum + (p.chosen?.price ?? 0), 0) * 100) / 100

  // Step 2 -> 3: save the chosen option for each item, then show the summary.
  const handleSeeSummaryClicked = async () => {
    if (isSaving) return
    setIsSaving(true)
    setSaveError(null)
    try {
      const choices = plan
        .filter((p) => p.chosen)
        .map((p) => ({ resultId: p.resultId, optionId: p.chosen.optionId }))
      await submitOption(sessionId, choices)
      setStep(3)
    } catch (err) {
      setSaveError(err.message)
    } finally {
      setIsSaving(false)
    }
  }

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
                        Finding your options… <Loader2 className="size-5 animate-spin" />
                      </>
                    ) : (
                      <>
                        Review my list <ArrowRight className="size-5" />
                      </>
                    )}
                  </button>
                </div>

                <ErrorBanner message={submitError} />
              </div>
              <SidePanel />
            </div>
          </>
        )}

        {step === 2 && (
          <ReviewPage
            plan={plan}
            prefs={submitted.prefs}
            total={total}
            onSelect={(itemId, optionKey) => setSelections((s) => ({ ...s, [itemId]: optionKey }))}
            onBack={() => setStep(1)}
            onNext={handleSeeSummaryClicked}
            isSaving={isSaving}
            saveError={saveError}
          />
        )}

        {step === 3 && (
          <SummaryPage
            plan={plan}
            prefs={submitted.prefs}
            userLocation={userLocation}
            total={total}
            onBack={() => setStep(2)}
          />
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

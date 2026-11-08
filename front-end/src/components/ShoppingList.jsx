import { useState } from 'react'
import { Plus, X } from 'lucide-react'
import { Card, Field, Label, SectionTitle } from './ui.jsx'

const MAX_ITEMS = 10
const MAX_DESCRIPTION_LENGTH = 30

export default function ShoppingList({ items, onChange }) {
  const [description, setDescription] = useState('')
  const [quantity, setQuantity] = useState('1')

  const qty = quantity.trim() === "" ? NaN : Number(quantity)
  const qtyValid = Number.isFinite(qty) && qty > 0
  const canAdd =
    items.length < MAX_ITEMS &&
    description.trim() !== '' &&
    description.length <= MAX_DESCRIPTION_LENGTH &&
    qtyValid

  const addItem = (e) => {
    e.preventDefault()
    if (!canAdd) return
    // id is sent to the API as item_id; quantity is stored as a number (item_quantity).
    onChange([...items, { id: crypto.randomUUID(), description: description.trim(), quantity: qty }])
    setDescription('')
    setQuantity('1')
  }

  const removeItem = (id) => onChange(items.filter((i) => i.id !== id))

  return (
    <Card className="p-7">
      <SectionTitle
        number={3}
        title="Build your shopping list"
        aside={
          <span className="text-[15px] text-muted">
            {items.length} {items.length === 1 ? 'item' : 'items'}
          </span>
        }
      />

      <table className="mt-5 w-full text-left">
        <thead>
          <tr className="border-b border-line text-[13px] font-medium tracking-wide text-muted uppercase">
            <th className="px-4 py-3 font-medium">Item description</th>
            <th className="w-[200px] px-1 py-3 font-medium">Quantity</th>
            <th className="w-12" aria-label="Remove" />
          </tr>
        </thead>
        <tbody>
          {items.map((item) => (
            <tr key={item.id} className="border-b border-line last:border-b-0">
              <td className="px-4 py-3.5 text-[17px]">{item.description}</td>
              <td className="px-1 py-3.5 text-[15px] text-muted">{item.quantity}</td>
              <td className="py-3.5 text-right">
                <button
                  type="button"
                  onClick={() => removeItem(item.id)}
                  aria-label={`Remove ${item.description}`}
                  className="rounded p-1 text-ink/70 hover:bg-page hover:text-ink"
                >
                  <X className="size-5" strokeWidth={1.6} />
                </button>
              </td>
            </tr>
          ))}
          {items.length === 0 && (
            <tr>
              <td colSpan={3} className="px-4 py-6 text-center text-[15px] text-muted">
                Your list is empty — add your first item below.
              </td>
            </tr>
          )}
        </tbody>
      </table>

      <form onSubmit={addItem} className="mt-6 grid items-end gap-4 sm:grid-cols-[1fr_140px_auto]">
        <div>
          <Label htmlFor="item-description">Item description</Label>
          <Field
            id="item-description"
            placeholder="e.g. Greek yoghurt"
            maxLength={MAX_DESCRIPTION_LENGTH}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
          />
          <p className="mt-1 text-right text-xs text-muted" 
          >
            {description.length}/{MAX_DESCRIPTION_LENGTH}
          </p>
        </div>
        <div>
          <Label htmlFor="item-quantity">Quantity</Label>
          <Field
            id="item-quantity"
            inputMode="decimal"
            value={quantity}
            onChange={(e) => setQuantity(e.target.value)}
            aria-invalid={!qtyValid}
          />
        </div>
        <button
          type="submit"
          disabled={!canAdd}
          className="flex h-14 items-center justify-center gap-3 rounded-lg border border-line bg-white px-6 text-[17px] font-semibold transition hover:border-brand-500 disabled:cursor-not-allowed disabled:opacity-60"
        >
          Add item <Plus className="size-5" />
        </button>
      </form>
      {items.length >= MAX_ITEMS && (
        <p className="mt-3 text-[15px] text-red-600">You can add up to {MAX_ITEMS} items.</p>
      )}
    </Card>
  )
}

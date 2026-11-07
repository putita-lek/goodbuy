import { useState } from 'react'
import { Plus, X } from 'lucide-react'
import { Card, Field, Label, SectionTitle } from './ui.jsx'

export default function ShoppingList({ items, onChange }) {
  const [description, setDescription] = useState('')
  const [quantity, setQuantity] = useState('1')

  const addItem = (e) => {
    e.preventDefault()
    const desc = description.trim()
    if (!desc) return
    onChange([...items, { id: crypto.randomUUID(), description: desc, quantity: quantity.trim() || '1' }])
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
            value={description}
            onChange={(e) => setDescription(e.target.value)}
          />
        </div>
        <div>
          <Label htmlFor="item-quantity">Quantity</Label>
          <Field id="item-quantity" value={quantity} onChange={(e) => setQuantity(e.target.value)} />
        </div>
        <button
          type="submit"
          disabled={!description.trim()}
          className="flex h-14 items-center justify-center gap-3 rounded-lg border border-line bg-white px-6 text-[17px] font-semibold transition hover:border-brand-500 disabled:cursor-not-allowed disabled:opacity-60"
        >
          Add item <Plus className="size-5" />
        </button>
      </form>
    </Card>
  )
}

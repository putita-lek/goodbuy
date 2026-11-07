import { Apple, Cherry, Egg, Leaf, Milk, ShoppingBasket, Wheat } from 'lucide-react'

const ICONS = [
  [/milk/i, Milk],
  [/egg/i, Egg],
  [/apple/i, Apple],
  [/bread|sourdough|loaf/i, Wheat],
  [/tomato/i, Cherry],
  [/spinach|salad|lettuce/i, Leaf],
]

export default function ItemIcon({ description, className = 'size-5' }) {
  const Icon = ICONS.find(([re]) => re.test(description))?.[1] ?? ShoppingBasket
  return <Icon className={className} strokeWidth={1.6} />
}

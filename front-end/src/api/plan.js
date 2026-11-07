const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? ''

const VALUE_TAG = {
  local: 'Local',
  organic: 'Organic',
  recycle: 'Recycle',
}

/**
 * Converts the first-page state into the API payload.
 * The order of priorities and values is meaningful: index 0 is rank 1.
 */
export function handleReviewMyList(prefs, items) {
  return {
    preferences: {
      priorities: prefs.priorities.map((id, index) => ({ id, rank: index + 1 })),
      values: prefs.values.map((id, index) => ({ id, rank: index + 1 })),
      location: prefs.location.trim(),
      maxDistanceMiles: Number(prefs.maxDistance),
      budgetGbp: Number(prefs.budget),
    },
    items: items.map(({ id, description, quantity }) => ({
      clientId: id,
      description: description.trim(),
      quantity: quantity.trim(),
    })),
  }
}

/**
 * Requests the first page of shop/item matches.
 *
 * The endpoint and response contract are documented in API_DRAFT.md.
 * Keeping fetch here makes it possible to swap the mock data source without
 * coupling the UI to transport or error-handling details.
 */
export async function requestPlan(prefs, items, { signal } = {}) {
  const response = await fetch(`${API_BASE_URL}/api/v1/shopping-plans`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(buildPlanRequest(prefs, items)),
    signal,
  })

  if (!response.ok) {
    let message = `Unable to create shopping plan (${response.status})`
    try {
      const error = await response.json()
      if (error.message) message = error.message
    } catch {
      // Keep the HTTP error when the server does not return JSON.
    }
    throw new Error(message)
  }

  return response.json()
}

/**
 * Adapts the API contract to the plan shape already used by the review and
 * summary pages. API values remain the source of truth for prices, shops,
 * distances, tags, and ranking.
 */
export function normalizePlanResponse(response, items) {
  const responseItems = response?.items ?? []

  return items.map((item) => {
    const result = responseItems.find(
      (candidate) => String(candidate.clientId) === String(item.id),
    )
    const options = (result?.matches ?? [])
      .filter((match) => match.available !== false)
      .map((match, index) => {
        const matchedTags = (match.matchedValues ?? []).map(
          (value) => VALUE_TAG[value] ?? value,
        )
        const rank = match.rank ?? index + 1
        return {
          shop: {
            id: match.shop.id,
            name: match.shop.name,
            type: match.shop.type,
            distance: match.shop.distanceMiles,
            tags: match.shop.tags ?? matchedTags,
          },
          price: match.priceGbp,
          matchedTags,
          rank: rank === 1 ? 'best' : rank === 2 ? 'runner' : 'other',
          reason: match.reason ?? null,
          letter: String.fromCharCode(65 + index),
        }
      })

    return {
      item: result?.requested
        ? { ...item, description: result.requested.description, quantity: result.requested.quantity }
        : item,
      options,
      chosen: options[0] ?? null,
    }
  })
}

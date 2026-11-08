// Cloudflare Worker for goodbuy.sohamjadhav.com.
//
// - Static files (the built React app in dist/) are served by Cloudflare directly.
// - Requests to /api/* are forwarded to the Go API (API_ORIGIN), without the /api
//   prefix: /api/submit-preference -> {API_ORIGIN}/submit-preference.
// - A cron trigger pings the API's /health every 10 minutes, so the free API host
//   never goes to sleep and visitors don't wait for it to wake up.

export default {
  async fetch(request, env) {
    const url = new URL(request.url)
    if (url.pathname === '/api' || url.pathname.startsWith('/api/')) {
      return proxyToApi(request, url, env)
    }
    return env.ASSETS.fetch(request)
  },

  async scheduled(_event, env, ctx) {
    ctx.waitUntil(
      fetch(`${env.API_ORIGIN}/health`, { headers: { 'User-Agent': 'goodbuy-keepalive' } }).catch(() => {}),
    )
  },
}

async function proxyToApi(request, url, env) {
  if (!env.API_ORIGIN) {
    return Response.json({ error: 'API_ORIGIN is not configured' }, { status: 500 })
  }
  const target = new URL(url.pathname.replace(/^\/api/, '') || '/', env.API_ORIGIN)
  target.search = url.search

  const headers = new Headers(request.headers)
  headers.delete('host')
  headers.delete('cookie') // the API doesn't use cookies; don't forward any

  try {
    return await fetch(target, {
      method: request.method,
      headers,
      body: ['GET', 'HEAD'].includes(request.method) ? undefined : request.body,
      redirect: 'manual',
    })
  } catch {
    return Response.json({ error: 'The GoodBuy API is not reachable right now. Please try again.' }, { status: 502 })
  }
}

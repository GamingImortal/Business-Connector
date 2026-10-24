// Talks to the Go backend (proxied by Vite).

const OWNER_KEY = 'bc-chat-owner'
const CUSTOMER_KEY = 'bc-chat-customer'

export function loadOwner() {
  try {
    return JSON.parse(localStorage.getItem(OWNER_KEY))
  } catch {
    return null
  }
}

export function saveOwner(owner) {
  if (owner) localStorage.setItem(OWNER_KEY, JSON.stringify(owner))
  else localStorage.removeItem(OWNER_KEY)
}

// Each browser gets its own customer identity so chats and bookings persist across visits.
export function loadCustomer() {
  try {
    const saved = JSON.parse(localStorage.getItem(CUSTOMER_KEY))
    if (saved?.id) return saved
  } catch {
    // fall through and create a new identity
  }
  const customer = { id: `c-${crypto.randomUUID()}`, name: '' }
  localStorage.setItem(CUSTOMER_KEY, JSON.stringify(customer))
  return customer
}

export function saveCustomer(customer) {
  localStorage.setItem(CUSTOMER_KEY, JSON.stringify(customer))
}

export async function api(path, { method = 'GET', body, token } = {}) {
  const headers = { 'Content-Type': 'application/json' }
  if (token) headers.Authorization = `Bearer ${token}`
  let res
  try {
    res = await fetch(`/api${path}`, { method, headers, body: body ? JSON.stringify(body) : undefined })
  } catch {
    throw new Error('Cannot reach the server. Is the backend running on port 8080?')
  }
  if (res.status === 204) return null
  const data = await res.json().catch(() => null)
  if (!res.ok) {
    const err = new Error(data?.error || `Request failed (${res.status})`)
    err.status = res.status
    throw err
  }
  return data
}

export function formatTime(rfc3339) {
  if (!rfc3339) return ''
  const d = new Date(rfc3339)
  const today = new Date()
  if (d.toDateString() === today.toDateString()) {
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false })
  }
  return d.toLocaleDateString([], { weekday: 'short' })
}

export function money(v) {
  return Number.isInteger(v) ? `$${v}` : `$${Number(v).toFixed(2)}`
}

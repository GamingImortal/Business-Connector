// Local-time date helpers. Dates are exchanged with the backend as "YYYY-MM-DD".
export function isoDate(d) {
  return new Date(d.getTime() - d.getTimezoneOffset() * 60000).toISOString().slice(0, 10)
}

export function parseISO(s) {
  const [y, m, d] = s.split('-').map(Number)
  return new Date(y, m - 1, d)
}

export function addDays(d, n) {
  const copy = new Date(d)
  copy.setDate(copy.getDate() + n)
  return copy
}

export const todayISO = () => isoDate(new Date())

export function startOfWeek(d) {
  const day = (d.getDay() + 6) % 7 // Monday = 0
  return addDays(new Date(d.getFullYear(), d.getMonth(), d.getDate()), -day)
}

export function dayLabel(iso) {
  const d = parseISO(iso)
  const today = todayISO()
  const label = d.toLocaleDateString([], { weekday: 'short', day: 'numeric', month: 'short' })
  if (iso === today) return `Today · ${label}`
  if (iso === isoDate(addDays(new Date(), 1))) return `Tomorrow · ${label}`
  return label
}

export function money(v) {
  return Number.isInteger(v) ? `$${v}` : `$${Number(v).toFixed(2)}`
}

export function timeAgo(rfc3339) {
  const d = new Date(rfc3339)
  if (isoDate(d) === todayISO()) return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false })
  return d.toLocaleDateString([], { day: 'numeric', month: 'short' })
}

export const STATUSES = {
  confirmed: { label: 'Confirmed', icon: '✅' },
  completed: { label: 'Completed', icon: '✔️' },
  cancelled: { label: 'Cancelled', icon: '❌' },
  no_show: { label: 'No-show', icon: '🚫' },
}

export const WEEKDAYS = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday']

// Supports WhatsApp-style *bold* text.
export function formatText(text) {
  return text.split(/(\*[^*\n]+\*)/g).map((part, i) =>
    part.startsWith('*') && part.endsWith('*') && part.length > 2 ? <strong key={i}>{part.slice(1, -1)}</strong> : part,
  )
}

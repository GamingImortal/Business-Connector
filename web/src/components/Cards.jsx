import { useState } from 'react'
import { api, money } from '../api.js'

// Renders the interactive cards the backend attaches to bot messages.
// `actions` holds the callbacks for the current viewer (customer or owner).
export default function Card({ card, businessId, actions }) {
  switch (card.type) {
    case 'offerings': return <OfferingsCard card={card} actions={actions} />
    case 'slots': return <SlotsCard card={card} businessId={businessId} actions={actions} />
    case 'booking': return <BookingCard reservation={card.reservation} actions={actions} />
    case 'bookings': return <BookingsCard card={card} actions={actions} />
    case 'businesses': return <BusinessesCard card={card} actions={actions} />
    default: return null
  }
}

// The Business Connect directory: tap Chat to start talking to a business in the same conversation.
function BusinessesCard({ card, actions }) {
  if (!card.items.length) return <small className="muted">Loading businesses…</small>
  return (
    <div className="card-box">
      {card.items.map((b) => (
        <div key={b.id} className="card-line biz-line">
          <span className="biz-avatar" style={{ background: b.color }}>{b.avatar}</span>
          <div className="grow">
            <span><strong>{b.name}</strong></span>
            <small>{b.category} · ⭐ {b.rating ? b.rating.toFixed(1) : 'New'}{b.location ? ` · ${b.location}` : ''}</small>
          </div>
          {actions.choose && <button className="mini-btn chat-btn" onClick={() => actions.choose(b)}>💬 Chat</button>}
        </div>
      ))}
    </div>
  )
}

function OfferingsCard({ card, actions }) {
  const [items, setItems] = useState(card.items)
  const groups = []
  for (const item of items) {
    const key = item.category || 'Other'
    let g = groups.find((x) => x.key === key)
    if (!g) groups.push((g = { key, items: [] }))
    g.items.push(item)
  }

  const toggle = async (item) => {
    const updated = await actions.toggleOffering(item)
    if (updated) setItems((list) => list.map((x) => (x.id === updated.id ? updated : x)))
  }

  return (
    <div className="card-box">
      {groups.map((g) => (
        <div key={g.key} className="card-group">
          <div className="card-group-title">{g.key}</div>
          {g.items.map((o) => (
            <div key={o.id} className={`card-line ${o.available ? '' : 'sold-out'}`}>
              <div className="grow">
                <span>{o.name}</span>
                {o.description && <small>{o.description}</small>}
              </div>
              <strong>{money(o.price)}</strong>
              {card.owner && (
                <button className="mini-btn" onClick={() => toggle(o)}>{o.available ? 'Sold out' : 'Available'}</button>
              )}
              {!card.owner && o.durationMins > 0 && <small className="nowrap">{o.durationMins} min</small>}
            </div>
          ))}
        </div>
      ))}
    </div>
  )
}

const localISO = (date) =>
  `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`

function shiftDate(iso, days) {
  const [y, m, d] = iso.split('-').map(Number)
  return localISO(new Date(y, m - 1, d + days))
}

function SlotsCard({ card, businessId, actions }) {
  const [view, setView] = useState(card)
  const [service, setService] = useState(card.services?.[0]?.id || '')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  const [booked, setBooked] = useState(null)
  const restaurant = card.businessType === 'restaurant'
  const today = localISO(new Date())

  const reload = async (date, partySize) => {
    setError('')
    try {
      const d = await api(`/businesses/${businessId}/availability?date=${date}&party=${partySize}`)
      setView((v) => ({ ...v, date: d.date, dateLabel: d.dateLabel, slots: d.slots, partySize }))
    } catch (e) {
      setError(e.message)
    }
  }

  const book = async (time) => {
    if (!actions.book) return
    setBusy(true)
    setError('')
    try {
      await actions.book({ date: view.date, time, partySize: view.partySize, offeringId: service ? Number(service) : null })
      setBooked(time)
    } catch (e) {
      setError(e.message)
      reload(view.date, view.partySize)
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="card-box">
      <div className="slot-nav">
        <button className="mini-btn" disabled={view.date <= today} onClick={() => reload(shiftDate(view.date, -1), view.partySize)}>‹</button>
        <strong>{view.dateLabel}</strong>
        <button className="mini-btn" onClick={() => reload(shiftDate(view.date, 1), view.partySize)}>›</button>
      </div>
      {restaurant && (
        <div className="slot-option">
          Guests
          <button className="mini-btn" disabled={view.partySize <= 1} onClick={() => reload(view.date, view.partySize - 1)}>−</button>
          <strong>{view.partySize}</strong>
          <button className="mini-btn" onClick={() => reload(view.date, view.partySize + 1)}>+</button>
        </div>
      )}
      {!restaurant && card.services?.length > 0 && (
        <select className="slot-option" value={service} onChange={(e) => setService(e.target.value)}>
          {card.services.map((s) => (
            <option key={s.id} value={s.id}>{s.name} – {money(s.price)}</option>
          ))}
        </select>
      )}
      {view.slots.length === 0 ? (
        <small className="muted">Closed on this day — try the next one ›</small>
      ) : (
        <div className="slots">
          {view.slots.map((s) => (
            <button key={s.time} className={`slot ${booked === s.time ? 'booked' : ''}`}
              disabled={!s.available || busy || !!booked || !actions.book}
              onClick={() => book(s.time)} title={s.available ? 'Tap to book' : 'Fully booked'}>
              {s.time}
            </button>
          ))}
        </div>
      )}
      {error && <div className="card-error">{error}</div>}
    </div>
  )
}

function reservationSummary(r) {
  const parts = [r.date, r.time]
  if (r.offeringName) parts.push(r.offeringName)
  if (r.partySize > 1) parts.push(`${r.partySize} guests`)
  return parts.join(' · ')
}

const STATUS_LABEL = { confirmed: '✅ Confirmed', completed: '✔️ Completed', cancelled: '❌ Cancelled', no_show: '🚫 No-show' }

function BookingCard({ reservation, actions }) {
  const [r, setR] = useState(reservation)
  const [busy, setBusy] = useState(false)
  const cancel = async () => {
    if (!confirm('Cancel this booking?')) return
    setBusy(true)
    try {
      setR(await actions.cancelBooking(r))
    } catch (e) {
      alert(e.message)
    } finally {
      setBusy(false)
    }
  }
  return (
    <div className="card-box booking">
      <div><strong>Booking #{r.id}</strong> · {STATUS_LABEL[r.status]}</div>
      <div>{reservationSummary(r)}</div>
      {actions.cancelBooking && r.status === 'confirmed' && (
        <button className="mini-btn danger" disabled={busy} onClick={cancel}>Cancel booking</button>
      )}
    </div>
  )
}

function BookingsCard({ card, actions }) {
  const [list, setList] = useState(card.reservations)
  const replace = (u) => u && setList((l) => l.map((r) => (r.id === u.id ? u : r)))

  const run = async (fn) => {
    try {
      replace(await fn())
    } catch (e) {
      alert(e.message)
    }
  }

  return (
    <div className="card-box">
      {list.map((r) => (
        <div key={r.id} className={`card-line ${r.status === 'confirmed' ? '' : 'sold-out'}`}>
          <div className="grow">
            <span>#{r.id} · {r.time} · {r.customerName}</span>
            <small>{reservationSummary(r)}{r.notes ? ` · ${r.notes}` : ''}</small>
          </div>
          {r.status !== 'confirmed' && <small className="nowrap">{STATUS_LABEL[r.status]}</small>}
          {r.status === 'confirmed' && card.owner && actions.setStatus && (
            <span className="nowrap">
              <button className="mini-btn" title="Completed" onClick={() => run(() => actions.setStatus(r, 'completed'))}>✔</button>
              <button className="mini-btn" title="No-show" onClick={() => run(() => actions.setStatus(r, 'no_show'))}>🚫</button>
              <button className="mini-btn danger" title="Cancel" onClick={() => run(() => actions.setStatus(r, 'cancelled'))}>✕</button>
            </span>
          )}
          {r.status === 'confirmed' && !card.owner && actions.cancelBooking && (
            <button className="mini-btn danger" onClick={() => confirm('Cancel this booking?') && run(() => actions.cancelBooking(r))}>Cancel</button>
          )}
        </div>
      ))}
    </div>
  )
}

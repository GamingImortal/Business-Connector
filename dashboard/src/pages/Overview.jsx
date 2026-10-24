import { useEffect, useState } from 'react'
import { api } from '../api.js'
import { addDays, dayLabel, isoDate, todayISO } from '../format.jsx'
import { ReservationActions, StatusPill } from '../components/reservations.jsx'

export default function Overview({ business, goTo }) {
  const [reservations, setReservations] = useState([])
  const [offerings, setOfferings] = useState([])
  const [conversations, setConversations] = useState([])
  const [error, setError] = useState('')

  const today = todayISO()
  const weekEnd = isoDate(addDays(new Date(), 6))

  useEffect(() => {
    Promise.all([
      api(`/owner/reservations?from=${today}&to=${weekEnd}`),
      api('/owner/offerings'),
      api('/owner/conversations'),
    ])
      .then(([r, o, c]) => {
        setReservations(r)
        setOfferings(o)
        setConversations(c)
      })
      .catch((e) => setError(e.message))
  }, [today, weekEnd])

  const active = (r) => r.status === 'confirmed' || r.status === 'completed'
  const todays = reservations.filter((r) => r.date === today)
  const upcoming = reservations.filter((r) => r.date > today && r.status === 'confirmed').slice(0, 8)
  const restaurant = business.type === 'restaurant'
  const guestsToday = todays.filter(active).reduce((n, r) => n + r.partySize, 0)
  const unread = conversations.reduce((n, c) => n + c.unread, 0)

  const replace = (updated) => setReservations((list) => list.map((r) => (r.id === updated.id ? updated : r)))

  return (
    <div className="page">
      <header className="page-header">
        <div>
          <h1>Good {greeting()}, {business.name} 👋</h1>
          <p className="muted">Here's what's happening at your business.</p>
        </div>
      </header>
      {error && <div className="error">{error}</div>}

      <div className="stats">
        <Stat label={restaurant ? 'Reservations today' : 'Appointments today'} value={todays.filter(active).length}
          sub={restaurant ? `${guestsToday} guests` : null} onClick={() => goTo('calendar')} />
        <Stat label="Next 7 days" value={reservations.filter(active).length} onClick={() => goTo('reservations')} />
        <Stat label="Unread messages" value={unread} onClick={() => goTo('messages')} />
        <Stat label={restaurant ? 'Menu items' : 'Listed offerings'} value={offerings.length}
          sub={`${offerings.filter((o) => !o.available).length} unavailable`} onClick={() => goTo('offerings')} />
        <Stat label="Rating" value={business.rating ? `⭐ ${business.rating.toFixed(1)}` : '—'} />
      </div>

      <div className="grid-2">
        <section className="card">
          <h2>{dayLabel(today)}</h2>
          {todays.length === 0 && <p className="muted">No bookings today.</p>}
          <ul className="booking-list">
            {todays.map((r) => (
              <li key={r.id}>
                <span className="time">{r.time}</span>
                <div className="grow">
                  <strong>{r.customerName}</strong>
                  <small>
                    {restaurant ? `${r.partySize} guests` : r.offeringName || 'Appointment'}
                    {r.notes && ` · ${r.notes}`}
                  </small>
                </div>
                <StatusPill status={r.status} />
                <ReservationActions reservation={r} onChange={replace} />
              </li>
            ))}
          </ul>
        </section>

        <section className="card">
          <h2>Coming up</h2>
          {upcoming.length === 0 && <p className="muted">Nothing booked for the rest of the week yet.</p>}
          <ul className="booking-list">
            {upcoming.map((r) => (
              <li key={r.id}>
                <span className="time">{r.time}</span>
                <div className="grow">
                  <strong>{r.customerName}</strong>
                  <small>{dayLabel(r.date)} · {restaurant ? `${r.partySize} guests` : r.offeringName || 'Appointment'}</small>
                </div>
                {r.source === 'chat' && <span className="tag">via chat</span>}
              </li>
            ))}
          </ul>
        </section>
      </div>

      {offerings.length === 0 && (
        <div className="callout">
          👉 Customers can't see any prices yet. <button className="link-btn" onClick={() => goTo('offerings')}>Add your first {restaurant ? 'menu item' : 'offering'}</button>
        </div>
      )}
    </div>
  )
}

function Stat({ label, value, sub, onClick }) {
  return (
    <button className="stat" onClick={onClick} disabled={!onClick}>
      <span className="stat-value">{value}</span>
      <span className="stat-label">{label}</span>
      {sub && <small>{sub}</small>}
    </button>
  )
}

function greeting() {
  const h = new Date().getHours()
  return h < 12 ? 'morning' : h < 17 ? 'afternoon' : 'evening'
}

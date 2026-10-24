import { useEffect, useState } from 'react'
import { api } from '../api.js'
import { addDays, isoDate, startOfWeek, todayISO } from '../format.jsx'
import { ReservationForm } from '../components/reservations.jsx'

export default function Calendar({ business, goTo }) {
  const [weekStart, setWeekStart] = useState(() => startOfWeek(new Date()))
  const [reservations, setReservations] = useState([])
  const [form, setForm] = useState(null) // { reservation } or { defaults }
  const [showCancelled, setShowCancelled] = useState(false)
  const [error, setError] = useState('')

  const days = Array.from({ length: 7 }, (_, i) => addDays(weekStart, i))
  const from = isoDate(days[0])
  const to = isoDate(days[6])

  const load = () => api(`/owner/reservations?from=${from}&to=${to}`).then(setReservations).catch((e) => setError(e.message))
  useEffect(() => {
    load()
  }, [from, to]) // eslint-disable-line react-hooks/exhaustive-deps

  const restaurant = business.type === 'restaurant'
  const today = todayISO()

  return (
    <div className="page">
      <header className="page-header">
        <div>
          <h1>Calendar</h1>
          <p className="muted">
            {business.hoursText} · {business.slotMinutes}-minute slots · capacity {business.capacity} {restaurant ? 'seats' : 'bookings'} per slot ·{' '}
            <button className="link-btn" onClick={() => goTo('settings')}>change</button>
          </p>
        </div>
        <div className="toolbar">
          <button className="secondary-btn" onClick={() => setWeekStart(addDays(weekStart, -7))}>‹</button>
          <button className="secondary-btn" onClick={() => setWeekStart(startOfWeek(new Date()))}>This week</button>
          <button className="secondary-btn" onClick={() => setWeekStart(addDays(weekStart, 7))}>›</button>
          <label className="checkbox small">
            <input type="checkbox" checked={showCancelled} onChange={(e) => setShowCancelled(e.target.checked)} /> Show cancelled
          </label>
        </div>
      </header>
      {error && <div className="error">{error}</div>}

      <div className="week">
        {days.map((d) => {
          const iso = isoDate(d)
          const hours = business.hours[d.getDay()]
          const dayRes = reservations.filter((r) => r.date === iso && (showCancelled || (r.status !== 'cancelled' && r.status !== 'no_show')))
          const booked = dayRes.filter((r) => r.status === 'confirmed' || r.status === 'completed')
          const guests = booked.reduce((n, r) => n + r.partySize, 0)
          return (
            <div key={iso} className={`day ${iso === today ? 'today' : ''} ${hours.closed ? 'closed' : ''} ${iso < today ? 'past' : ''}`}>
              <div className="day-head">
                <strong>{d.toLocaleDateString([], { weekday: 'short' })}</strong>
                <span>{d.getDate()} {d.toLocaleDateString([], { month: 'short' })}</span>
                <small>{hours.closed ? 'Closed' : `${hours.open}–${hours.close}`}</small>
                {booked.length > 0 && <small className="day-load">{restaurant ? `${guests} guests` : `${booked.length} booked`}</small>}
              </div>
              <div className="day-body">
                {dayRes.map((r) => (
                  <button key={r.id} className={`event event-${r.status}`} onClick={() => setForm({ reservation: r })}>
                    <span className="event-time">{r.time}</span>
                    <span className="event-name">{r.customerName}</span>
                    <small>{restaurant ? `👥 ${r.partySize}` : r.offeringName}</small>
                  </button>
                ))}
              </div>
              <button className="add-event" onClick={() => setForm({ defaults: { date: iso } })}>+ Add</button>
            </div>
          )
        })}
      </div>

      {form && (
        <ReservationForm
          business={business}
          reservation={form.reservation}
          defaults={form.defaults}
          onClose={() => setForm(null)}
          onSaved={() => {
            setForm(null)
            load()
          }}
        />
      )}
    </div>
  )
}

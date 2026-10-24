import { useEffect, useState } from 'react'
import { api } from '../api.js'
import { STATUSES, addDays, dayLabel, isoDate, todayISO } from '../format.jsx'
import { ReservationActions, ReservationForm, StatusPill } from '../components/reservations.jsx'

export default function Reservations({ business }) {
  const [from, setFrom] = useState(todayISO())
  const [to, setTo] = useState(isoDate(addDays(new Date(), 30)))
  const [status, setStatus] = useState('')
  const [query, setQuery] = useState('')
  const [list, setList] = useState([])
  const [form, setForm] = useState(null)
  const [error, setError] = useState('')

  const load = () => api(`/owner/reservations?from=${from}&to=${to}`).then(setList).catch((e) => setError(e.message))
  useEffect(() => {
    load()
  }, [from, to]) // eslint-disable-line react-hooks/exhaustive-deps

  const restaurant = business.type === 'restaurant'
  const q = query.trim().toLowerCase()
  const shown = list.filter((r) =>
    (!status || r.status === status) &&
    (!q || `${r.customerName} ${r.customerPhone} ${r.notes} ${r.offeringName}`.toLowerCase().includes(q)),
  )
  const replace = (updated) => setList((l) => l.map((r) => (r.id === updated.id ? updated : r)))

  return (
    <div className="page">
      <header className="page-header">
        <div>
          <h1>Reservations</h1>
          <p className="muted">Bookings made in the chat appear here automatically. Add walk-ins and phone bookings yourself.</p>
        </div>
        <button className="primary-btn" onClick={() => setForm({})}>+ New reservation</button>
      </header>
      {error && <div className="error">{error}</div>}

      <div className="filters">
        <label>From<input type="date" value={from} onChange={(e) => setFrom(e.target.value)} /></label>
        <label>To<input type="date" value={to} onChange={(e) => setTo(e.target.value)} /></label>
        <label>
          Status
          <select value={status} onChange={(e) => setStatus(e.target.value)}>
            <option value="">All</option>
            {Object.entries(STATUSES).map(([k, s]) => <option key={k} value={k}>{s.label}</option>)}
          </select>
        </label>
        <label className="grow">Search<input value={query} onChange={(e) => setQuery(e.target.value)} placeholder="Name, phone, notes…" /></label>
      </div>

      <section className="card">
        {shown.length === 0 ? <p className="muted">No reservations match.</p> : (
          <table className="table">
            <thead>
              <tr>
                <th>#</th><th>When</th><th>Customer</th>
                <th>{restaurant ? 'Guests' : 'Service'}</th>
                <th>Notes</th><th>Status</th><th />
              </tr>
            </thead>
            <tbody>
              {shown.map((r) => (
                <tr key={r.id} className="clickable" onClick={() => setForm({ reservation: r })}>
                  <td className="muted">{r.id}</td>
                  <td><strong>{r.time}</strong><small className="block muted">{dayLabel(r.date)}</small></td>
                  <td>
                    {r.customerName}
                    {r.customerPhone && <small className="block muted">{r.customerPhone}</small>}
                    {r.source === 'chat' && <span className="tag">via chat</span>}
                  </td>
                  <td>{restaurant ? r.partySize : r.offeringName || '—'}</td>
                  <td className="notes">{r.notes}</td>
                  <td><StatusPill status={r.status} /></td>
                  <td onClick={(e) => e.stopPropagation()}><ReservationActions reservation={r} onChange={replace} /></td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>

      {form && (
        <ReservationForm
          business={business}
          reservation={form.reservation}
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

import { useEffect, useState } from 'react'
import { api } from '../api.js'
import { STATUSES, todayISO } from '../format.jsx'

export function StatusPill({ status }) {
  const s = STATUSES[status] || { label: status, icon: '' }
  return <span className={`pill pill-${status}`}>{s.icon} {s.label}</span>
}

// Quick status buttons for a reservation. Calls onChange with the updated reservation.
export function ReservationActions({ reservation, onChange }) {
  const [busy, setBusy] = useState(false)
  const setStatus = async (status) => {
    setBusy(true)
    try {
      onChange(await api(`/owner/reservations/${reservation.id}`, { method: 'PUT', body: { status } }))
    } catch (e) {
      alert(e.message)
    } finally {
      setBusy(false)
    }
  }
  if (reservation.status !== 'confirmed') {
    return (
      <div className="actions">
        <button className="small-btn" disabled={busy} onClick={() => setStatus('confirmed')} title="Restore">↺ Restore</button>
      </div>
    )
  }
  return (
    <div className="actions">
      <button className="small-btn ok" disabled={busy} onClick={() => setStatus('completed')} title="Mark completed">✔ Done</button>
      <button className="small-btn" disabled={busy} onClick={() => setStatus('no_show')} title="Mark no-show">🚫</button>
      <button className="small-btn danger" disabled={busy} onClick={() => setStatus('cancelled')} title="Cancel">✕</button>
    </div>
  )
}

// Modal for adding or editing a reservation.
export function ReservationForm({ business, reservation, defaults, onSaved, onClose }) {
  const isNew = !reservation
  const [offerings, setOfferings] = useState([])
  const [form, setForm] = useState(() => ({
    customerName: '', customerPhone: '', date: todayISO(), time: '', partySize: business.type === 'restaurant' ? 2 : 1,
    offeringId: null, notes: '', status: 'confirmed', ...defaults, ...reservation,
  }))
  const [slots, setSlots] = useState([])
  const [error, setError] = useState('')
  const set = (k, v) => setForm((f) => ({ ...f, [k]: v }))

  useEffect(() => {
    api('/owner/offerings').then((list) => setOfferings(list.filter((o) => o.kind === 'service'))).catch(() => {})
  }, [])

  useEffect(() => {
    if (!form.date) return
    api(`/owner/availability?date=${form.date}&party=${form.partySize}`).then((d) => setSlots(d.slots)).catch(() => setSlots([]))
  }, [form.date, form.partySize])

  const save = async (e) => {
    e.preventDefault()
    setError('')
    try {
      const body = { ...form, partySize: Number(form.partySize) || 1, offeringId: form.offeringId ? Number(form.offeringId) : null }
      const saved = isNew
        ? await api('/owner/reservations', { method: 'POST', body })
        : await api(`/owner/reservations/${reservation.id}`, { method: 'PUT', body })
      onSaved(saved)
    } catch (err) {
      setError(err.message)
    }
  }

  const remove = async () => {
    if (!confirm('Delete this reservation permanently?')) return
    await api(`/owner/reservations/${reservation.id}`, { method: 'DELETE' })
    onSaved(null)
  }

  return (
    <div className="modal-backdrop" onClick={onClose}>
      <form className="modal" onClick={(e) => e.stopPropagation()} onSubmit={save}>
        <header>
          <h3>{isNew ? 'New reservation' : `Reservation #${reservation.id}`}</h3>
          <button type="button" className="icon-btn" onClick={onClose}>✕</button>
        </header>
        {error && <div className="error">{error}</div>}
        <div className="row">
          <label>Customer name<input value={form.customerName} onChange={(e) => set('customerName', e.target.value)} required autoFocus /></label>
          <label>Phone<input value={form.customerPhone} onChange={(e) => set('customerPhone', e.target.value)} /></label>
        </div>
        <div className="row">
          <label>Date<input type="date" value={form.date} onChange={(e) => set('date', e.target.value)} required /></label>
          <label>Time<input type="time" value={form.time} onChange={(e) => set('time', e.target.value)} required /></label>
          {business.type === 'restaurant' && (
            <label>Guests<input type="number" min={1} max={business.capacity} value={form.partySize} onChange={(e) => set('partySize', e.target.value)} /></label>
          )}
        </div>
        {slots.length > 0 && (
          <div className="slot-picker">
            {slots.map((s) => (
              <button type="button" key={s.time} className={`slot ${form.time === s.time ? 'selected' : ''} ${s.available ? '' : 'full'}`}
                onClick={() => set('time', s.time)} title={`${s.remaining} of ${business.capacity} free`}>
                {s.time}<small>{s.remaining} free</small>
              </button>
            ))}
          </div>
        )}
        {slots.length === 0 && form.date && <p className="hint">Closed on this day — you can still add a booking manually.</p>}
        {offerings.length > 0 && (
          <label>
            Service
            <select value={form.offeringId || ''} onChange={(e) => set('offeringId', e.target.value || null)}>
              <option value="">—</option>
              {offerings.map((o) => <option key={o.id} value={o.id}>{o.name}</option>)}
            </select>
          </label>
        )}
        {!isNew && (
          <label>
            Status
            <select value={form.status} onChange={(e) => set('status', e.target.value)}>
              {Object.entries(STATUSES).map(([k, s]) => <option key={k} value={k}>{s.icon} {s.label}</option>)}
            </select>
          </label>
        )}
        <label>Notes<textarea rows={2} value={form.notes} onChange={(e) => set('notes', e.target.value)} /></label>
        <footer>
          {!isNew && <button type="button" className="link-btn danger" onClick={remove}>Delete</button>}
          <span className="spacer" />
          <button type="button" className="secondary-btn" onClick={onClose}>Cancel</button>
          <button className="primary-btn">Save</button>
        </footer>
      </form>
    </div>
  )
}

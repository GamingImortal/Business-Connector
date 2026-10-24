import { useState } from 'react'
import { api } from '../api.js'
import { WEEKDAYS } from '../format.jsx'

// Monday first, like most calendars.
const DAY_ORDER = [1, 2, 3, 4, 5, 6, 0]

export default function Settings({ business, setBusiness }) {
  const [form, setForm] = useState(business)
  const [status, setStatus] = useState('')
  const [error, setError] = useState('')
  const restaurant = form.type === 'restaurant'

  const set = (k, v) => {
    setForm((f) => ({ ...f, [k]: v }))
    setStatus('')
  }
  const setDay = (weekday, changes) =>
    set('hours', form.hours.map((h) => (h.weekday === weekday ? { ...h, ...changes } : h)))
  const setAuto = (changes) => set('autoReply', { ...form.autoReply, ...changes })
  const setRule = (id, changes) =>
    setAuto({ rules: form.autoReply.rules.map((r) => (r.id === id ? { ...r, ...changes } : r)) })

  const save = async (e) => {
    e.preventDefault()
    setError('')
    try {
      const body = {
        ...form,
        slotMinutes: Number(form.slotMinutes),
        capacity: Number(form.capacity),
        autoReply: { ...form.autoReply, rules: form.autoReply.rules.filter((r) => r.keywords.trim() && r.reply.trim()) },
      }
      const updated = await api('/owner/business', { method: 'PUT', body })
      setBusiness(updated)
      setForm(updated)
      setStatus('✓ Saved')
    } catch (err) {
      setError(err.message)
    }
  }

  return (
    <div className="page">
      <header className="page-header">
        <div>
          <h1>Settings</h1>
          <p className="muted">Your profile, opening hours and booking rules are used by the chat assistant to answer customers.</p>
        </div>
      </header>

      <form onSubmit={save} className="settings">
        <section className="card form">
          <h2>Business profile</h2>
          <div className="row">
            <label>Name<input value={form.name} onChange={(e) => set('name', e.target.value)} required /></label>
            <label>Category<input value={form.category} onChange={(e) => set('category', e.target.value)} /></label>
          </div>
          <div className="row">
            <label>
              Type
              <select value={form.type} onChange={(e) => set('type', e.target.value)}>
                <option value="service">Service business (appointments)</option>
                <option value="restaurant">Restaurant (table reservations)</option>
              </select>
            </label>
            <label>Icon (emoji)<input value={form.avatar} onChange={(e) => set('avatar', e.target.value)} maxLength={4} /></label>
            <label>Colour<input type="color" value={form.color} onChange={(e) => set('color', e.target.value)} /></label>
          </div>
          <div className="row">
            <label>Location<input value={form.location} onChange={(e) => set('location', e.target.value)} /></label>
            <label>Phone<input value={form.phone} onChange={(e) => set('phone', e.target.value)} /></label>
          </div>
          <label>About<textarea rows={2} value={form.about} onChange={(e) => set('about', e.target.value)} /></label>
        </section>

        <section className="card form">
          <h2>Opening hours & bookings</h2>
          <table className="hours">
            <tbody>
              {DAY_ORDER.map((d) => {
                const h = form.hours[d]
                return (
                  <tr key={d} className={h.closed ? 'dimmed' : ''}>
                    <td>{WEEKDAYS[d]}</td>
                    <td>
                      <label className="checkbox small">
                        <input type="checkbox" checked={!h.closed} onChange={(e) => setDay(d, { closed: !e.target.checked })} /> Open
                      </label>
                    </td>
                    <td><input type="time" value={h.open} disabled={h.closed} onChange={(e) => setDay(d, { open: e.target.value })} /></td>
                    <td>to</td>
                    <td><input type="time" value={h.close} disabled={h.closed} onChange={(e) => setDay(d, { close: e.target.value })} /></td>
                  </tr>
                )
              })}
            </tbody>
          </table>
          <div className="row">
            <label>
              {restaurant ? 'Table sitting length (minutes)' : 'Appointment slot length (minutes)'}
              <input type="number" min={15} max={480} step={15} value={form.slotMinutes} onChange={(e) => set('slotMinutes', e.target.value)} />
            </label>
            <label>
              {restaurant ? 'Seats (guests per sitting)' : 'Bookings at the same time'}
              <input type="number" min={1} max={1000} value={form.capacity} onChange={(e) => set('capacity', e.target.value)} />
            </label>
          </div>
          <p className="hint">
            {restaurant
              ? 'A reservation is accepted when there are enough free seats for the whole party.'
              : 'For example, a barbershop with 2 chairs can take 2 bookings in the same slot.'}
          </p>
        </section>

        <section className="card form">
          <h2>🤖 Automated chat replies</h2>
          <label className="checkbox">
            <input type="checkbox" checked={form.autoReply.enabled} onChange={(e) => setAuto({ enabled: e.target.checked })} />
            Reply to customers automatically (prices, menu, hours, bookings and your keyword replies)
          </label>
          <fieldset disabled={!form.autoReply.enabled}>
            <label>👋 Greeting <small>sent on a customer's first message</small>
              <textarea rows={2} value={form.autoReply.greeting} onChange={(e) => setAuto({ greeting: e.target.value })} />
            </label>
            <div className="field-label">🔑 Keyword replies <small>checked before the automatic answers</small></div>
            {form.autoReply.rules.map((r) => (
              <div className="rule" key={r.id}>
                <div className="rule-top">
                  <input placeholder="Keywords, comma separated" value={r.keywords} onChange={(e) => setRule(r.id, { keywords: e.target.value })} />
                  <button type="button" className="small-btn danger"
                    onClick={() => setAuto({ rules: form.autoReply.rules.filter((x) => x.id !== r.id) })}>🗑</button>
                </div>
                <textarea rows={2} placeholder="Reply" value={r.reply} onChange={(e) => setRule(r.id, { reply: e.target.value })} />
              </div>
            ))}
            <button type="button" className="link-btn"
              onClick={() => setAuto({ rules: [...form.autoReply.rules, { id: `r${Date.now()}`, keywords: '', reply: '' }] })}>
              + Add keyword reply
            </button>
            <label>💬 Fallback <small>sent when nothing else matches</small>
              <textarea rows={2} value={form.autoReply.fallback} onChange={(e) => setAuto({ fallback: e.target.value })} />
            </label>
          </fieldset>
        </section>

        <div className="save-bar">
          {error && <span className="error inline">{error}</span>}
          <span className="saved-note">{status}</span>
          <button className="primary-btn">Save changes</button>
        </div>
      </form>

      <ChangePIN />
    </div>
  )
}

function ChangePIN() {
  const [currentPin, setCurrent] = useState('')
  const [newPin, setNew] = useState('')
  const [message, setMessage] = useState('')
  const digits = (fn) => (e) => fn(e.target.value.replace(/\D/g, ''))

  const submit = async (e) => {
    e.preventDefault()
    try {
      await api('/owner/pin', { method: 'PUT', body: { currentPin, newPin } })
      setMessage('✓ PIN changed')
      setCurrent('')
      setNew('')
    } catch (err) {
      setMessage(err.message)
    }
  }

  return (
    <form className="card form" onSubmit={submit}>
      <h2>🔒 Change PIN</h2>
      <div className="row">
        <label>Current PIN<input type="password" inputMode="numeric" maxLength={6} value={currentPin} onChange={digits(setCurrent)} /></label>
        <label>New PIN (4–6 digits)<input type="password" inputMode="numeric" maxLength={6} value={newPin} onChange={digits(setNew)} /></label>
        <button className="secondary-btn align-end" disabled={currentPin.length < 4 || newPin.length < 4}>Change PIN</button>
      </div>
      {message && <p className="hint">{message}</p>}
    </form>
  )
}

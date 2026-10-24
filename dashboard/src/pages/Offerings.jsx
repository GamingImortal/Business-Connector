import { useEffect, useState } from 'react'
import { api } from '../api.js'
import { money } from '../format.jsx'

const KINDS = {
  menu: { label: 'Menu item', plural: '🍽️ Menu' },
  service: { label: 'Service', plural: '✂️ Services' },
  product: { label: 'Product', plural: '🛍️ Products' },
}

export default function Offerings({ business }) {
  const [items, setItems] = useState([])
  const [editing, setEditing] = useState(null) // offering being edited, or {} for new
  const [error, setError] = useState('')
  const defaultKind = business.type === 'restaurant' ? 'menu' : 'service'

  useEffect(() => {
    api('/owner/offerings').then(setItems).catch((e) => setError(e.message))
  }, [])

  const save = async (o) => {
    const body = { ...o, price: Number(o.price) || 0, durationMins: Number(o.durationMins) || 0 }
    const saved = o.id
      ? await api(`/owner/offerings/${o.id}`, { method: 'PUT', body })
      : await api('/owner/offerings', { method: 'POST', body })
    setItems((list) => (o.id ? list.map((x) => (x.id === saved.id ? saved : x)) : [...list, saved]))
    setEditing(null)
  }

  const toggle = async (o) => {
    try {
      await save({ ...o, available: !o.available })
    } catch (e) {
      setError(e.message)
    }
  }

  const remove = async (o) => {
    if (!confirm(`Delete "${o.name}"?`)) return
    try {
      await api(`/owner/offerings/${o.id}`, { method: 'DELETE' })
      setItems((list) => list.filter((x) => x.id !== o.id))
    } catch (e) {
      setError(e.message)
    }
  }

  const groups = Object.keys(KINDS)
    .map((kind) => ({ kind, items: items.filter((o) => o.kind === kind) }))
    .filter((g) => g.items.length)

  return (
    <div className="page">
      <header className="page-header">
        <div>
          <h1>{business.type === 'restaurant' ? 'Menu & products' : 'Services & products'}</h1>
          <p className="muted">Everything listed here is shown to customers in the chat, with prices. Toggle items off when they're sold out.</p>
        </div>
        <button className="primary-btn" onClick={() => setEditing({ kind: defaultKind, available: true })}>+ Add item</button>
      </header>
      {error && <div className="error">{error}</div>}

      {groups.length === 0 && <div className="card empty">Nothing listed yet. Click <strong>+ Add item</strong> to get started.</div>}

      {groups.map((g) => (
        <section className="card" key={g.kind}>
          <h2>{KINDS[g.kind].plural}</h2>
          <table className="table">
            <thead>
              <tr>
                <th>Name</th><th>Category</th><th>Price</th>
                {g.kind === 'service' && <th>Duration</th>}
                <th>Available</th><th />
              </tr>
            </thead>
            <tbody>
              {g.items.map((o) => (
                <tr key={o.id} className={o.available ? '' : 'dimmed'}>
                  <td><strong>{o.name}</strong>{o.description && <small className="block muted">{o.description}</small>}</td>
                  <td>{o.category || '—'}</td>
                  <td>{money(o.price)}</td>
                  {g.kind === 'service' && <td>{o.durationMins ? `${o.durationMins} min` : '—'}</td>}
                  <td>
                    <input type="checkbox" className="switch" checked={o.available} onChange={() => toggle(o)}
                      title={o.available ? 'Shown to customers' : 'Hidden (sold out)'} />
                  </td>
                  <td className="row-actions">
                    <button className="small-btn" onClick={() => setEditing(o)}>Edit</button>
                    <button className="small-btn danger" onClick={() => remove(o)}>🗑</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </section>
      ))}

      {editing && <OfferingForm offering={editing} onSave={save} onClose={() => setEditing(null)} />}
    </div>
  )
}

function OfferingForm({ offering, onSave, onClose }) {
  const [form, setForm] = useState({ name: '', category: '', description: '', price: '', durationMins: '', ...offering })
  const [error, setError] = useState('')
  const set = (k) => (e) => setForm((f) => ({ ...f, [k]: e.target.value }))

  const submit = async (e) => {
    e.preventDefault()
    try {
      await onSave(form)
    } catch (err) {
      setError(err.message)
    }
  }

  return (
    <div className="modal-backdrop" onClick={onClose}>
      <form className="modal" onClick={(e) => e.stopPropagation()} onSubmit={submit}>
        <header>
          <h3>{offering.id ? 'Edit item' : 'New item'}</h3>
          <button type="button" className="icon-btn" onClick={onClose}>✕</button>
        </header>
        {error && <div className="error">{error}</div>}
        <div className="row">
          <label>
            Type
            <select value={form.kind} onChange={set('kind')}>
              {Object.entries(KINDS).map(([k, v]) => <option key={k} value={k}>{v.label}</option>)}
            </select>
          </label>
          <label>Category<input value={form.category} onChange={set('category')} placeholder={form.kind === 'menu' ? 'e.g. Mains' : 'e.g. Cuts'} /></label>
        </div>
        <label>Name<input value={form.name} onChange={set('name')} required autoFocus /></label>
        <label>Description<input value={form.description} onChange={set('description')} /></label>
        <div className="row">
          <label>Price ($)<input type="number" min={0} step="0.01" value={form.price} onChange={set('price')} required /></label>
          {form.kind === 'service' && (
            <label>Duration (minutes)<input type="number" min={0} step={5} value={form.durationMins} onChange={set('durationMins')} /></label>
          )}
        </div>
        <label className="checkbox">
          <input type="checkbox" checked={form.available} onChange={(e) => setForm((f) => ({ ...f, available: e.target.checked }))} />
          Available to customers
        </label>
        <footer>
          <span className="spacer" />
          <button type="button" className="secondary-btn" onClick={onClose}>Cancel</button>
          <button className="primary-btn">Save</button>
        </footer>
      </form>
    </div>
  )
}

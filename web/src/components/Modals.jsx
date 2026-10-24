import { useState } from 'react'
import { api } from '../api.js'

// Business owners sign in with their business and PIN to use the chat as that business.
export function PinModal({ businesses, onLogin, onClose }) {
  const [businessId, setBusinessId] = useState(businesses[0]?.id || '')
  const [pin, setPin] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  const submit = async (e) => {
    e.preventDefault()
    setBusy(true)
    setError('')
    try {
      onLogin(await api('/auth/login', { method: 'POST', body: { businessId, pin } }))
    } catch (err) {
      setError(err.message)
      setPin('')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="modal-backdrop" onClick={onClose}>
      <form className="modal" onClick={(e) => e.stopPropagation()} onSubmit={submit}>
        <h3>🔒 Business owner login</h3>
        <p className="muted">Enter your business PIN to see your reservations, chats and assistant.</p>
        <label>
          Business
          <select value={businessId} onChange={(e) => setBusinessId(e.target.value)}>
            {businesses.map((b) => <option key={b.id} value={b.id}>{b.avatar} {b.name}</option>)}
          </select>
        </label>
        <label>
          PIN
          <input className="pin-input" type="password" inputMode="numeric" maxLength={6} autoFocus value={pin}
            onChange={(e) => setPin(e.target.value.replace(/\D/g, ''))} placeholder="••••" />
        </label>
        {error && <div className="form-error">{error}</div>}
        <div className="modal-actions">
          <small className="muted">Demo PIN: 1234</small>
          <button type="button" className="secondary-btn" onClick={onClose}>Cancel</button>
          <button className="primary-btn" disabled={busy || pin.length < 4}>{busy ? 'Checking…' : 'Unlock'}</button>
        </div>
      </form>
    </div>
  )
}

// Asked once, so businesses know who they are talking to and bookings carry a name.
export function NameModal({ onSave, onClose }) {
  const [name, setName] = useState('')
  return (
    <div className="modal-backdrop" onClick={onClose}>
      <form className="modal" onClick={(e) => e.stopPropagation()}
        onSubmit={(e) => { e.preventDefault(); if (name.trim()) onSave(name.trim()) }}>
        <h3>👋 What's your name?</h3>
        <p className="muted">Businesses see this name on your chats and bookings.</p>
        <input autoFocus value={name} onChange={(e) => setName(e.target.value)} placeholder="e.g. Thandi" maxLength={40} />
        <div className="modal-actions">
          <button type="button" className="secondary-btn" onClick={onClose}>Cancel</button>
          <button className="primary-btn" disabled={!name.trim()}>Continue</button>
        </div>
      </form>
    </div>
  )
}

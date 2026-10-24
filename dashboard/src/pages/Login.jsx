import { useEffect, useState } from 'react'
import { api } from '../api.js'

const AVATARS = ['🏪', '💈', '💅', '🍲', '🍕', '☕', '🔧', '⚡', '🧹', '📚', '📸', '🏋️', '🚗', '🐾']

export default function Login({ onLogin }) {
  const [tab, setTab] = useState('login')
  const [businesses, setBusinesses] = useState([])
  const [error, setError] = useState('')

  useEffect(() => {
    api('/businesses').then(setBusinesses).catch((e) => setError(e.message))
  }, [])

  return (
    <div className="login-screen">
      <div className="login-card">
        <h1>🏪 Business Connect</h1>
        <p className="muted">Dashboard for business owners</p>
        <div className="tabs">
          <button className={tab === 'login' ? 'active' : ''} onClick={() => setTab('login')}>Log in</button>
          <button className={tab === 'register' ? 'active' : ''} onClick={() => setTab('register')}>List your business</button>
        </div>
        {error && <div className="error">{error}</div>}
        {tab === 'login'
          ? <LoginForm businesses={businesses} onLogin={onLogin} setError={setError} />
          : <RegisterForm onLogin={onLogin} setError={setError} />}
      </div>
    </div>
  )
}

function LoginForm({ businesses, onLogin, setError }) {
  const [businessId, setBusinessId] = useState('')
  const [pin, setPin] = useState('')
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    if (!businessId && businesses.length) setBusinessId(businesses[0].id)
  }, [businesses, businessId])

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
    <form onSubmit={submit} className="form">
      <label>
        Business
        <select value={businessId} onChange={(e) => setBusinessId(e.target.value)}>
          {businesses.map((b) => (
            <option key={b.id} value={b.id}>{b.avatar} {b.name}</option>
          ))}
        </select>
      </label>
      <label>
        PIN
        <input
          className="pin-input"
          type="password"
          inputMode="numeric"
          autoComplete="current-password"
          maxLength={6}
          value={pin}
          onChange={(e) => setPin(e.target.value.replace(/\D/g, ''))}
          placeholder="••••"
          autoFocus
        />
      </label>
      <button className="primary-btn" disabled={busy || pin.length < 4}>{busy ? 'Checking…' : 'Log in'}</button>
      <p className="hint">Demo businesses use PIN <strong>1234</strong>.</p>
    </form>
  )
}

function RegisterForm({ onLogin, setError }) {
  const [form, setForm] = useState({
    name: '', type: 'service', category: '', avatar: '🏪', location: '', phone: '', about: '', pin: '', pin2: '',
  })
  const [busy, setBusy] = useState(false)
  const set = (k) => (e) => setForm((f) => ({ ...f, [k]: e.target.value }))

  const submit = async (e) => {
    e.preventDefault()
    if (form.pin !== form.pin2) return setError('PINs do not match')
    setBusy(true)
    setError('')
    try {
      const { pin2, ...body } = form
      if (body.type === 'restaurant') {
        body.slotMinutes = 90
        if (!body.category) body.category = 'Restaurant'
      }
      onLogin(await api('/businesses', { method: 'POST', body }))
    } catch (err) {
      setError(err.message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <form onSubmit={submit} className="form">
      <label>
        Business name
        <input value={form.name} onChange={set('name')} required placeholder="e.g. Chipo's Salon" />
      </label>
      <div className="radio-row">
        <label className={form.type === 'service' ? 'selected' : ''}>
          <input type="radio" name="type" value="service" checked={form.type === 'service'} onChange={set('type')} />
          ✂️ Service business<small>Appointments: salons, plumbers, tutors…</small>
        </label>
        <label className={form.type === 'restaurant' ? 'selected' : ''}>
          <input type="radio" name="type" value="restaurant" checked={form.type === 'restaurant'} onChange={set('type')} />
          🍽️ Restaurant<small>Menu and table reservations</small>
        </label>
      </div>
      <div className="row">
        <label>
          Category
          <input value={form.category} onChange={set('category')} placeholder={form.type === 'restaurant' ? 'Restaurant' : 'e.g. Barber'} />
        </label>
        <label>
          Icon
          <select value={form.avatar} onChange={set('avatar')}>
            {AVATARS.map((a) => <option key={a}>{a}</option>)}
          </select>
        </label>
      </div>
      <div className="row">
        <label>Location<input value={form.location} onChange={set('location')} /></label>
        <label>Phone<input value={form.phone} onChange={set('phone')} /></label>
      </div>
      <div className="row">
        <label>
          PIN (4–6 digits)
          <input type="password" inputMode="numeric" maxLength={6} value={form.pin}
            onChange={(e) => setForm((f) => ({ ...f, pin: e.target.value.replace(/\D/g, '') }))} required />
        </label>
        <label>
          Confirm PIN
          <input type="password" inputMode="numeric" maxLength={6} value={form.pin2}
            onChange={(e) => setForm((f) => ({ ...f, pin2: e.target.value.replace(/\D/g, '') }))} required />
        </label>
      </div>
      <button className="primary-btn" disabled={busy}>{busy ? 'Creating…' : 'Create business'}</button>
      <p className="hint">You can add your offerings, prices and opening hours next.</p>
    </form>
  )
}

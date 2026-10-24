import { useCallback, useEffect, useState } from 'react'
import { api, getToken, setToken } from './api.js'
import Login from './pages/Login.jsx'
import Overview from './pages/Overview.jsx'
import Offerings from './pages/Offerings.jsx'
import Calendar from './pages/Calendar.jsx'
import Reservations from './pages/Reservations.jsx'
import Messages from './pages/Messages.jsx'
import Settings from './pages/Settings.jsx'

const PAGES = [
  { id: 'overview', label: 'Overview', icon: '🏠' },
  { id: 'offerings', label: 'Offerings', icon: '📋' },
  { id: 'calendar', label: 'Calendar', icon: '🗓️' },
  { id: 'reservations', label: 'Reservations', icon: '📅' },
  { id: 'messages', label: 'Messages', icon: '💬' },
  { id: 'settings', label: 'Settings', icon: '⚙️' },
]

export default function App() {
  const [business, setBusiness] = useState(null)
  const [loading, setLoading] = useState(!!getToken())
  const [page, setPage] = useState('overview')
  const [unread, setUnread] = useState(0)

  useEffect(() => {
    const onLogout = () => setBusiness(null)
    window.addEventListener('bc-logout', onLogout)
    if (getToken()) {
      api('/owner/business')
        .then(setBusiness)
        .catch(() => setToken(null))
        .finally(() => setLoading(false))
    }
    return () => window.removeEventListener('bc-logout', onLogout)
  }, [])

  const refreshUnread = useCallback(() => {
    api('/owner/conversations')
      .then((list) => setUnread(list.reduce((n, c) => n + c.unread, 0)))
      .catch(() => {})
  }, [])

  useEffect(() => {
    if (!business) return
    refreshUnread()
    const t = setInterval(refreshUnread, 5000)
    return () => clearInterval(t)
  }, [business, refreshUnread])

  const onLogin = ({ token, business }) => {
    setToken(token)
    setBusiness(business)
    setPage('overview')
  }

  const logout = async () => {
    await api('/auth/logout', { method: 'POST' }).catch(() => {})
    setToken(null)
    setBusiness(null)
  }

  if (loading) return <div className="center-screen">Loading…</div>
  if (!business) return <Login onLogin={onLogin} />

  const props = { business, setBusiness, goTo: setPage }

  return (
    <div className="layout">
      <aside className="nav">
        <div className="nav-brand">
          <span className="avatar" style={{ background: business.color }}>{business.avatar}</span>
          <div>
            <strong>{business.name}</strong>
            <small>{business.type === 'restaurant' ? 'Restaurant' : business.category || 'Service business'}</small>
          </div>
        </div>
        <nav>
          {PAGES.map((p) => (
            <button key={p.id} className={page === p.id ? 'active' : ''} onClick={() => setPage(p.id)}>
              <span>{p.icon}</span> {p.label}
              {p.id === 'messages' && unread > 0 && <span className="badge">{unread}</span>}
            </button>
          ))}
        </nav>
        <div className="nav-footer">
          <a href="http://localhost:5173" target="_blank" rel="noreferrer">Open chat platform ↗</a>
          <button className="link-btn" onClick={logout}>Log out</button>
        </div>
      </aside>
      <main className="content">
        {page === 'overview' && <Overview {...props} />}
        {page === 'offerings' && <Offerings {...props} />}
        {page === 'calendar' && <Calendar {...props} />}
        {page === 'reservations' && <Reservations {...props} />}
        {page === 'messages' && <Messages {...props} onRead={refreshUnread} />}
        {page === 'settings' && <Settings {...props} />}
      </main>
    </div>
  )
}

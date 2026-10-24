import { useState } from 'react'
import Avatar from './Avatar.jsx'

const FILTERS = ['All', 'Unread', 'Auto-reply on']

export default function Sidebar({ businesses, messages, activeId, unread, typing, role, onRoleChange, onSelect, onReset }) {
  const [query, setQuery] = useState('')
  const [filter, setFilter] = useState('All')

  const q = query.trim().toLowerCase()
  const list = businesses.filter((b) => {
    if (q && !`${b.name} ${b.category}`.toLowerCase().includes(q)) return false
    if (filter === 'Unread' && !unread[b.id]) return false
    if (filter === 'Auto-reply on' && !b.autoReply.enabled) return false
    return true
  })

  return (
    <aside className={`sidebar ${activeId ? 'has-active' : ''}`}>
      <header className="sidebar-header">
        <div className="me">
          <Avatar emoji={role === 'customer' ? '🙂' : '🏪'} color="#6b7c85" />
          <span className="app-title">Business Connect</span>
        </div>
        <button className="icon-btn" title="Reset demo data" onClick={onReset}>↺</button>
      </header>

      <div className="role-switch" role="tablist" aria-label="View as">
        <span>View as:</span>
        {['customer', 'business'].map((r) => (
          <button
            key={r}
            className={role === r ? 'active' : ''}
            onClick={() => onRoleChange(r)}
          >
            {r === 'customer' ? '🙂 Customer' : '🏪 Business'}
          </button>
        ))}
      </div>

      <div className="search">
        <span className="search-icon">🔍</span>
        <input
          placeholder="Search businesses"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
        />
      </div>

      <div className="filters">
        {FILTERS.map((f) => (
          <button key={f} className={filter === f ? 'chip active' : 'chip'} onClick={() => setFilter(f)}>
            {f}
          </button>
        ))}
      </div>

      <ul className="chat-list">
        {list.map((b) => {
          const history = messages[b.id] || []
          const last = history[history.length - 1]
          const count = unread[b.id] || 0
          return (
            <li
              key={b.id}
              className={`chat-item ${b.id === activeId ? 'active' : ''}`}
              onClick={() => onSelect(b.id)}
            >
              <Avatar emoji={b.avatar} color={b.color} />
              <div className="chat-item-body">
                <div className="chat-item-top">
                  <span className="chat-name">{b.name}</span>
                  <span className={`chat-time ${count ? 'unread' : ''}`}>{last?.time || ''}</span>
                </div>
                <div className="chat-item-bottom">
                  <span className="chat-preview">
                    {typing[b.id] ? (
                      <em className="typing-text">typing…</em>
                    ) : last ? (
                      <>
                        {last.from === 'client' && <Ticks status={last.status} />}
                        {last.text.split('\n')[0]}
                      </>
                    ) : (
                      <span className="muted">{b.category} · ⭐ {b.rating}</span>
                    )}
                  </span>
                  {count > 0 && <span className="badge">{count}</span>}
                </div>
              </div>
            </li>
          )
        })}
        {list.length === 0 && <li className="no-results">No businesses found</li>}
      </ul>
    </aside>
  )
}

export function Ticks({ status }) {
  if (!status) return null
  return <span className={`ticks ${status === 'read' ? 'read' : ''}`}>{status === 'sent' ? '✓' : '✓✓'}</span>
}

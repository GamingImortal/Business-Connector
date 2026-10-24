import { useState } from 'react'
import Avatar from './Avatar.jsx'

// Chat list. Items are { id, name, avatar, color, preview, previewMine, previewStatus, time, unread, typing, search }.
export default function Sidebar({ items, activeId, role, owner, onRoleChange, onSelect, onLogout, filters, emptyText }) {
  const [query, setQuery] = useState('')
  const [filter, setFilter] = useState('All')

  const q = query.trim().toLowerCase()
  const active = filters.find((f) => f.label === filter) || filters[0]
  const list = items.filter((it) => (!q || it.search.toLowerCase().includes(q)) && (!active || active.test(it)))

  return (
    <aside className={`sidebar ${activeId ? 'has-active' : ''}`}>
      <header className="sidebar-header">
        <div className="me">
          <Avatar emoji={role === 'customer' ? '🙂' : owner?.business.avatar || '🏪'} color={role === 'customer' ? '#6b7c85' : owner?.business.color} />
          <span className="app-title">{role === 'business' && owner ? owner.business.name : 'Business Connect'}</span>
        </div>
        {role === 'business' && owner && (
          <>
            <a className="icon-btn" href="http://localhost:5174" target="_blank" rel="noreferrer" title="Open the business dashboard">📊</a>
            <button className="icon-btn" title="Log out" onClick={onLogout}>🚪</button>
          </>
        )}
      </header>

      <div className="role-switch" role="tablist" aria-label="View as">
        <span>View as:</span>
        {['customer', 'business'].map((r) => (
          <button key={r} className={role === r ? 'active' : ''} onClick={() => onRoleChange(r)}>
            {r === 'customer' ? '🙂 Customer' : '🏪 Business owner'}
          </button>
        ))}
      </div>

      {filters.length > 0 && <>
      <div className="search">
        <span className="search-icon">🔍</span>
        <input placeholder={role === 'customer' ? 'Search businesses' : 'Search chats'} value={query} onChange={(e) => setQuery(e.target.value)} />
      </div>

      <div className="filters">
        {filters.map((f) => (
          <button key={f.label} className={filter === f.label ? 'chip active' : 'chip'} onClick={() => setFilter(f.label)}>
            {f.label}
          </button>
        ))}
      </div>
      </>}

      <ul className="chat-list">
        {list.map((it) => (
          <li key={it.id} className={`chat-item ${it.id === activeId ? 'active' : ''}`} onClick={() => onSelect(it.id)}>
            <Avatar emoji={it.avatar} color={it.color} />
            <div className="chat-item-body">
              <div className="chat-item-top">
                <span className="chat-name">{it.name}</span>
                <span className={`chat-time ${it.unread ? 'unread' : ''}`}>{it.time}</span>
              </div>
              <div className="chat-item-bottom">
                <span className="chat-preview">
                  {it.typing ? (
                    <em className="typing-text">typing…</em>
                  ) : (
                    <>
                      {it.previewMine && <Ticks status={it.previewStatus} />}
                      {it.preview}
                    </>
                  )}
                </span>
                {it.unread > 0 && <span className="badge">{it.unread}</span>}
              </div>
            </div>
          </li>
        ))}
        {list.length === 0 && <li className="no-results">{emptyText}</li>}
      </ul>
    </aside>
  )
}

export function Ticks({ status }) {
  if (!status) return null
  return <span className={`ticks ${status === 'read' ? 'read' : ''}`}>{status === 'sent' ? '✓' : '✓✓'}</span>
}

import { useEffect, useRef, useState } from 'react'
import { api } from '../api.js'
import { formatText, timeAgo } from '../format.jsx'

// The business's customer chats, so owners can reply manually from the dashboard.
export default function Messages({ onRead }) {
  const [conversations, setConversations] = useState([])
  const [active, setActive] = useState(null)
  const [messages, setMessages] = useState([])
  const [text, setText] = useState('')
  const [error, setError] = useState('')
  const bottomRef = useRef(null)

  useEffect(() => {
    const load = () => api('/owner/conversations').then(setConversations).catch((e) => setError(e.message))
    load()
    const t = setInterval(load, 4000)
    return () => clearInterval(t)
  }, [])

  useEffect(() => {
    if (!active) return
    const load = () =>
      api(`/owner/conversations/${encodeURIComponent(active)}`)
        .then((m) => {
          setMessages((prev) => (prev.length === m.length ? prev : m))
          onRead()
        })
        .catch((e) => setError(e.message))
    setMessages([])
    load()
    const t = setInterval(load, 3000)
    return () => clearInterval(t)
  }, [active, onRead])

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages.length])

  const send = async (e) => {
    e.preventDefault()
    const value = text.trim()
    if (!value) return
    setText('')
    try {
      const msg = await api(`/owner/conversations/${encodeURIComponent(active)}`, { method: 'POST', body: { text: value } })
      setMessages((m) => [...m, msg])
    } catch (err) {
      setError(err.message)
    }
  }

  const current = conversations.find((c) => c.customerId === active)

  return (
    <div className="page messages-page">
      <header className="page-header">
        <div>
          <h1>Messages</h1>
          <p className="muted">Automated replies answer customers instantly; jump in here whenever you want to reply yourself.</p>
        </div>
      </header>
      {error && <div className="error">{error}</div>}
      <div className="inbox card">
        <ul className="inbox-list">
          {conversations.length === 0 && <li className="muted pad">No customer chats yet.</li>}
          {conversations.map((c) => (
            <li key={c.customerId} className={c.customerId === active ? 'active' : ''} onClick={() => setActive(c.customerId)}>
              <span className="avatar small">🙂</span>
              <div className="grow">
                <div className="spread">
                  <strong>{c.customerName || 'Customer'}</strong>
                  <small className="muted">{timeAgo(c.lastMessage.createdAt)}</small>
                </div>
                <div className="spread">
                  <small className="preview">{c.lastMessage.from === 'business' && (c.lastMessage.auto ? '🤖 ' : 'You: ')}{c.lastMessage.text.split('\n')[0]}</small>
                  {c.unread > 0 && <span className="badge">{c.unread}</span>}
                </div>
              </div>
            </li>
          ))}
        </ul>
        <div className="thread">
          {!active ? <div className="thread-empty muted">Select a chat to read and reply.</div> : (
            <>
              <div className="thread-head"><strong>{current?.customerName || 'Customer'}</strong></div>
              <div className="thread-body">
                {messages.map((m) => (
                  <div key={m.id} className={`bubble-row ${m.from === 'business' ? 'out' : 'in'}`}>
                    <div className="bubble">
                      {m.auto && <div className="auto-tag">🤖 Automated</div>}
                      {formatText(m.text)}
                      {m.card && <div className="card-note">[{cardLabel(m.card)}]</div>}
                      <small className="bubble-time">{timeAgo(m.createdAt)}</small>
                    </div>
                  </div>
                ))}
                <div ref={bottomRef} />
              </div>
              <form className="composer" onSubmit={send}>
                <input value={text} onChange={(e) => setText(e.target.value)} placeholder="Type a reply…" />
                <button className="primary-btn" disabled={!text.trim()}>Send</button>
              </form>
            </>
          )}
        </div>
      </div>
    </div>
  )
}

function cardLabel(card) {
  switch (card.type) {
    case 'offerings': return `${card.items.length} items shown`
    case 'slots': return `available times for ${card.dateLabel}`
    case 'booking': return `booking #${card.reservation.id} · ${card.reservation.date} ${card.reservation.time}`
    case 'bookings': return `${card.reservations.length} bookings listed`
    default: return card.type
  }
}

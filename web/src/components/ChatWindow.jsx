import { useEffect, useRef, useState } from 'react'
import Avatar from './Avatar.jsx'
import { Ticks } from './Sidebar.jsx'

export default function ChatWindow({ business, messages, typing, role, onSend, onBack, onOpenInfo, onOpenAutoReply }) {
  const [text, setText] = useState('')
  const bottomRef = useRef(null)
  const inputRef = useRef(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages.length, typing])

  useEffect(() => {
    inputRef.current?.focus()
  }, [role])

  const submit = (e) => {
    e.preventDefault()
    const value = text.trim()
    if (!value) return
    onSend(value)
    setText('')
  }

  const quickReplies = business.autoReply.enabled
    ? business.autoReply.rules.map((r) => r.keywords.split(',')[0].trim()).filter(Boolean)
    : []

  // From the customer's view, their messages are "mine" (right side). From the business view, the business's are.
  const mine = role === 'customer' ? 'client' : 'business'

  return (
    <section className="chat">
      <header className="chat-header">
        <button className="icon-btn back-btn" onClick={onBack} title="Back">←</button>
        <div className="chat-header-info" onClick={onOpenInfo}>
          <Avatar emoji={business.avatar} color={business.color} size={40} />
          <div>
            <div className="chat-name">{business.name}</div>
            <div className="chat-status">
              {typing ? <span className="typing-text">typing…</span> : `${business.category} · ⭐ ${business.rating} · click for info`}
            </div>
          </div>
        </div>
        <button
          className={`autoreply-btn ${business.autoReply.enabled ? 'on' : ''}`}
          onClick={onOpenAutoReply}
          title="Set up automated responses"
        >
          🤖 Auto-reply {business.autoReply.enabled ? 'ON' : 'OFF'}
        </button>
      </header>

      {role === 'business' && (
        <div className="role-banner">
          You are replying as <strong>{business.name}</strong>. Switch to "Customer" to chat as a client.
        </div>
      )}

      <div className="messages">
        <div className="encryption-note">
          🔒 This is a demo chat. Automated replies are set up by the business.
        </div>
        {messages.map((m) => (
          <div key={m.id} className={`bubble-row ${m.from === mine ? 'out' : 'in'}`}>
            <div className="bubble">
              {m.auto && <div className="auto-tag">🤖 Automated reply</div>}
              <span className="bubble-text">{formatText(m.text)}</span>
              <span className="bubble-meta">
                {m.time}
                {m.from === 'client' && m.from === mine && <Ticks status={m.status} />}
              </span>
            </div>
          </div>
        ))}
        {typing && (
          <div className="bubble-row in">
            <div className="bubble typing-bubble"><span /><span /><span /></div>
          </div>
        )}
        <div ref={bottomRef} />
      </div>

      {role === 'customer' && quickReplies.length > 0 && (
        <div className="quick-replies">
          {quickReplies.map((q) => (
            <button key={q} className="chip" onClick={() => onSend(q)}>{q}</button>
          ))}
        </div>
      )}

      <form className="composer" onSubmit={submit}>
        <button type="button" className="icon-btn" title="Emoji" onClick={() => setText((t) => t + '😊')}>😊</button>
        <input
          ref={inputRef}
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder={role === 'customer' ? 'Type a message' : `Reply as ${business.name}`}
        />
        <button type="submit" className="send-btn" title="Send" disabled={!text.trim()}>➤</button>
      </form>
    </section>
  )
}

// Supports WhatsApp-style *bold* text.
function formatText(text) {
  return text.split(/(\*[^*\n]+\*)/g).map((part, i) =>
    part.startsWith('*') && part.endsWith('*') && part.length > 2
      ? <strong key={i}>{part.slice(1, -1)}</strong>
      : part,
  )
}

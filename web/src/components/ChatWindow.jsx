import { useEffect, useRef, useState } from 'react'
import Avatar from './Avatar.jsx'
import { Ticks } from './Sidebar.jsx'
import { formatTime } from '../api.js'

const SpeechRecognition = typeof window !== 'undefined' && (window.SpeechRecognition || window.webkitSpeechRecognition)

// A WhatsApp-style conversation. Messages are { id, mine, text, createdAt, auto, status, card, sender }.
// `sender` ({ name, avatar, color }) labels messages in a thread shared by several businesses.
export default function ChatWindow({
  title, avatar, color, subtitle, onOpenInfo, headerExtra, banner, note, context,
  messages, typing, quickReplies = [], placeholder, onSend, onBack, renderCard,
}) {
  const [text, setText] = useState('')
  const [listening, setListening] = useState(false)
  const bottomRef = useRef(null)
  const inputRef = useRef(null)
  const recognitionRef = useRef(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages.length, typing])

  useEffect(() => {
    inputRef.current?.focus()
    return () => recognitionRef.current?.abort()
  }, [])

  const submit = (e) => {
    e.preventDefault()
    const value = text.trim()
    if (!value) return
    onSend(value)
    setText('')
  }

  // Voice input: speak a message and it is sent once you stop talking.
  const toggleVoice = () => {
    if (listening) {
      recognitionRef.current?.stop()
      return
    }
    const rec = new SpeechRecognition()
    rec.lang = navigator.language || 'en-US'
    rec.interimResults = true
    rec.onresult = (e) => {
      const transcript = Array.from(e.results).map((r) => r[0].transcript).join('')
      setText(transcript)
      if (e.results[e.results.length - 1].isFinal) {
        onSend(transcript.trim())
        setText('')
      }
    }
    rec.onend = () => setListening(false)
    rec.onerror = () => setListening(false)
    recognitionRef.current = rec
    rec.start()
    setListening(true)
  }

  return (
    <section className="chat">
      <header className="chat-header">
        <button className="icon-btn back-btn" onClick={onBack} title="Back">←</button>
        <div className="chat-header-info" onClick={onOpenInfo}>
          <Avatar emoji={avatar} color={color} size={40} />
          <div>
            <div className="chat-name">{title}</div>
            <div className="chat-status">{typing ? <span className="typing-text">typing…</span> : subtitle}</div>
          </div>
        </div>
        {headerExtra}
      </header>

      {banner && <div className="role-banner">{banner}</div>}

      <div className="messages">
        {note && <div className="encryption-note">{note}</div>}
        {messages.map((m, i) => {
          const prev = messages[i - 1]
          const showSender = m.sender && !m.mine && (!prev || prev.mine || prev.sender?.name !== m.sender.name)
          // Outgoing messages say who they went to whenever that changes.
          const prevOut = messages.slice(0, i).reverse().find((x) => x.mine)
          const showTo = m.sender && m.mine && m.sender.name !== prevOut?.sender?.name
          return (
            <div key={m.id} className={`bubble-row ${m.mine ? 'out' : 'in'}`}>
              <div className={`bubble ${m.card ? 'has-card' : ''}`}>
                {showSender && <div className="sender" style={{ color: m.sender.color }}>{m.sender.avatar} {m.sender.name}</div>}
                {showTo && <div className="sender to">to {m.sender.avatar} {m.sender.name}</div>}
                {m.auto && !m.sender && <div className="auto-tag">🤖 Automated reply</div>}
                <span className="bubble-text">{formatText(m.text)}</span>
                {m.card && renderCard?.(m.card, m)}
                <span className="bubble-meta">
                  {m.auto && m.sender && <span title="Automated reply">🤖 </span>}
                  {formatTime(m.createdAt)}
                  {m.mine && m.status && <Ticks status={m.status} />}
                </span>
              </div>
            </div>
          )
        })}
        {typing && (
          <div className="bubble-row in">
            <div className="bubble typing-bubble"><span /><span /><span /></div>
          </div>
        )}
        <div ref={bottomRef} />
      </div>

      {context}

      {quickReplies.length > 0 && (
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
          placeholder={listening ? 'Listening… speak now' : placeholder}
        />
        {SpeechRecognition && !text.trim() ? (
          <button type="button" className={`send-btn mic ${listening ? 'listening' : ''}`} onClick={toggleVoice}
            title={listening ? 'Stop listening' : 'Speak your message'}>🎤</button>
        ) : (
          <button type="submit" className="send-btn" title="Send" disabled={!text.trim()}>➤</button>
        )}
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

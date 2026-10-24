import { useEffect, useRef, useState } from 'react'
import { businesses as seedBusinesses, initialMessages } from './data/businesses.js'
import { getAutoReplies, nowTime } from './autoReply.js'
import Sidebar from './components/Sidebar.jsx'
import ChatWindow from './components/ChatWindow.jsx'
import AutoReplyPanel from './components/AutoReplyPanel.jsx'
import BusinessInfoPanel from './components/BusinessInfoPanel.jsx'
import EmptyChat from './components/EmptyChat.jsx'

const STORAGE_KEY = 'business-connect-v1'

function loadState() {
  try {
    const saved = JSON.parse(localStorage.getItem(STORAGE_KEY))
    if (saved?.businesses && saved?.messages) return saved
  } catch {
    // ignore corrupt storage and fall back to seed data
  }
  return { businesses: seedBusinesses, messages: initialMessages }
}

let idCounter = Date.now()
const newId = () => `m${idCounter++}`

export default function App() {
  const [state, setState] = useState(loadState)
  const [activeId, setActiveId] = useState(null)
  const [role, setRole] = useState('customer') // 'customer' | 'business'
  const [panel, setPanel] = useState(null) // null | 'info' | 'autoreply'
  const [typing, setTyping] = useState({}) // businessId -> bool
  const [unread, setUnread] = useState({ 'tendai-plumbing': 1 })
  const timers = useRef([])
  const activeIdRef = useRef(activeId)
  activeIdRef.current = activeId

  const { businesses, messages } = state
  const active = businesses.find((b) => b.id === activeId)

  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(state))
  }, [state])

  useEffect(() => () => timers.current.forEach(clearTimeout), [])

  const appendMessage = (businessId, msg) =>
    setState((s) => ({
      ...s,
      messages: { ...s.messages, [businessId]: [...(s.messages[businessId] || []), msg] },
    }))

  const markClientMessagesRead = (businessId) =>
    setState((s) => ({
      ...s,
      messages: {
        ...s.messages,
        [businessId]: s.messages[businessId].map((m) =>
          m.from === 'client' ? { ...m, status: 'read' } : m,
        ),
      },
    }))

  const openChat = (id) => {
    setActiveId(id)
    setUnread((u) => ({ ...u, [id]: 0 }))
    if (panel === 'info') setPanel(null)
  }

  const sendMessage = (text) => {
    const business = active
    const history = messages[business.id] || []

    if (role === 'business') {
      appendMessage(business.id, { id: newId(), from: 'business', text, time: nowTime() })
      return
    }

    appendMessage(business.id, { id: newId(), from: 'client', text, time: nowTime(), status: 'sent' })

    const replies = getAutoReplies(business.autoReply, text, history)
    if (!replies.length) return

    // Simulate the business "typing" before each automated reply.
    timers.current.push(setTimeout(() => {
      markClientMessagesRead(business.id)
      setTyping((t) => ({ ...t, [business.id]: true }))
    }, 500))
    replies.forEach((reply, i) => {
      timers.current.push(setTimeout(() => {
        appendMessage(business.id, { id: newId(), from: 'business', text: reply, time: nowTime(), auto: true })
        if (i === replies.length - 1) setTyping((t) => ({ ...t, [business.id]: false }))
        if (activeIdRef.current !== business.id) {
          setUnread((u) => ({ ...u, [business.id]: (u[business.id] || 0) + 1 }))
        }
      }, 1400 + i * 900))
    })
  }

  const updateAutoReply = (businessId, autoReply) =>
    setState((s) => ({
      ...s,
      businesses: s.businesses.map((b) => (b.id === businessId ? { ...b, autoReply } : b)),
    }))

  const resetDemo = () => {
    if (!confirm('Reset all chats and automated responses to the demo data?')) return
    localStorage.removeItem(STORAGE_KEY)
    setState({ businesses: seedBusinesses, messages: initialMessages })
    setUnread({})
    setPanel(null)
  }

  return (
    <div className="app-backdrop">
      <div className="app">
        <Sidebar
          businesses={businesses}
          messages={messages}
          activeId={activeId}
          unread={unread}
          typing={typing}
          role={role}
          onRoleChange={setRole}
          onSelect={openChat}
          onReset={resetDemo}
        />
        <main className="main">
          {active ? (
            <ChatWindow
              key={active.id}
              business={active}
              messages={messages[active.id] || []}
              typing={!!typing[active.id]}
              role={role}
              onSend={sendMessage}
              onBack={() => setActiveId(null)}
              onOpenInfo={() => setPanel('info')}
              onOpenAutoReply={() => setPanel('autoreply')}
            />
          ) : (
            <EmptyChat />
          )}
        </main>
        {active && panel === 'info' && (
          <BusinessInfoPanel
            business={active}
            onClose={() => setPanel(null)}
            onOpenAutoReply={() => setPanel('autoreply')}
          />
        )}
        {active && panel === 'autoreply' && (
          <AutoReplyPanel
            key={active.id}
            business={active}
            onSave={(autoReply) => updateAutoReply(active.id, autoReply)}
            onClose={() => setPanel(null)}
          />
        )}
      </div>
    </div>
  )
}

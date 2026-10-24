import { useCallback, useEffect, useRef, useState } from 'react'
import { api, formatTime, loadCustomer, loadOwner, saveCustomer, saveOwner } from './api.js'
import Sidebar from './components/Sidebar.jsx'
import ChatWindow from './components/ChatWindow.jsx'
import AutoReplyPanel from './components/AutoReplyPanel.jsx'
import BusinessInfoPanel from './components/BusinessInfoPanel.jsx'
import EmptyChat from './components/EmptyChat.jsx'
import Card from './components/Cards.jsx'
import { NameModal, PinModal } from './components/Modals.jsx'

const ASSISTANT_ID = '__assistant'
// Customers have a single conversation with Business Connect, which routes them to businesses.
const CONCIERGE_ID = 'business-connect'
const CONCIERGE = { name: 'Business Connect', avatar: '💬', color: '#00a884' }
const CONTEXT_KEY = 'bc-chat-context'
const SEEN_KEY = 'bc-chat-seen-id'
const POLL_MS = 3000
const DISCOVERY_QUICK_REPLIES = ['show all businesses', 'I need a haircut', "I'm hungry", 'I need a plumber', 'my bookings']
const OWNER_FILTERS = [
  { label: 'All', test: () => true },
  { label: 'Unread', test: (it) => it.unread > 0 },
]
const OWNER_QUICK_REPLIES = ['today', 'tomorrow', 'upcoming', 'stats', 'menu', 'messages', 'help']

// Adds new messages to a thread, skipping ones already there, in server order.
function merge(thread = [], incoming) {
  const ids = new Set(thread.map((m) => m.id))
  const added = incoming.filter((m) => !ids.has(m.id))
  if (!added.length) return thread
  return [...thread, ...added].sort((a, b) => (typeof a.id === 'number' ? a.id : Infinity) - (typeof b.id === 'number' ? b.id : Infinity))
}

const lastServerId = (thread = []) => thread.reduce((max, m) => (typeof m.id === 'number' && m.id > max ? m.id : max), 0)
const nowISO = () => new Date().toISOString()
let localId = 0
const tempId = () => `local-${++localId}`

export default function App() {
  const [businesses, setBusinesses] = useState([])
  const [loadError, setLoadError] = useState('')
  const [customer, setCustomer] = useState(loadCustomer)
  const [owner, setOwner] = useState(loadOwner) // { token, business }
  const [role, setRole] = useState('customer')
  const [activeId, setActiveId] = useState(CONCIERGE_ID)
  const [panel, setPanel] = useState(null) // null | 'info' | 'autoreply'
  const [modal, setModal] = useState(null) // null | 'pin' | { name: pendingText }

  // Customer side: one thread across the concierge and every business.
  const [thread, setThread] = useState([])
  const synced = useRef(0) // last message id fetched by polling
  const pending = useRef(false) // automated replies are still being "typed"
  const [context, setContextState] = useState(() => localStorage.getItem(CONTEXT_KEY) || '') // business being talked to
  const [seenId, setSeenId] = useState(() => Number(localStorage.getItem(SEEN_KEY)) || 0)
  const [typing, setTyping] = useState({})
  const timers = useRef([])

  const setContext = (id) => {
    localStorage.setItem(CONTEXT_KEY, id)
    setContextState(id)
  }

  // Owner side
  const [ownerConvs, setOwnerConvs] = useState([])
  const [ownerThreads, setOwnerThreads] = useState({}) // customerId -> messages
  const [assistant, setAssistant] = useState([])

  useEffect(() => () => timers.current.forEach(clearTimeout), [])
  useEffect(() => localStorage.setItem(SEEN_KEY, String(seenId)), [seenId])

  const loadBusinesses = useCallback(() => {
    api('/businesses')
      .then((list) => {
        setBusinesses(list)
        setLoadError('')
      })
      .catch((e) => setLoadError(e.message))
  }, [])
  useEffect(loadBusinesses, [loadBusinesses])

  const logout = useCallback(() => {
    if (owner) api('/auth/logout', { method: 'POST', token: owner.token }).catch(() => {})
    saveOwner(null)
    setOwner(null)
    setRole('customer')
    setActiveId(CONCIERGE_ID)
    setPanel(null)
    setAssistant([])
    setOwnerThreads({})
  }, [owner])

  // Owner calls: an expired session sends the owner back to the customer view.
  const ownerApi = useCallback(
    (path, opts = {}) =>
      api(path, { ...opts, token: owner?.token }).catch((e) => {
        if (e.status === 401) logout()
        throw e
      }),
    [owner, logout],
  )

  // ---- Polling ----

  useEffect(() => {
    if (role !== 'customer') return
    const load = () => {
      if (pending.current) return
      // Only polling advances the sync point, so messages sent meanwhile (e.g. an owner reply) are never skipped.
      api(`/customers/${customer.id}/thread?after=${synced.current}`)
        .then((msgs) => {
          if (!msgs.length || pending.current) return
          synced.current = lastServerId(msgs)
          setThread((t) => merge(t, msgs))
        })
        .catch(() => {})
    }
    load()
    const t = setInterval(load, POLL_MS)
    return () => clearInterval(t)
  }, [role, customer.id])

  useEffect(() => {
    if (role !== 'business' || !owner) return
    const load = () => ownerApi('/owner/conversations').then(setOwnerConvs).catch(() => {})
    load()
    const t = setInterval(load, POLL_MS + 1000)
    return () => clearInterval(t)
  }, [role, owner, ownerApi])

  useEffect(() => {
    if (role !== 'business' || !owner || !activeId || activeId === ASSISTANT_ID) return
    const cid = activeId
    const load = () =>
      ownerApi(`/owner/conversations/${encodeURIComponent(cid)}`)
        .then((msgs) => setOwnerThreads((t) => ({ ...t, [cid]: msgs })))
        .catch(() => {})
    load()
    const t = setInterval(load, POLL_MS)
    return () => clearInterval(t)
  }, [role, owner, activeId, ownerApi])

  // Messages in the open chat count as seen.
  useEffect(() => {
    if (role !== 'customer' || activeId !== CONCIERGE_ID) return
    const last = lastServerId(thread)
    if (last > seenId) setSeenId(last)
  }, [role, activeId, thread, seenId])

  // ---- Customer actions ----

  const businessById = Object.fromEntries(businesses.map((b) => [b.id, b]))
  const contextBusiness = businessById[context]

  const sendAsCustomer = async (text, nameOverride) => {
    const name = nameOverride || customer.name
    if (!name) {
      setModal({ name: text })
      return
    }
    const temp = { id: tempId(), from: 'client', businessId: context || CONCIERGE_ID, text, status: 'sent', createdAt: nowISO() }
    setThread((t) => [...t, temp])
    pending.current = true

    let result
    try {
      result = await api('/concierge', { method: 'POST', body: { customerId: customer.id, customerName: name, text, context } })
    } catch (e) {
      setThread((t) => t.map((m) => (m.id === temp.id ? { ...m, text: `${m.text}\n⚠️ Not sent: ${e.message}`, status: null } : m)))
      pending.current = false
      return
    }

    const [mine, ...replies] = result.messages
    setThread((t) => t.map((m) => (m.id === temp.id ? { ...mine, status: 'sent' } : m)))
    setContext(result.context)
    if (!replies.length) {
      pending.current = false
      return
    }

    // Show "typing" before each automated reply.
    timers.current.push(setTimeout(() => {
      setThread((t) => t.map((m) => (m.id === mine.id ? { ...m, status: mine.status } : m)))
      setTyping((ty) => ({ ...ty, [CONCIERGE_ID]: true }))
    }, 400))
    replies.forEach((reply, i) => {
      timers.current.push(setTimeout(() => {
        setThread((t) => merge(t, [reply]))
        if (i === replies.length - 1) {
          setTyping((ty) => ({ ...ty, [CONCIERGE_ID]: false }))
          pending.current = false
        }
      }, 1000 + i * 700))
    })
  }

  const saveName = (name) => {
    const updated = { ...customer, name }
    saveCustomer(updated)
    setCustomer(updated)
    const text = modal?.name
    setModal(null)
    if (text) sendAsCustomer(text, name)
  }

  const customerActions = (businessId) => ({
    choose: (b) => sendAsCustomer(b.name),
    book: async (booking) => {
      const res = await api(`/businesses/${businessId}/reservations`, {
        method: 'POST',
        body: { ...booking, customerId: customer.id, customerName: customer.name },
      })
      setThread((t) => merge(t, [res.message]))
      return res.reservation
    },
    cancelBooking: async (r) => {
      const res = await api(`/businesses/${r.businessId}/reservations/${r.id}/cancel`, {
        method: 'POST',
        body: { customerId: customer.id },
      })
      setThread((t) => merge(t, [res.message]))
      return res.reservation
    },
  })

  // Who sent each message in the shared thread, like the sender names in a WhatsApp group.
  const senderOf = (businessId) => {
    const b = businessById[businessId]
    return b ? { name: b.name, avatar: b.avatar, color: b.color } : CONCIERGE
  }

  // ---- Owner actions ----

  const sendToAssistant = async (text) => {
    const mine = { id: tempId(), from: 'me', text, createdAt: nowISO() }
    setAssistant((a) => [...a, mine])
    setTyping((t) => ({ ...t, [ASSISTANT_ID]: true }))
    try {
      const { replies } = await ownerApi('/owner/assistant', { method: 'POST', body: { text } })
      setTimeout(() => {
        setAssistant((a) => [...a, ...replies.map((r) => ({ ...r, id: tempId(), from: 'bot', createdAt: nowISO() }))])
        setTyping((t) => ({ ...t, [ASSISTANT_ID]: false }))
      }, 600)
    } catch (e) {
      setAssistant((a) => [...a, { id: tempId(), from: 'bot', text: `⚠️ ${e.message}`, createdAt: nowISO() }])
      setTyping((t) => ({ ...t, [ASSISTANT_ID]: false }))
    }
  }

  const replyToCustomer = async (text) => {
    const cid = activeId
    try {
      const msg = await ownerApi(`/owner/conversations/${encodeURIComponent(cid)}`, { method: 'POST', body: { text } })
      setOwnerThreads((t) => ({ ...t, [cid]: merge(t[cid], [msg]) }))
    } catch (e) {
      alert(e.message)
    }
  }

  const ownerActions = {
    toggleOffering: (o) => ownerApi(`/owner/offerings/${o.id}`, { method: 'PUT', body: { ...o, available: !o.available } }),
    setStatus: (r, status) => ownerApi(`/owner/reservations/${r.id}`, { method: 'PUT', body: { status } }),
  }

  const saveAutoReply = async (autoReply) => {
    const current = await ownerApi('/owner/business')
    const updated = await ownerApi('/owner/business', { method: 'PUT', body: { ...current, autoReply } })
    const next = { ...owner, business: updated }
    saveOwner(next)
    setOwner(next)
    loadBusinesses()
  }

  const onLogin = ({ token, business }) => {
    const next = { token, business }
    saveOwner(next)
    setOwner(next)
    setModal(null)
    setRole('business')
    setActiveId(ASSISTANT_ID)
    setAssistant([{
      id: tempId(),
      from: 'bot',
      createdAt: nowISO(),
      text: `Welcome back, ${business.name}! 🔓 I'm your business assistant. Ask me about *today*'s reservations, *upcoming* bookings, free *slots*, your *menu* or this week's *stats*. Type *help* for everything I can do.`,
    }])
  }

  const changeRole = (r) => {
    if (r === role) return
    setPanel(null)
    if (r === 'business' && !owner) {
      setModal('pin')
      return
    }
    setRole(r)
    setActiveId(r === 'business' ? ASSISTANT_ID : CONCIERGE_ID)
    if (r === 'business' && assistant.length === 0) onLogin(owner)
  }

  // ---- Sidebar items ----

  let items
  if (role === 'customer') {
    const last = thread[thread.length - 1]
    items = [{
      id: CONCIERGE_ID, ...CONCIERGE,
      search: 'business connect',
      preview: last ? last.text.split('\n')[0] : 'Find and book local businesses',
      previewMine: last?.from === 'client',
      previewStatus: last?.status,
      time: last ? formatTime(last.createdAt) : '',
      unread: activeId === CONCIERGE_ID ? 0 : thread.filter((m) => m.from !== 'client' && typeof m.id === 'number' && m.id > seenId).length,
      typing: !!typing[CONCIERGE_ID],
    }]
  } else {
    const lastAssistant = assistant[assistant.length - 1]
    items = [
      {
        id: ASSISTANT_ID, name: '🤖 My assistant', avatar: '🤖', color: '#00a884', search: 'assistant',
        preview: lastAssistant ? lastAssistant.text.split('\n')[0] : 'Ask about reservations, stats…',
        time: lastAssistant ? formatTime(lastAssistant.createdAt) : '', unread: 0, typing: !!typing[ASSISTANT_ID],
      },
      ...ownerConvs.map((c) => ({
        id: c.customerId, name: c.customerName || 'Customer', avatar: '🙂', color: '#6b7c85',
        search: c.customerName || 'customer',
        preview: `${c.lastMessage.from === 'business' && c.lastMessage.auto ? '🤖 ' : ''}${c.lastMessage.text.split('\n')[0]}`,
        previewMine: c.lastMessage.from === 'business' && !c.lastMessage.auto,
        time: formatTime(c.lastMessage.createdAt),
        unread: c.customerId === activeId ? 0 : c.unread,
      })),
    ]
  }

  // ---- Main chat ----

  let chat = null
  if (role === 'customer' && activeId === CONCIERGE_ID) {
    const quick = !contextBusiness
      ? DISCOVERY_QUICK_REPLIES
      : !contextBusiness.autoReply.enabled
        ? ['back']
        : contextBusiness.type === 'restaurant'
          ? ['menu', 'book a table', 'hours', 'location', 'my bookings', 'back']
          : ['prices', 'book', 'hours', 'location', 'my bookings', 'back']
    const welcome = {
      id: 'welcome', from: 'business', businessId: CONCIERGE_ID, auto: true,
      text: "Hi 👋 I'm *Business Connect*. Tell me what you need, like *\"I need a haircut\"* or *\"I'm hungry\"*, or tap a business to start chatting with it.",
      card: { type: 'businesses', items: businesses },
    }
    chat = (
      <ChatWindow
        key="concierge"
        title={CONCIERGE.name}
        avatar={CONCIERGE.avatar}
        color={CONCIERGE.color}
        subtitle={contextBusiness ? `Chatting with ${contextBusiness.avatar} ${contextBusiness.name} · click for info` : 'Find, chat with and book local businesses'}
        onOpenInfo={() => contextBusiness && setPanel('info')}
        note="🔒 One chat for every business. Business Connect passes your messages to the business you're talking to. Tap 🎤 to speak."
        messages={[welcome, ...thread].map((m) => ({ ...m, mine: m.from === 'client', sender: senderOf(m.businessId) }))}
        typing={!!typing[CONCIERGE_ID]}
        quickReplies={quick}
        context={contextBusiness && (
          <div className="context-bar">
            <span>Talking to <strong>{contextBusiness.avatar} {contextBusiness.name}</strong></span>
            <button className="mini-btn" onClick={() => setPanel('info')}>ℹ️ Info</button>
            <button className="mini-btn" onClick={() => { setContext(''); setPanel(null) }} title="Your next message goes to Business Connect">✕ Leave</button>
          </div>
        )}
        placeholder={contextBusiness ? `Message ${contextBusiness.name}` : 'What do you need? e.g. "I need a haircut"'}
        onSend={(t) => sendAsCustomer(t)}
        onBack={() => setActiveId(null)}
        renderCard={(card, m) => <Card card={card} businessId={card.businessId || m.businessId} actions={customerActions(card.businessId || m.businessId)} />}
      />
    )
  } else if (role === 'business' && owner && activeId === ASSISTANT_ID) {
    chat = (
      <ChatWindow
        key="assistant"
        title="🤖 My assistant"
        avatar={owner.business.avatar}
        color={owner.business.color}
        subtitle={`Private to ${owner.business.name} · click for business info`}
        onOpenInfo={() => setPanel('info')}
        headerExtra={
          <button className={`autoreply-btn ${owner.business.autoReply.enabled ? 'on' : ''}`} onClick={() => setPanel('autoreply')}>
            🤖 Auto-reply {owner.business.autoReply.enabled ? 'ON' : 'OFF'}
          </button>
        }
        note="🔒 Only you can see this chat. You're signed in with your business PIN."
        messages={assistant.map((m) => ({ ...m, mine: m.from === 'me' }))}
        typing={!!typing[ASSISTANT_ID]}
        quickReplies={OWNER_QUICK_REPLIES}
        placeholder="Ask about reservations, stats, your menu…"
        onSend={sendToAssistant}
        onBack={() => setActiveId(null)}
        renderCard={(card) => <Card card={card} businessId={owner.business.id} actions={ownerActions} />}
      />
    )
  } else if (role === 'business' && owner && activeId) {
    const conv = ownerConvs.find((c) => c.customerId === activeId)
    chat = (
      <ChatWindow
        key={activeId}
        title={conv?.customerName || 'Customer'}
        avatar="🙂"
        color="#6b7c85"
        subtitle="Customer"
        banner={<>You are replying as <strong>{owner.business.name}</strong>. Automated replies still answer new messages.</>}
        messages={(ownerThreads[activeId] || []).map((m) => ({ ...m, mine: m.from === 'business', status: null }))}
        placeholder={`Reply as ${owner.business.name}`}
        onSend={replyToCustomer}
        onBack={() => setActiveId(null)}
        renderCard={(card) => <Card card={card} businessId={owner.business.id} actions={{}} />}
      />
    )
  }

  const infoBusiness = role === 'customer' ? contextBusiness : owner?.business

  return (
    <div className="app-backdrop">
      <div className="app">
        <Sidebar
          key={role}
          items={items}
          activeId={activeId}
          role={role}
          owner={owner}
          onRoleChange={changeRole}
          onSelect={(id) => {
            setActiveId(id)
          }}
          onLogout={logout}
          filters={role === 'customer' ? [] : OWNER_FILTERS}
          emptyText={loadError || 'No chats found'}
        />
        <main className="main">{chat || <EmptyChat role={role} />}</main>
        {infoBusiness && panel === 'info' && (chat || role === 'business') && (
          <BusinessInfoPanel
            business={infoBusiness}
            onClose={() => setPanel(null)}
            onOpenAutoReply={role === 'business' ? () => setPanel('autoreply') : null}
          />
        )}
        {role === 'business' && owner && panel === 'autoreply' && (
          <AutoReplyPanel
            key={owner.business.id}
            business={owner.business}
            onSave={saveAutoReply}
            onClose={() => setPanel(null)}
          />
        )}
      </div>
      {modal === 'pin' && <PinModal businesses={businesses} onLogin={onLogin} onClose={() => setModal(null)} />}
      {modal?.name !== undefined && <NameModal onSave={saveName} onClose={() => setModal(null)} />}
    </div>
  )
}

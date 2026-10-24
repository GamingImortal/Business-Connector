import { useState } from 'react'
import { getAutoReplies } from '../autoReply.js'

// Lets a business configure its automated responses: greeting, keyword rules and fallback.
export default function AutoReplyPanel({ business, onSave, onClose }) {
  const [draft, setDraft] = useState(business.autoReply)
  const [saved, setSaved] = useState(false)
  const [testText, setTestText] = useState('')

  const update = (changes) => {
    setDraft((d) => ({ ...d, ...changes }))
    setSaved(false)
  }
  const updateRule = (id, changes) =>
    update({ rules: draft.rules.map((r) => (r.id === id ? { ...r, ...changes } : r)) })
  const addRule = () =>
    update({ rules: [...draft.rules, { id: `r${Date.now()}`, keywords: '', reply: '' }] })
  const removeRule = (id) => update({ rules: draft.rules.filter((r) => r.id !== id) })

  const [error, setError] = useState('')
  const save = async () => {
    setError('')
    try {
      await onSave({ ...draft, rules: draft.rules.filter((r) => r.keywords.trim() && r.reply.trim()) })
      setSaved(true)
    } catch (e) {
      setError(e.message)
    }
  }

  const preview = testText.trim() ? getAutoReplies({ ...draft, enabled: true }, testText, [{ from: 'business' }]) : []

  return (
    <aside className="side-panel">
      <header className="panel-header">
        <button className="icon-btn" onClick={onClose} title="Close">✕</button>
        <span>Automated responses</span>
      </header>

      <div className="panel-body">
        <p className="panel-intro">
          Set up how <strong>{business.name}</strong> replies automatically when a customer sends a message.
        </p>

        <label className="toggle-row">
          <span>
            <strong>Auto-reply</strong>
            <small>{draft.enabled ? 'Customers get instant replies, including prices, hours and bookings' : 'All replies are manual'}</small>
          </span>
          <input
            type="checkbox"
            className="switch"
            checked={draft.enabled}
            onChange={(e) => update({ enabled: e.target.checked })}
          />
        </label>

        <fieldset disabled={!draft.enabled}>
          <div className="field">
            <label>👋 Greeting message <small>sent on a customer's first message</small></label>
            <textarea rows={2} value={draft.greeting} onChange={(e) => update({ greeting: e.target.value })} />
          </div>

          <div className="field">
            <label>🔑 Keyword replies <small>checked before the automatic answers about prices, hours and bookings</small></label>
            {draft.rules.map((rule) => (
              <div className="rule" key={rule.id}>
                <div className="rule-top">
                  <input
                    placeholder="Keywords, comma separated (e.g. price, cost)"
                    value={rule.keywords}
                    onChange={(e) => updateRule(rule.id, { keywords: e.target.value })}
                  />
                  <button type="button" className="icon-btn danger" onClick={() => removeRule(rule.id)} title="Delete rule">🗑</button>
                </div>
                <textarea
                  rows={3}
                  placeholder="Reply message"
                  value={rule.reply}
                  onChange={(e) => updateRule(rule.id, { reply: e.target.value })}
                />
              </div>
            ))}
            <button type="button" className="link-btn" onClick={addRule}>+ Add keyword reply</button>
          </div>

          <div className="field">
            <label>💬 Fallback message <small>sent when no keyword matches</small></label>
            <textarea rows={2} value={draft.fallback} onChange={(e) => update({ fallback: e.target.value })} />
          </div>

          <div className="field">
            <label>🧪 Test your replies</label>
            <input
              placeholder="Type a sample customer message…"
              value={testText}
              onChange={(e) => setTestText(e.target.value)}
            />
            {preview.map((p, i) => (
              <div key={i} className="preview-bubble">{p}</div>
            ))}
          </div>
        </fieldset>
      </div>

      <footer className="panel-footer">
        <span className="saved-note">{error || (saved ? '✓ Saved' : '')}</span>
        <button className="primary-btn" onClick={save}>Save</button>
      </footer>
    </aside>
  )
}

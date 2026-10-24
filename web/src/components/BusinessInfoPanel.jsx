import { useEffect, useState } from 'react'
import Avatar from './Avatar.jsx'
import { api, money } from '../api.js'

export default function BusinessInfoPanel({ business, onClose, onOpenAutoReply }) {
  const [offerings, setOfferings] = useState([])

  useEffect(() => {
    api(`/businesses/${business.id}`).then((d) => setOfferings(d.offerings)).catch(() => setOfferings([]))
  }, [business.id])

  return (
    <aside className="side-panel">
      <header className="panel-header">
        <button className="icon-btn" onClick={onClose} title="Close">✕</button>
        <span>Business info</span>
      </header>
      <div className="panel-body">
        <div className="info-hero">
          <Avatar emoji={business.avatar} color={business.color} size={120} />
          <h2>{business.name}</h2>
          <p className="muted">{business.category} · ⭐ {business.rating ? business.rating.toFixed(1) : 'New'}</p>
        </div>
        {business.about && (
          <div className="info-section">
            <h4>About</h4>
            <p>{business.about}</p>
          </div>
        )}
        <div className="info-section">
          <h4>📍 Location</h4>
          <p>{business.location || '—'}</p>
        </div>
        {business.phone && (
          <div className="info-section">
            <h4>📞 Phone</h4>
            <p>{business.phone}</p>
          </div>
        )}
        <div className="info-section">
          <h4>🕗 Hours</h4>
          <p>{business.hoursText}</p>
        </div>
        {offerings.length > 0 && (
          <div className="info-section">
            <h4>{business.type === 'restaurant' ? '🍽️ Menu' : '📋 Services & prices'}</h4>
            {offerings.map((o) => (
              <p key={o.id} className="price-line">
                <span>{o.name}</span>
                <strong>{money(o.price)}</strong>
              </p>
            ))}
          </div>
        )}
        {onOpenAutoReply && (
          <div className="info-section clickable" onClick={onOpenAutoReply}>
            <h4>🤖 Automated responses</h4>
            <p>
              {business.autoReply.enabled ? `On · ${business.autoReply.rules.length} keyword replies` : 'Off'}{' '}
              <span className="link-btn">Manage ›</span>
            </p>
          </div>
        )}
      </div>
    </aside>
  )
}

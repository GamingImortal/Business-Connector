import Avatar from './Avatar.jsx'

export default function BusinessInfoPanel({ business, onClose, onOpenAutoReply }) {
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
          <p className="muted">{business.category} · ⭐ {business.rating}</p>
        </div>
        <div className="info-section">
          <h4>About</h4>
          <p>{business.about}</p>
        </div>
        <div className="info-section">
          <h4>📍 Location</h4>
          <p>{business.location}</p>
        </div>
        <div className="info-section">
          <h4>🕗 Hours</h4>
          <p>{business.hours}</p>
        </div>
        <div className="info-section clickable" onClick={onOpenAutoReply}>
          <h4>🤖 Automated responses</h4>
          <p>
            {business.autoReply.enabled
              ? `On · ${business.autoReply.rules.length} keyword replies`
              : 'Off'}{' '}
            <span className="link-btn">Manage ›</span>
          </p>
        </div>
      </div>
    </aside>
  )
}

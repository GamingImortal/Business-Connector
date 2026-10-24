export default function EmptyChat() {
  return (
    <div className="empty-chat">
      <div className="empty-icon">💬</div>
      <h1>Business Connect Web</h1>
      <p>
        Chat with local businesses, get instant answers from their automated replies,
        and book services, all in one place.
      </p>
      <p className="muted">Select a business on the left to start chatting.</p>
      <div className="empty-footer">🔒 Demo app: messages are stored only in this browser</div>
    </div>
  )
}

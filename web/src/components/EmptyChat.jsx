export default function EmptyChat({ role }) {
  return (
    <div className="empty-chat">
      <div className="empty-icon">💬</div>
      <h1>Business Connect Web</h1>
      {role === 'customer' ? (
        <p>
          Chat with local businesses: ask for prices or the menu, check opening hours,
          and book appointments or tables, all in one place. Tap 🎤 to speak instead of typing.
        </p>
      ) : (
        <p>
          Ask your 🤖 assistant about today's reservations, free slots or this week's stats,
          or reply to your customers' chats.
        </p>
      )}
      <p className="muted">Select a chat on the left to start.</p>
      <div className="empty-footer">🔒 Demo app: data is stored by the Business Connect backend</div>
    </div>
  )
}

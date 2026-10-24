// Previews which keyword replies would be sent for a customer's message.
// The backend does the real matching (and also answers from listed prices, hours and bookings).
export function getAutoReplies(autoReply, text, history) {
  if (!autoReply?.enabled) return []

  const replies = []
  const isFirstContact = !history.some((m) => m.from === 'business')
  if (isFirstContact && autoReply.greeting.trim()) {
    replies.push(autoReply.greeting)
  }

  const lower = text.toLowerCase()
  const matched = autoReply.rules.filter((rule) =>
    rule.keywords
      .split(',')
      .map((k) => k.trim().toLowerCase())
      .filter(Boolean)
      .some((k) => new RegExp(`\\b${escapeRegExp(k)}\\b`).test(lower)),
  )

  if (matched.length) {
    replies.push(...matched.map((r) => r.reply))
  } else if (autoReply.fallback.trim()) {
    replies.push(autoReply.fallback)
  }
  return replies
}

function escapeRegExp(s) {
  return s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

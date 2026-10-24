// Works out which automated replies a business sends for a customer's message.
// Returns an array of reply texts (possibly empty).
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

export function nowTime() {
  return new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false })
}

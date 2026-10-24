// Seed data: five businesses, each with its own automated-response setup.
// A rule matches when any of its comma-separated keywords appears in the customer's message.

export const businesses = [
  {
    id: 'fresh-cuts',
    name: 'Fresh Cuts Barbershop',
    category: 'Barber',
    avatar: '💈',
    color: '#d97706',
    rating: 4.8,
    location: 'Avondale',
    hours: 'Mon–Sat, 08:00–18:00',
    about: 'Fades, beard trims and kids cuts. Walk-ins welcome.',
    autoReply: {
      enabled: true,
      greeting: 'Hi 👋 Welcome to Fresh Cuts Barbershop! How can we help you today?',
      fallback: "Thanks for your message! A barber will get back to you shortly. Type *prices*, *hours* or *book* for quick answers.",
      rules: [
        { id: 'r1', keywords: 'price, prices, cost, how much', reply: '💇‍♂️ Our prices:\n• Fade – $6\n• Fade + Beard – $8\n• Kids cut – $4\n• Beard trim – $3' },
        { id: 'r2', keywords: 'hours, open, close, time', reply: '🕗 We are open Mon–Sat, 08:00–18:00. Closed on Sundays.' },
        { id: 'r3', keywords: 'book, booking, appointment, slot', reply: '📅 Available slots this Saturday: 09:00, 11:30, 14:00, 16:30. Reply with the time you want.' },
        { id: 'r4', keywords: 'where, location, address', reply: '📍 Shop 12, Avondale Shopping Centre.' },
      ],
    },
  },
  {
    id: 'tendai-plumbing',
    name: 'Tendai Plumbing',
    category: 'Plumber',
    avatar: '🔧',
    color: '#2563eb',
    rating: 4.6,
    location: 'Harare CBD',
    hours: 'Mon–Sun, 07:00–20:00 (emergencies 24/7)',
    about: 'Leaks, geysers, blocked drains and new installations.',
    autoReply: {
      enabled: true,
      greeting: 'Hello, Tendai Plumbing here 🔧. Please describe the problem and your area.',
      fallback: "Got it 👍 I'm currently on a job and will reply within the hour. For emergencies type *urgent*.",
      rules: [
        { id: 'r1', keywords: 'urgent, emergency, burst, flood', reply: '🚨 Emergency line: call +263 77 000 0000. Callout within 60 minutes.' },
        { id: 'r2', keywords: 'price, cost, quote, how much', reply: '💵 Callout fee is $15. Final quote is given after inspection.' },
        { id: 'r3', keywords: 'geyser, hot water', reply: '♨️ Geyser repairs start from $25 and installations from $120 (excluding parts).' },
      ],
    },
  },
  {
    id: 'glow-beauty',
    name: 'Glow Beauty Studio',
    category: 'Hair & Nails',
    avatar: '💅',
    color: '#db2777',
    rating: 4.9,
    location: 'Borrowdale',
    hours: 'Tue–Sun, 09:00–19:00',
    about: 'Braids, wigs, gel nails and lashes.',
    autoReply: {
      enabled: true,
      greeting: 'Hey lovely 💖 Welcome to Glow Beauty Studio!',
      fallback: 'Thank you! One of our stylists will reply soon ✨',
      rules: [
        { id: 'r1', keywords: 'braids, braid, knotless', reply: '✨ Knotless braids: $25 (medium), $35 (small). Takes 4–6 hours.' },
        { id: 'r2', keywords: 'nails, gel, manicure', reply: '💅 Gel overlay $10, acrylic set $18, pedicure $12.' },
        { id: 'r3', keywords: 'book, appointment, available', reply: '📅 Next openings: Thu 10:00, Fri 13:00, Sat 09:00. Which one works for you?' },
      ],
    },
  },
  {
    id: 'maths-ruvimbo',
    name: 'Maths with Ruvimbo',
    category: 'Tutor',
    avatar: '📚',
    color: '#7c3aed',
    rating: 4.7,
    location: 'Online & Mount Pleasant',
    hours: 'Weekdays 15:00–20:00, Sat 08:00–13:00',
    about: 'O & A Level Maths tutoring, one-on-one or small groups.',
    autoReply: {
      enabled: true,
      greeting: 'Hi! 📚 Thanks for reaching out to Maths with Ruvimbo.',
      fallback: "Thanks! I'll reply after my current lesson. Type *rates* or *schedule* for quick info.",
      rules: [
        { id: 'r1', keywords: 'rate, rates, price, fee, cost', reply: '💵 $8/hour one-on-one, $5/hour per student in groups of 3–5.' },
        { id: 'r2', keywords: 'schedule, when, times, available', reply: '🗓️ Weekdays 15:00–20:00 and Saturdays 08:00–13:00.' },
        { id: 'r3', keywords: 'online, zoom, virtual', reply: '💻 Yes, online lessons are on Google Meet. A link is sent after booking.' },
      ],
    },
  },
  {
    id: 'sparkle-clean',
    name: 'Sparkle Cleaning Co.',
    category: 'Cleaning',
    avatar: '🧹',
    color: '#059669',
    rating: 4.5,
    location: 'Greater Harare',
    hours: 'Mon–Sat, 07:00–17:00',
    about: 'Home, office and move-out cleaning. Eco-friendly products.',
    autoReply: {
      enabled: false,
      greeting: 'Welcome to Sparkle Cleaning Co. 🧽',
      fallback: "Thanks for your enquiry! We'll send you a quote shortly.",
      rules: [
        { id: 'r1', keywords: 'price, quote, cost', reply: '🏠 Standard home clean from $30, deep clean from $55, office from $40.' },
      ],
    },
  },
]

// Opening conversation history so the chat list looks lived-in.
export const initialMessages = {
  'fresh-cuts': [
    { id: 'm1', from: 'client', text: 'Hi, are you open today?', time: '09:58', status: 'read' },
    { id: 'm2', from: 'business', text: '🕗 We are open Mon–Sat, 08:00–18:00. Closed on Sundays.', time: '09:58', auto: true },
  ],
  'tendai-plumbing': [
    { id: 'm1', from: 'business', text: 'Your geyser repair is booked for Monday at 10:00 ✅', time: 'Mon', },
  ],
  'glow-beauty': [],
  'maths-ruvimbo': [
    { id: 'm1', from: 'client', text: 'Thanks for the lesson!', time: 'Sun', status: 'read' },
    { id: 'm2', from: 'business', text: 'You are welcome 😊 See you next week.', time: 'Sun' },
  ],
  'sparkle-clean': [],
}

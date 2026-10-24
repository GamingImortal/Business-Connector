# Business Connect

> **Find, chat with, and book local service providers — all from one chat window.**

Business Connect is a chat-first marketplace that connects customers with local service businesses — hairdressers, barbers, plumbers, electricians, tutors, mechanics, cleaners, photographers, nail techs, personal trainers and more.

Customers already talk to businesses through messaging apps. Business Connect keeps that familiar experience and adds what's missing: **discovery, verified ratings, and real bookings** — so a conversation can turn straight into a confirmed appointment.

---

## 1. The Problem

**For customers**
- Finding a trustworthy provider depends on word of mouth or scattered social media pages.
- Ratings, if they exist at all, are spread across many platforms and are hard to trust.
- Booking means back-and-forth messages ("Are you free Saturday?" … "What about 2pm?"), and often no confirmation.

**For small businesses**
- Most can't afford, or don't want, a website or booking system.
- Their phone's messaging app becomes an unorganised mix of enquiries, bookings and personal chats.
- No-shows and forgotten appointments cost them money.
- It's hard to build a public reputation beyond regular customers.

## 2. The Solution

A single **chat interface** (modelled on WhatsApp Web) where:

1. **Customers** search for a service, compare providers by rating, price and distance, then start a chat.
2. **Inside the chat**, the customer can view services, check availability, book, pay a deposit and leave a review — using interactive message cards instead of endless text.
3. **Businesses** manage enquiries, bookings and their reputation from the same inbox, with no extra tools to learn.

> The original idea — *connecting hairdressers with clients, rating each hairdresser, and booking via chat* — is our **launch category**. The platform is designed so any appointment-based service business can join.

---

## 3. Who It's For

| Persona | Example | What they need |
|---|---|---|
| **Customer** | Thandi needs a braiding appointment this Saturday | Find a well-rated stylist nearby, see prices, book quickly |
| **Solo provider** | Tendai, a freelance plumber | Get new leads, cut down on scheduling messages, build a reputation |
| **Small business** | "Fresh Cuts" barbershop with 4 barbers | Route bookings to the right staff member, track ratings per staff member |
| **Business owner / admin** | Salon owner | View bookings, staff performance and reviews |

### Example categories
💇 Hair & Beauty · 🔧 Plumbing & Electrical · 🚗 Mechanics · 🧹 Cleaning · 📚 Tutoring · 📸 Photography · 🏋️ Fitness · 🐾 Pet Care · 🎉 Events & Catering

---

## 4. Core Features

### 4.1 Chat-first interface (the main interface)
- Two-panel layout like WhatsApp Web: **conversation list** on the left, **active chat** on the right.
- Ordinary text messages plus **rich interactive cards**:
  - 📋 **Service menu card** — services, durations and prices
  - 📅 **Availability card** — tap a time slot to book
  - ✅ **Booking confirmation card** — date, time, provider, location, with *Reschedule* / *Cancel* buttons
  - ⭐ **Review request card** — sent automatically after the appointment
  - 💳 **Payment / deposit card** (later phase)
- A built-in **"Business Connect Assistant"** chat for discovery: *"Find me a barber near Avondale under $10"*.

### 4.2 Discovery
- Search by category, location, price range, rating and availability.
- Results appear as provider cards in the assistant chat, each with a **Chat** button.

### 4.3 Ratings & reviews
- Only customers with a completed booking can leave a review (verified reviews).
- Ratings at **business level** and **individual staff level** (e.g. each hairdresser in a salon).
- Rating breakdown: quality, punctuality, value, communication.

### 4.4 Booking
- Real-time availability based on the provider's working hours.
- Booking, rescheduling and cancelling inside the chat.
- Automatic reminders (24h and 1h before).
- Optional deposit to reduce no-shows.

### 4.5 Business profile
- Logo, description, location, opening hours, gallery/portfolio, service catalogue, staff list.
- Shown as a "contact info" side panel, like WhatsApp's contact details panel.

### 4.6 Business tools
- Quick replies and saved messages.
- Labels: *New lead*, *Booked*, *Completed*, *Follow up*.
- Simple dashboard: bookings this week, average rating, repeat customers.

---

## 5. Mock Chat Interface (Hackathon MVP)

For the hackathon, we'll build a **mock chat interface that looks and feels like WhatsApp Web**. It uses mock data and scripted responses, with no real messaging backend.

### 5.1 Layout

```
┌───────────────────────────────────────────────────────────────────────────────┐
│ ┌──────────────────────────┐ ┌──────────────────────────────────────────────┐ │
│ │ 👤 Me          💬  ⋮      │ │ 🟢 Fresh Cuts Barbershop  ⭐ 4.8   🔍  📅  ⋮  │ │
│ ├──────────────────────────┤ ├──────────────────────────────────────────────┤ │
│ │ 🔍 Search or start chat  │ │                                              │ │
│ │ [All][Unread][Bookings]  │ │   ┌────────────────────────────────┐         │ │
│ ├──────────────────────────┤ │   │ Hi! Do you have a slot on Sat? │  10:02  │ │
│ │ 🤖 Business Connect      │ │   └────────────────────────────────┘         │ │
│ │    Assistant      10:15  │ │                                              │ │
│ │    Found 3 barbers near… │ │         ┌──────────────────────────────────┐ │ │
│ ├──────────────────────────┤ │         │ 📅 Available on Sat 26 Oct       │ │ │
│ │ 💈 Fresh Cuts     10:03  │ │         │  [09:00] [11:30] [14:00] [16:30] │ │ │
│ │    📅 Available on Sat…  │ │         │  with: Tino ⭐4.9 ▾              │ │ │
│ ├──────────────────────────┤ │         └──────────────────────────── 10:03 ┘ │ │
│ │ 🔧 Tendai Plumbing  Mon  │ │                                              │ │
│ │    ✅ Booking confirmed  │ │   ┌────────────────────────────────┐         │ │
│ ├──────────────────────────┤ │   │ 11:30 please ✓✓                │  10:04  │ │
│ │ 📚 MathsWithRuvimbo Sun  │ │   └────────────────────────────────┘         │ │
│ │    ⭐ How was your…      │ │                                              │ │
│ │                          │ ├──────────────────────────────────────────────┤ │
│ │                          │ │ 😊  📎  [ Type a message…             ]  🎤  │ │
│ └──────────────────────────┘ └──────────────────────────────────────────────┘ │
└───────────────────────────────────────────────────────────────────────────────┘
```

### 5.2 UI elements to copy from WhatsApp Web
- Green/teal header bar and a light, patterned chat background.
- Sent messages in light green bubbles on the right; received messages in white bubbles on the left.
- Timestamps and ✓ / ✓✓ read ticks.
- Chat list with avatar, name, last-message preview, time and unread badge.
- Search bar and filter chips at the top of the chat list.
- A contact-info drawer that slides in from the right (used here for the **business profile**).
- Dark mode (nice to have).

### 5.3 What's mocked
| Area | Mock approach |
|---|---|
| Users & businesses | Static JSON seed data (around 10 businesses across 4 categories) |
| Messages | In-memory/localStorage conversation history |
| Business replies | Scripted auto-replies triggered by keywords or card actions |
| Assistant search | Keyword matching over the seed data (*"barber"*, *"plumber"*, *"near"*, *"under $"*) |
| Availability | Generated time slots from each business's opening hours |
| Bookings & reviews | Saved to localStorage, shown in chats and on profiles |
| Payments | Not included in the MVP; a "Deposit paid ✓" badge is shown for demo purposes |

### 5.4 Demo script (around 3 minutes)
1. Open the app; it looks like WhatsApp Web.
2. In the **Assistant** chat, type: *"I need a haircut near Avondale on Saturday"*.
3. The assistant returns 3 provider cards with ratings, prices and distance.
4. Tap **Chat** on *Fresh Cuts*; a new conversation opens.
5. The business sends its **service menu card**; pick *"Fade + Beard – $8"*.
6. The **availability card** appears; choose barber *Tino ⭐4.9* at *11:30*.
7. A **booking confirmation card** appears, and the chat moves under the *Bookings* filter.
8. Fast-forward ("simulate appointment completed"); a **review card** appears and you leave 5 ⭐.
9. Open the business profile drawer: the new review and the updated rating are shown.
10. Switch to **Business view** to see the same conversation from the barber's side, with labels and a bookings summary.

---

## 6. MVP Scope

**In scope (hackathon)**
- [ ] WhatsApp Web-style layout (chat list + chat window + profile drawer)
- [ ] Assistant chat with mock search
- [ ] Interactive cards: service menu, availability, booking confirmation, review
- [ ] Ratings per business and per staff member
- [ ] Customer view / Business view toggle
- [ ] Seed data for multiple business categories
- [ ] Responsive layout (on mobile it collapses to a single panel, like WhatsApp)

**Out of scope (for now)**
- Real authentication, real-time messaging and payments
- Real WhatsApp integration
- Maps and geolocation

---

## 7. Suggested Tech Stack

| Layer | Choice | Why |
|---|---|---|
| Frontend | React + Vite + TypeScript | Fast to set up, component-based chat UI |
| Styling | Tailwind CSS | Quick to match the WhatsApp look |
| State | Zustand or React Context + localStorage | Simple persistence without a backend |
| Mock data | `/src/data/*.json` | Businesses, staff, services, reviews |
| Icons | Lucide / emoji | Lightweight |
| Later: backend | Node.js (Express/Nest) + PostgreSQL + WebSockets | Real-time chat and bookings |
| Later: AI assistant | LLM-powered natural-language search & booking | Replaces keyword matching |
| Later: channels | WhatsApp Business API | Meet customers where they already are |

### Proposed data model (mock)
```
Business    { id, name, category, location, rating, reviewCount, hours, services[], staff[], gallery[] }
Staff       { id, businessId, name, avatar, rating, specialties[] }
Service     { id, businessId, name, durationMins, price }
Customer    { id, name, avatar }
Conversation{ id, customerId, businessId, labels[], messages[] }
Message     { id, from, type: text|serviceMenu|availability|booking|review, payload, timestamp, status }
Booking     { id, customerId, businessId, staffId, serviceId, startTime, status }
Review      { id, bookingId, businessId, staffId, stars, breakdown{}, comment }
```

---

## 8. Business Model

| Revenue stream | Description |
|---|---|
| **Freemium subscription** | Free: profile + chat + up to 30 bookings/month. **Pro** (~$5–10/month): unlimited bookings, reminders, staff accounts, analytics |
| **Booking fee** | Small commission (e.g. 3–5%) on deposits/payments processed through the platform |
| **Featured listings** | Businesses pay to appear at the top of assistant search results in their area |
| **Partnerships** | Product suppliers (e.g. hair products, tools) advertise to relevant businesses |

### Why it can win
- **Familiar UX**: no learning curve, because it looks like the chat app people already use daily.
- **Built for informal and small businesses**: no website needed; works from a phone.
- **Trust layer**: verified, per-staff ratings that word of mouth can't provide at scale.
- **Horizontal by design**: starts with hairdressers and expands category by category on the same chat engine.

---

## 9. Success Metrics
- Chat-to-booking conversion rate
- No-show rate (target: fewer no-shows than before using reminders and deposits)
- Number of active businesses and categories
- Average time from first message to confirmed booking
- Review completion rate after appointments
- Repeat booking rate

---

## 10. Roadmap

| Phase | Focus |
|---|---|
| **0 – Hackathon** | Mock WhatsApp Web-style chat, assistant search, booking and review cards, seed data |
| **1 – Pilot** | Real backend, auth, real-time chat; onboard 20 hairdressers/barbers in one city |
| **2 – Expand** | Add plumbing, tutoring and cleaning categories; reminders by SMS/WhatsApp; deposits |
| **3 – Scale** | AI assistant for natural-language booking, WhatsApp Business API channel, business analytics, mobile apps |

---

## 11. Getting Started

The mock WhatsApp Web-style app lives in [`web/`](web/) (React + Vite, using yarn).

```bash
cd web
yarn install
yarn dev        # http://localhost:5173
```

- Chat with 5 demo businesses as a **Customer**; replies come from each business's automated responses.
- Click **🤖 Auto-reply** in a chat header to set the greeting, keyword replies and fallback message.
- Switch **View as → Business** to reply manually as that business.
- Data is saved in the browser's localStorage; use **↺** to reset the demo.

---

## 12. Team
_Add team members and roles here._

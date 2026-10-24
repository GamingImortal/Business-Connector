package main

import "fmt"

// Demo PIN for every seeded business.
const demoPIN = "1234"

func week(open, close string, closedDays ...int) []DayHours {
	hours := make([]DayHours, 7)
	for d := range hours {
		hours[d] = DayHours{Weekday: d, Open: open, Close: close}
	}
	for _, d := range closedDays {
		hours[d].Closed = true
	}
	return hours
}

type seedBusiness struct {
	Business
	offerings []Offering
}

func seedData() []seedBusiness {
	svc := func(category, name string, price float64, mins int, desc string) Offering {
		return Offering{Kind: "service", Category: category, Name: name, Price: price, DurationMins: mins, Description: desc, Available: true}
	}
	product := func(category, name string, price float64, desc string) Offering {
		return Offering{Kind: "product", Category: category, Name: name, Price: price, Description: desc, Available: true}
	}
	dish := func(category, name string, price float64, desc string) Offering {
		return Offering{Kind: "menu", Category: category, Name: name, Price: price, Description: desc, Available: true}
	}

	return []seedBusiness{
		{
			Business: Business{
				ID: "fresh-cuts", Name: "Fresh Cuts Barbershop", Category: "Barber", Type: "service",
				Avatar: "💈", Color: "#d97706", Rating: 4.8, Location: "Shop 12, Avondale Shopping Centre",
				Phone: "+263 77 111 2222", About: "Fades, beard trims and kids cuts. Walk-ins welcome.",
				Hours: week("08:00", "18:00", 0), SlotMinutes: 30, Capacity: 2,
				AutoReply: AutoReply{
					Enabled:  true,
					Greeting: "Hi 👋 Welcome to Fresh Cuts Barbershop! How can we help you today?",
					Fallback: "Thanks for your message! A barber will get back to you shortly. Type *prices*, *hours* or *book* for quick answers.",
					Rules: []Rule{
						{ID: "r1", Keywords: "walk in, walk-in, walkin", Reply: "🚶 Walk-ins are welcome, but booked customers are served first."},
					},
				},
			},
			offerings: []Offering{
				svc("Cuts", "Fade", 6, 30, "Skin, low, mid or high fade"),
				svc("Cuts", "Fade + Beard", 8, 45, "Fade with a beard line-up and trim"),
				svc("Cuts", "Kids cut", 4, 30, "Under 12s"),
				svc("Grooming", "Beard trim", 3, 15, ""),
				product("Products", "Beard oil", 7, "30ml, cedarwood scent"),
			},
		},
		{
			Business: Business{
				ID: "mamas-kitchen", Name: "Mama's Kitchen", Category: "Restaurant", Type: "restaurant",
				Avatar: "🍲", Color: "#b91c1c", Rating: 4.7, Location: "45 Samora Machel Ave, Harare",
				Phone: "+263 77 555 0101", About: "Home-style Zimbabwean food. Family tables, outdoor seating and takeaway.",
				Hours: week("11:00", "22:00", 1), SlotMinutes: 90, Capacity: 40,
				AutoReply: AutoReply{
					Enabled:  true,
					Greeting: "Mhoroi! 🍲 Welcome to Mama's Kitchen. Ask for the *menu* or *book a table*.",
					Fallback: "Thanks! Our team will reply shortly. You can also ask for the *menu*, *hours* or *book a table for 4 on Friday at 19:00*.",
					Rules: []Rule{
						{ID: "r1", Keywords: "takeaway, take away, delivery, deliver", Reply: "🛵 Takeaway is available all day. Delivery within 5km for $2 — order by phone on +263 77 555 0101."},
						{ID: "r2", Keywords: "vegetarian, vegan", Reply: "🥬 Vegetarian options: Muriwo une Dovi, Covo & Sadza, Bean stew. Just ask your server!"},
					},
				},
			},
			offerings: []Offering{
				dish("Mains", "Sadza & Beef Stew", 6, "Slow-cooked beef with greens"),
				dish("Mains", "Oxtail", 12, "Served with sadza or rice"),
				dish("Mains", "Grilled Bream", 10, "Whole Kariba bream with chips"),
				dish("Mains", "Muriwo une Dovi", 5, "Greens in peanut butter sauce (vegetarian)"),
				dish("Sides", "Covo & Sadza", 3, ""),
				dish("Desserts", "Malva Pudding", 4, "With custard"),
				dish("Drinks", "Mazoe Orange", 1.5, ""),
				dish("Drinks", "Maheu", 1, "Traditional maize drink"),
			},
		},
		{
			Business: Business{
				ID: "tendai-plumbing", Name: "Tendai Plumbing", Category: "Plumber", Type: "service",
				Avatar: "🔧", Color: "#2563eb", Rating: 4.6, Location: "Harare CBD (callouts across Harare)",
				Phone: "+263 77 000 0000", About: "Leaks, geysers, blocked drains and new installations.",
				Hours: week("07:00", "19:00"), SlotMinutes: 120, Capacity: 1,
				AutoReply: AutoReply{
					Enabled:  true,
					Greeting: "Hello, Tendai Plumbing here 🔧. Please describe the problem and your area.",
					Fallback: "Got it 👍 I'm currently on a job and will reply within the hour. For emergencies type *urgent*.",
					Rules: []Rule{
						{ID: "r1", Keywords: "urgent, emergency, burst, flood", Reply: "🚨 Emergency line: call +263 77 000 0000. Callout within 60 minutes, 24/7."},
					},
				},
			},
			offerings: []Offering{
				svc("Callouts", "Callout & inspection", 15, 60, "Final quote given after inspection"),
				svc("Geysers", "Geyser repair", 25, 120, "Excluding parts"),
				svc("Geysers", "Geyser installation", 120, 240, "Excluding the geyser"),
				svc("Drains", "Blocked drain", 30, 90, ""),
			},
		},
		{
			Business: Business{
				ID: "glow-beauty", Name: "Glow Beauty Studio", Category: "Hair & Nails", Type: "service",
				Avatar: "💅", Color: "#db2777", Rating: 4.9, Location: "Borrowdale Village, Shop 4",
				Phone: "+263 78 222 3333", About: "Braids, wigs, gel nails and lashes.",
				Hours: week("09:00", "19:00", 1), SlotMinutes: 60, Capacity: 3,
				AutoReply: AutoReply{
					Enabled:  true,
					Greeting: "Hey lovely 💖 Welcome to Glow Beauty Studio!",
					Fallback: "Thank you! One of our stylists will reply soon ✨",
					Rules:    []Rule{},
				},
			},
			offerings: []Offering{
				svc("Hair", "Knotless braids (medium)", 25, 300, "Hair included"),
				svc("Hair", "Knotless braids (small)", 35, 360, "Hair included"),
				svc("Hair", "Wig install", 20, 90, ""),
				svc("Nails", "Gel overlay", 10, 60, ""),
				svc("Nails", "Acrylic set", 18, 90, ""),
				svc("Nails", "Pedicure", 12, 60, ""),
			},
		},
		{
			Business: Business{
				ID: "maths-ruvimbo", Name: "Maths with Ruvimbo", Category: "Tutor", Type: "service",
				Avatar: "📚", Color: "#7c3aed", Rating: 4.7, Location: "Online & Mount Pleasant",
				Phone: "+263 71 444 5555", About: "O & A Level Maths tutoring, one-on-one or small groups.",
				Hours:       append(week("15:00", "20:00", 0)[:6], DayHours{Weekday: 6, Open: "08:00", Close: "13:00"}),
				SlotMinutes: 60, Capacity: 1,
				AutoReply: AutoReply{
					Enabled:  true,
					Greeting: "Hi! 📚 Thanks for reaching out to Maths with Ruvimbo.",
					Fallback: "Thanks! I'll reply after my current lesson. Type *rates* or *book* for quick info.",
					Rules: []Rule{
						{ID: "r1", Keywords: "online, zoom, virtual, meet", Reply: "💻 Yes, online lessons are on Google Meet. A link is sent after booking."},
					},
				},
			},
			offerings: []Offering{
				svc("Lessons", "One-on-one lesson", 8, 60, "O or A Level"),
				svc("Lessons", "Group lesson (per student)", 5, 60, "Groups of 3–5"),
				product("Resources", "Past paper pack", 3, "PDF, 10 years of papers with worked answers"),
			},
		},
		{
			Business: Business{
				ID: "sparkle-clean", Name: "Sparkle Cleaning Co.", Category: "Cleaning", Type: "service",
				Avatar: "🧹", Color: "#059669", Rating: 4.5, Location: "Greater Harare",
				Phone: "+263 77 666 7777", About: "Home, office and move-out cleaning. Eco-friendly products.",
				Hours: week("07:00", "17:00", 0), SlotMinutes: 120, Capacity: 2,
				AutoReply: AutoReply{
					Enabled:  false,
					Greeting: "Welcome to Sparkle Cleaning Co. 🧽",
					Fallback: "Thanks for your enquiry! We'll send you a quote shortly.",
					Rules:    []Rule{},
				},
			},
			offerings: []Offering{
				svc("Cleaning", "Standard home clean", 30, 180, ""),
				svc("Cleaning", "Deep clean", 55, 300, ""),
				svc("Cleaning", "Office clean", 40, 180, "Per 100m²"),
			},
		},
	}
}

// Seed fills an empty database with demo businesses and a few bookings.
func (s *Store) Seed() error {
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM businesses`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}

	firstOffering := map[string]int64{}
	for _, sb := range seedData() {
		if err := s.CreateBusiness(sb.Business, demoPIN); err != nil {
			return fmt.Errorf("seed %s: %w", sb.ID, err)
		}
		for _, o := range sb.offerings {
			o.BusinessID = sb.ID
			if err := s.CreateOffering(&o); err != nil {
				return err
			}
			if _, ok := firstOffering[sb.ID]; !ok {
				firstOffering[sb.ID] = o.ID
			}
		}
	}

	// A few bookings around today so the calendar isn't empty.
	t := today()
	d := func(offset int) string { return t.AddDate(0, 0, offset).Format(dateLayout) }
	ref := func(id string) *int64 { v := firstOffering[id]; return &v }
	bookings := []Reservation{
		{BusinessID: "fresh-cuts", CustomerName: "Farai", CustomerPhone: "+263 77 123 4567", OfferingID: ref("fresh-cuts"), Date: d(0), Time: "16:00", PartySize: 1},
		{BusinessID: "fresh-cuts", CustomerName: "Kuda", OfferingID: ref("fresh-cuts"), Date: d(1), Time: "09:00", PartySize: 1},
		{BusinessID: "fresh-cuts", CustomerName: "Tawanda", Date: d(-1), Time: "11:00", PartySize: 1, Status: "completed"},
		{BusinessID: "mamas-kitchen", CustomerName: "Chipo", CustomerPhone: "+263 78 987 6543", Date: d(0), Time: "18:30", PartySize: 4, Notes: "Birthday 🎂"},
		{BusinessID: "mamas-kitchen", CustomerName: "Moyo family", Date: d(0), Time: "20:00", PartySize: 8},
		{BusinessID: "mamas-kitchen", CustomerName: "Rudo", Date: d(1), Time: "12:30", PartySize: 2, Notes: "Outdoor table"},
		{BusinessID: "glow-beauty", CustomerName: "Thandi", OfferingID: ref("glow-beauty"), Date: d(2), Time: "10:00", PartySize: 1},
		{BusinessID: "tendai-plumbing", CustomerName: "Mrs Ncube", OfferingID: ref("tendai-plumbing"), Date: d(1), Time: "09:00", PartySize: 1, Notes: "Leaking geyser, Avondale"},
	}
	for _, r := range bookings {
		if r.Status == "" {
			r.Status = "confirmed"
		}
		r.Source = "dashboard"
		if err := s.CreateReservation(&r); err != nil {
			return err
		}
	}
	return nil
}

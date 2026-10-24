package main

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ---- Cards: structured messages the chat UI renders as interactive cards ----

type OfferingsCard struct {
	Type  string     `json:"type"` // "offerings"
	Title string     `json:"title"`
	Items []Offering `json:"items"`
	Owner bool       `json:"owner,omitempty"` // show availability toggles
}

type SlotsCard struct {
	Type         string     `json:"type"` // "slots"
	BusinessID   string     `json:"businessId"`
	BusinessType string     `json:"businessType"`
	Date         string     `json:"date"`
	DateLabel    string     `json:"dateLabel"`
	PartySize    int        `json:"partySize"`
	Slots        []Slot     `json:"slots"`
	Services     []Offering `json:"services,omitempty"`
}

type BookingCard struct {
	Type        string      `json:"type"` // "booking"
	Reservation Reservation `json:"reservation"`
}

type BookingsCard struct {
	Type         string        `json:"type"` // "bookings"
	Title        string        `json:"title"`
	Reservations []Reservation `json:"reservations"`
	Owner        bool          `json:"owner,omitempty"` // show owner actions instead of "cancel"
}

var (
	errSlotTaken = errors.New("we're fully booked at that time")
	errBadSlot   = errors.New("that time is outside our opening hours")
	errClosed    = errors.New("we're closed that day")
	errPast      = errors.New("that time has already passed")
)

// Book validates and stores a reservation. Bookings from chat must fit an open slot;
// owners booking from the dashboard can override the calendar.
func (s *Store) Book(b Business, r *Reservation) error {
	d, err := parseDate(r.Date)
	if err != nil {
		return fmt.Errorf("invalid date %q", r.Date)
	}
	if minutes(r.Time) < 0 {
		return fmt.Errorf("invalid time %q", r.Time)
	}
	r.Time = hhmm(minutes(r.Time))
	r.BusinessID = b.ID
	r.PartySize = max(r.PartySize, 1)
	if r.Status == "" {
		r.Status = "confirmed"
	}
	if r.Source != "dashboard" {
		if err := s.checkBookable(b, d, minutes(r.Time), r.PartySize); err != nil {
			return err
		}
	}
	return s.CreateReservation(r)
}

// checkBookable reports why a customer can't book a time, or nil if they can.
func (s *Store) checkBookable(b Business, d time.Time, start, partySize int) error {
	if d.Before(today()) || (d.Equal(today()) && start <= nowMinutes()) {
		return errPast
	}
	open, lastStart, ok := openWindow(b, d)
	if !ok {
		return errClosed
	}
	if start < open || start > lastStart {
		return errBadSlot
	}
	day := d.Format(dateLayout)
	res, err := s.ListReservations(b.ID, day, day)
	if err != nil {
		return err
	}
	if remainingAt(b, start, res) < units(b, partySize) {
		return errSlotTaken
	}
	return nil
}

// isBookingError reports whether err is a reason to show the customer other times.
func isBookingError(err error) bool {
	return errors.Is(err, errSlotTaken) || errors.Is(err, errBadSlot) || errors.Is(err, errClosed) || errors.Is(err, errPast)
}

func money(v float64) string {
	if v == float64(int64(v)) {
		return fmt.Sprintf("$%d", int64(v))
	}
	return fmt.Sprintf("$%.2f", v)
}

func reservationLine(b Business, r Reservation) string {
	parts := []string{fmt.Sprintf("#%d", r.ID), r.Time, r.CustomerName}
	if b.Type == "restaurant" {
		parts = append(parts, fmt.Sprintf("%d guests", r.PartySize))
	}
	if r.OfferingName != "" {
		parts = append(parts, r.OfferingName)
	}
	if r.Status != "confirmed" {
		parts = append(parts, strings.ReplaceAll(r.Status, "_", " "))
	}
	return strings.Join(parts, " · ")
}

func ruleMatches(rule Rule, lower string) bool {
	for _, k := range strings.Split(rule.Keywords, ",") {
		k = strings.ToLower(strings.TrimSpace(k))
		if k != "" && containsWord(lower, k) {
			return true
		}
	}
	return false
}

// mentionedOfferings finds listed items named in a message, e.g. "how much is the fade?".
// Exact names win, so "chicken stew" doesn't also return "Beef Stew".
func mentionedOfferings(offerings []Offering, lower string) []Offering {
	var exact, partial []Offering
	for _, o := range offerings {
		name := strings.ToLower(o.Name)
		if strings.Contains(lower, name) {
			exact = append(exact, o)
			continue
		}
		for _, w := range strings.FieldsFunc(name, func(r rune) bool { return !('a' <= r && r <= 'z') }) {
			if len(w) > 3 && containsWord(lower, w) {
				partial = append(partial, o)
				break
			}
		}
	}
	if len(exact) > 0 {
		return exact
	}
	return partial
}

func offeringLine(o Offering) string {
	line := fmt.Sprintf("• %s – %s", o.Name, money(o.Price))
	if o.DurationMins > 0 {
		line += fmt.Sprintf(" (%d min)", o.DurationMins)
	}
	if !o.Available {
		line += " — currently unavailable"
	}
	return line
}

func offeringsTitle(b Business) string {
	if b.Type == "restaurant" {
		return "🍽️ Our menu"
	}
	return "📋 Services & prices"
}

func customerHelp(b Business) string {
	book := "*book* an appointment"
	if b.Type == "restaurant" {
		book = "*reserve* a table (e.g. \"table for 4 on Friday at 19:00\")"
	}
	what := "*prices*"
	if b.Type == "restaurant" {
		what = "the *menu*"
	}
	return fmt.Sprintf("You can ask me for %s, our *hours*, *location*, or %s. Type *my bookings* to see your reservations.", what, book)
}

// CustomerReplies works out the automated answers to a customer's message, using the
// business's own keyword rules first and then the data listed on the dashboard.
func (s *Store) CustomerReplies(b Business, customerID, text string, firstContact bool) ([]Reply, error) {
	ar := b.AutoReply
	if !ar.Enabled {
		return nil, nil
	}
	var replies []Reply
	if firstContact && strings.TrimSpace(ar.Greeting) != "" {
		replies = append(replies, Reply{Text: ar.Greeting})
	}

	lower := strings.ToLower(text)
	matched := false
	for _, rule := range ar.Rules {
		if ruleMatches(rule, lower) {
			replies = append(replies, Reply{Text: rule.Reply})
			matched = true
		}
	}
	if matched {
		return replies, nil
	}

	smart, err := s.smartReplies(b, customerID, lower)
	if err != nil {
		return nil, err
	}
	if len(smart) > 0 {
		return append(replies, smart...), nil
	}

	if containsAny(lower, "hi", "hello", "hey", "help", "morning", "afternoon", "evening") {
		return append(replies, Reply{Text: customerHelp(b)}), nil
	}
	if strings.TrimSpace(ar.Fallback) != "" {
		return append(replies, Reply{Text: ar.Fallback}), nil
	}
	return append(replies, Reply{Text: customerHelp(b)}), nil
}

func (s *Store) smartReplies(b Business, customerID, lower string) ([]Reply, error) {
	var replies []Reply

	if containsAny(lower, "cancel") || strings.Contains(lower, "my booking") ||
		strings.Contains(lower, "my reservation") || strings.Contains(lower, "my appointment") {
		res, err := s.CustomerReservations(b.ID, customerID, today().Format(dateLayout))
		if err != nil {
			return nil, err
		}
		if len(res) == 0 {
			return []Reply{{Text: "You don't have any upcoming bookings with us. " + customerHelp(b)}}, nil
		}
		text := "📅 Your upcoming bookings:"
		if containsAny(lower, "cancel") {
			text = "Which booking would you like to cancel?"
		}
		return []Reply{{Text: text, Card: BookingsCard{Type: "bookings", Title: "Your bookings", Reservations: res}}}, nil
	}

	offerings, err := s.ListOfferings(b.ID, false)
	if err != nil {
		return nil, err
	}

	if containsAny(lower, "book", "booking", "reserve", "reservation", "table", "appointment", "slot", "slots",
		"available", "availability", "free", "tonight") {
		r, err := s.bookingReply(b, customerID, lower, offerings)
		if err != nil {
			return nil, err
		}
		return []Reply{r}, nil
	}

	wantsList := containsAny(lower, "menu", "price", "prices", "cost", "costs", "services", "service", "products",
		"product", "offer", "offers", "food", "drinks", "eat", "list", "charge", "rates", "rate") ||
		strings.Contains(lower, "how much")
	mentioned := mentionedOfferings(offerings, lower)
	if len(mentioned) > 0 {
		lines := make([]string, len(mentioned))
		for i, o := range mentioned {
			lines[i] = offeringLine(o)
			if o.Description != "" {
				lines[i] += "\n   " + o.Description
			}
		}
		replies = append(replies, Reply{Text: strings.Join(lines, "\n")})
	} else if wantsList {
		var available []Offering
		for _, o := range offerings {
			if o.Available {
				available = append(available, o)
			}
		}
		if len(available) == 0 {
			replies = append(replies, Reply{Text: "We haven't listed our prices yet. A member of the team will reply shortly."})
		} else {
			replies = append(replies, Reply{Text: offeringsTitle(b), Card: OfferingsCard{Type: "offerings", Title: offeringsTitle(b), Items: available}})
		}
	}

	if containsAny(lower, "hours", "open", "opening", "close", "closing", "closed", "when") {
		replies = append(replies, Reply{Text: "🕗 Our hours: " + b.HoursText})
	}
	if containsAny(lower, "where", "location", "address", "directions") {
		replies = append(replies, Reply{Text: "📍 " + b.Location})
	}
	if containsAny(lower, "phone", "call", "number", "contact", "whatsapp") && b.Phone != "" {
		replies = append(replies, Reply{Text: "📞 " + b.Phone})
	}
	return replies, nil
}

// bookingReply books straight away when the message names an open time ("book 11:30 on Saturday"),
// and otherwise offers the open slots as a card.
func (s *Store) bookingReply(b Business, customerID, lower string, offerings []Offering) (Reply, error) {
	party := 2
	if b.Type != "restaurant" {
		party = 1
	}
	if n, ok := findPartySize(lower); ok && b.Type == "restaurant" {
		party = n
	}
	if b.Type == "restaurant" && party > b.Capacity {
		return Reply{Text: fmt.Sprintf("For groups larger than %d, please call us on %s and we'll arrange it 🙏", b.Capacity, b.Phone)}, nil
	}

	var services []Offering
	for _, o := range offerings {
		if o.Kind == "service" && o.Available {
			services = append(services, o)
		}
	}

	date, hasDate := findDate(lower)
	wanted, hasTime := findTime(lower)
	if hasTime && !hasDate && minutes(wanted) <= nowMinutes() {
		date = today().AddDate(0, 0, 1)
	}

	if hasTime {
		r := Reservation{
			CustomerID:   customerID,
			CustomerName: s.CustomerName(customerID),
			Date:         date.Format(dateLayout),
			Time:         wanted,
			PartySize:    party,
			Source:       "chat",
		}
		if r.CustomerName == "" {
			r.CustomerName = "Customer"
		}
		if m := mentionedOfferings(services, lower); len(m) > 0 {
			r.OfferingID = &m[0].ID
		}
		err := s.Book(b, &r)
		if err == nil {
			r, _ = s.GetReservation(b.ID, r.ID)
			return Reply{Text: bookingConfirmation(b, r), Card: BookingCard{Type: "booking", Reservation: r}}, nil
		}
		if !isBookingError(err) {
			return Reply{}, err
		}
		// Offer the open times that day, or the next day with space.
		requested := dateLabel(date)
		date, slots, serr := s.NextAvailableDay(b, date, party)
		if serr != nil {
			return Reply{}, serr
		}
		return Reply{
			Text: fmt.Sprintf("Sorry, I can't book %s on %s: %s. Here are the open times for %s:", wanted, requested, err, dateLabel(date)),
			Card: SlotsCard{Type: "slots", BusinessID: b.ID, BusinessType: b.Type, Date: date.Format(dateLayout), DateLabel: dateLabel(date),
				PartySize: party, Slots: slots, Services: services},
		}, nil
	}

	var slots []Slot
	var err error
	if hasDate {
		slots, err = s.SlotsFor(b, date, party)
	} else {
		date, slots, err = s.NextAvailableDay(b, today(), party)
	}
	if err != nil {
		return Reply{}, err
	}
	free := 0
	for _, sl := range slots {
		if sl.Available {
			free++
		}
	}
	text := fmt.Sprintf("📅 Open times for %s — tap one to book:", dateLabel(date))
	if b.Type == "restaurant" {
		text = fmt.Sprintf("📅 Tables for %d on %s — tap a time to reserve:", party, dateLabel(date))
	}
	if free == 0 {
		text = fmt.Sprintf("😕 We're fully booked or closed on %s. Try another day, e.g. \"book tomorrow\".", dateLabel(date))
	}
	return Reply{
		Text: text,
		Card: SlotsCard{Type: "slots", BusinessID: b.ID, BusinessType: b.Type, Date: date.Format(dateLayout), DateLabel: dateLabel(date),
			PartySize: party, Slots: slots, Services: services},
	}, nil
}

func bookingConfirmation(b Business, r Reservation) string {
	d, _ := parseDate(r.Date)
	what := ""
	if r.OfferingName != "" {
		what = " for " + r.OfferingName
	}
	if b.Type == "restaurant" {
		return fmt.Sprintf("✅ Table for %d reserved on %s at %s. See you then, %s!", r.PartySize, dateLabel(d), r.Time, r.CustomerName)
	}
	return fmt.Sprintf("✅ You're booked%s on %s at %s. See you then, %s!", what, dateLabel(d), r.Time, r.CustomerName)
}

// ---- Owner assistant ----

func ownerHelp(b Business) string {
	items := "*sold out <item>* / *available <item>*"
	return "I'm your business assistant 🤖. Try:\n" +
		"• *today* / *tomorrow* / *saturday* – reservations for a day\n" +
		"• *upcoming* – the next bookings\n" +
		"• *slots friday* – what's still free\n" +
		"• *confirm 12*, *done 12*, *no show 12*, *cancel 12* – update a booking\n" +
		"• *menu* – your listed " + map[bool]string{true: "menu", false: "services"}[b.Type == "restaurant"] + "\n" +
		"• " + items + " – change what customers can order\n" +
		"• *stats* – this week's summary\n" +
		"• *messages* – unread customer chats"
}

var statusCommands = []struct {
	words  []string
	status string
	label  string
}{
	{[]string{"no show", "noshow", "no-show"}, "no_show", "marked as no-show 🚫"},
	{[]string{"cancel"}, "cancelled", "cancelled ❌"},
	{[]string{"done", "complete", "completed", "finished"}, "completed", "marked as completed ✔️"},
	{[]string{"confirm", "confirmed", "restore"}, "confirmed", "confirmed ✅"},
}

// OwnerReplies answers a business owner's questions about their own business.
func (s *Store) OwnerReplies(b Business, text string) ([]Reply, error) {
	lower := strings.ToLower(strings.TrimSpace(text))

	// Status changes: "confirm 12", "cancel #12", "no show 12".
	if m := idRe.FindStringSubmatch(lower); m != nil && !strings.Contains(lower, ":") {
		for _, c := range statusCommands {
			for _, w := range c.words {
				if strings.Contains(lower, w) {
					var id int64
					fmt.Sscan(m[1], &id)
					r, err := s.GetReservation(b.ID, id)
					if errors.Is(err, errNotFound) {
						return []Reply{{Text: fmt.Sprintf("I couldn't find booking #%d.", id)}}, nil
					}
					if err != nil {
						return nil, err
					}
					r.Status = c.status
					if err := s.UpdateReservation(r); err != nil {
						return nil, err
					}
					return []Reply{{Text: fmt.Sprintf("Booking #%d (%s, %s %s) %s", r.ID, r.CustomerName, r.Date, r.Time, c.label)}}, nil
				}
			}
		}
	}

	// Availability of listed items: "sold out oxtail", "available oxtail".
	for _, prefix := range []string{"sold out", "unavailable", "hide", "available", "back on", "show"} {
		if strings.HasPrefix(lower, prefix+" ") {
			return s.toggleOffering(b, strings.TrimSpace(strings.TrimPrefix(lower, prefix)),
				prefix == "available" || prefix == "back on" || prefix == "show")
		}
	}

	switch {
	case containsAny(lower, "help", "hi", "hello", "hey", "commands") && len(lower) < 20:
		return []Reply{{Text: ownerHelp(b)}}, nil
	case containsAny(lower, "stats", "summary", "report", "week", "revenue", "performance"):
		return s.ownerStats(b)
	case containsAny(lower, "messages", "unread", "chats", "inbox"):
		return s.ownerInbox(b)
	case containsAny(lower, "menu", "services", "prices", "products", "offerings", "items", "catalogue"):
		list, err := s.ListOfferings(b.ID, false)
		if err != nil {
			return nil, err
		}
		if len(list) == 0 {
			return []Reply{{Text: "You haven't listed anything yet. Add items in the dashboard under *Offerings*."}}, nil
		}
		return []Reply{{Text: offeringsTitle(b), Card: OfferingsCard{Type: "offerings", Title: offeringsTitle(b), Items: list, Owner: true}}}, nil
	case containsAny(lower, "upcoming", "next"):
		res, err := s.ListReservations(b.ID, today().Format(dateLayout), "")
		if err != nil {
			return nil, err
		}
		var upcoming []Reservation
		for _, r := range res {
			if r.Status == "confirmed" && len(upcoming) < 10 {
				upcoming = append(upcoming, r)
			}
		}
		if len(upcoming) == 0 {
			return []Reply{{Text: "No upcoming bookings yet."}}, nil
		}
		return []Reply{{Text: fmt.Sprintf("📅 Next %d bookings:", len(upcoming)),
			Card: BookingsCard{Type: "bookings", Title: "Upcoming", Reservations: upcoming, Owner: true}}}, nil
	case containsAny(lower, "slots", "free", "availability", "available", "capacity", "space"):
		date, _ := findDate(lower)
		party := 1
		if n, ok := findPartySize(lower); ok {
			party = n
		}
		slots, err := s.SlotsFor(b, date, party)
		if err != nil {
			return nil, err
		}
		if len(slots) == 0 {
			return []Reply{{Text: fmt.Sprintf("You're closed on %s.", dateLabel(date))}}, nil
		}
		unit := "bookings"
		if b.Type == "restaurant" {
			unit = "seats"
		}
		lines := []string{fmt.Sprintf("🗓️ Availability for %s (%s left per slot):", dateLabel(date), unit)}
		for _, sl := range slots {
			lines = append(lines, fmt.Sprintf("• %s – %d of %d free", sl.Time, sl.Remaining, b.Capacity))
		}
		return []Reply{{Text: strings.Join(lines, "\n")}}, nil
	}

	if date, ok := findDate(lower); ok || containsAny(lower, "bookings", "booking", "reservations", "reservation",
		"appointments", "schedule", "calendar", "agenda", "diary") {
		return s.ownerDay(b, date)
	}
	return []Reply{{Text: "Sorry, I didn't catch that. " + ownerHelp(b)}}, nil
}

func (s *Store) ownerDay(b Business, date time.Time) ([]Reply, error) {
	d := date.Format(dateLayout)
	res, err := s.ListReservations(b.ID, d, d)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return []Reply{{Text: fmt.Sprintf("No bookings for %s yet.", dateLabel(date))}}, nil
	}
	active, guests := 0, 0
	for _, r := range res {
		if r.Status == "confirmed" || r.Status == "completed" {
			active++
			guests += r.PartySize
		}
	}
	summary := fmt.Sprintf("📅 %s: %d booking(s)", dateLabel(date), active)
	if b.Type == "restaurant" {
		summary += fmt.Sprintf(", %d guests", guests)
	}
	return []Reply{{Text: summary, Card: BookingsCard{Type: "bookings", Title: dateLabel(date), Reservations: res, Owner: true}}}, nil
}

func (s *Store) toggleOffering(b Business, name string, available bool) ([]Reply, error) {
	list, err := s.ListOfferings(b.ID, false)
	if err != nil {
		return nil, err
	}
	matches := mentionedOfferings(list, name)
	if len(matches) == 0 {
		return []Reply{{Text: fmt.Sprintf("I couldn't find \"%s\" in your listings. Type *menu* to see them.", name)}}, nil
	}
	var names []string
	for _, o := range matches {
		o.Available = available
		if err := s.UpdateOffering(o); err != nil {
			return nil, err
		}
		names = append(names, o.Name)
	}
	state := "is now hidden from customers (sold out) 🚫"
	if available {
		state = "is available to customers again ✅"
	}
	return []Reply{{Text: strings.Join(names, ", ") + " " + state}}, nil
}

func (s *Store) ownerStats(b Business) ([]Reply, error) {
	t := today()
	monday := t.AddDate(0, 0, -((int(t.Weekday()) + 6) % 7))
	sunday := monday.AddDate(0, 0, 6)
	res, err := s.ListReservations(b.ID, monday.Format(dateLayout), sunday.Format(dateLayout))
	if err != nil {
		return nil, err
	}
	offerings, err := s.ListOfferings(b.ID, false)
	if err != nil {
		return nil, err
	}
	price := map[int64]float64{}
	for _, o := range offerings {
		price[o.ID] = o.Price
	}

	counts := map[string]int{}
	customers := map[string]bool{}
	guests := 0
	revenue := 0.0
	for _, r := range res {
		counts[r.Status]++
		if r.Status == "confirmed" || r.Status == "completed" {
			customers[r.CustomerName] = true
			guests += r.PartySize
			if r.OfferingID != nil {
				revenue += price[*r.OfferingID]
			}
		}
	}
	unread := 0
	convs, _ := s.BusinessConversations(b.ID)
	for _, c := range convs {
		unread += c.Unread
	}

	lines := []string{
		fmt.Sprintf("📊 This week (%s – %s)", monday.Format("2 Jan"), sunday.Format("2 Jan")),
		fmt.Sprintf("• Bookings: %d confirmed, %d completed", counts["confirmed"], counts["completed"]),
		fmt.Sprintf("• Cancelled: %d · No-shows: %d", counts["cancelled"], counts["no_show"]),
		fmt.Sprintf("• Customers: %d", len(customers)),
	}
	if b.Type == "restaurant" {
		lines = append(lines, fmt.Sprintf("• Guests: %d", guests))
	} else {
		lines = append(lines, fmt.Sprintf("• Estimated revenue: %s", money(revenue)))
	}
	lines = append(lines, fmt.Sprintf("• Rating: ⭐ %.1f", b.Rating), fmt.Sprintf("• Unread messages: %d", unread))
	return []Reply{{Text: strings.Join(lines, "\n")}}, nil
}

func (s *Store) ownerInbox(b Business) ([]Reply, error) {
	convs, err := s.BusinessConversations(b.ID)
	if err != nil {
		return nil, err
	}
	var lines []string
	for _, c := range convs {
		if c.Unread > 0 {
			name := c.CustomerName
			if name == "" {
				name = "Customer"
			}
			lines = append(lines, fmt.Sprintf("• %s (%d): \"%s\"", name, c.Unread, firstLine(c.LastMessage.Text)))
		}
	}
	if len(lines) == 0 {
		return []Reply{{Text: "📭 No unread messages. You're all caught up!"}}, nil
	}
	return []Reply{{Text: "📬 Unread chats:\n" + strings.Join(lines, "\n")}}, nil
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")
	if r := []rune(line); len(r) > 60 {
		return string(r[:57]) + "…"
	}
	return line
}

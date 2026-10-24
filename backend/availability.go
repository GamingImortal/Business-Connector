package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const dateLayout = "2006-01-02"

// clock is replaceable in tests.
var clock = time.Now

func today() time.Time {
	t := clock()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func parseDate(s string) (time.Time, error) {
	return time.ParseInLocation(dateLayout, s, clock().Location())
}

// minutes converts "HH:MM" to minutes after midnight, or -1 if invalid.
func minutes(hhmm string) int {
	parts := strings.Split(hhmm, ":")
	if len(parts) != 2 {
		return -1
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil || h < 0 || h > 24 || m < 0 || m > 59 {
		return -1
	}
	return h*60 + m
}

func hhmm(mins int) string { return fmt.Sprintf("%02d:%02d", mins/60, mins%60) }

// normalizeHours always returns seven entries, Sunday first.
func normalizeHours(in []DayHours) []DayHours {
	out := make([]DayHours, 7)
	for i := range out {
		out[i] = DayHours{Weekday: i, Closed: true, Open: "09:00", Close: "17:00"}
	}
	for _, h := range in {
		if h.Weekday >= 0 && h.Weekday < 7 {
			out[h.Weekday] = h
		}
	}
	return out
}

// hoursSummary renders hours like "Mon–Sat 08:00–18:00 · Sun closed".
func hoursSummary(hours []DayHours) string {
	key := func(h DayHours) string {
		if h.Closed {
			return "closed"
		}
		return h.Open + "–" + h.Close
	}
	// Start the week on Monday.
	order := []int{1, 2, 3, 4, 5, 6, 0}
	var parts []string
	for i := 0; i < len(order); {
		j := i
		for j+1 < len(order) && key(hours[order[j+1]]) == key(hours[order[i]]) {
			j++
		}
		days := weekdayNames[order[i]][:3]
		if j > i {
			days += "–" + weekdayNames[order[j]][:3]
		}
		parts = append(parts, days+" "+key(hours[order[i]]))
		i = j + 1
	}
	return strings.Join(parts, " · ")
}

// units is how much of a slot's capacity a booking uses.
func units(b Business, partySize int) int {
	if b.Type == "restaurant" {
		return max(partySize, 1)
	}
	return 1
}

// remainingAt is how much capacity is left for a booking starting at start (minutes after midnight).
func remainingAt(b Business, start int, reservations []Reservation) int {
	step := max(b.SlotMinutes, 15)
	used := 0
	for _, r := range reservations {
		if r.Status == "cancelled" || r.Status == "no_show" {
			continue
		}
		rs := minutes(r.Time)
		if rs < start+step && start < rs+step {
			used += units(b, r.PartySize)
		}
	}
	return max(b.Capacity-used, 0)
}

// openWindow returns the opening and latest start time on a date, or ok=false if closed.
func openWindow(b Business, date time.Time) (open, lastStart int, ok bool) {
	h := b.Hours[int(date.Weekday())]
	open, closing := minutes(h.Open), minutes(h.Close)
	lastStart = closing - max(b.SlotMinutes, 15)
	if h.Closed || open < 0 || lastStart < open {
		return 0, 0, false
	}
	return open, lastStart, true
}

func nowMinutes() int { n := clock(); return n.Hour()*60 + n.Minute() }

// Slots lists suggested booking times on a date. A slot is available when the reservations
// overlapping it leave enough capacity for the requested party size.
func Slots(b Business, date time.Time, reservations []Reservation, partySize int) []Slot {
	open, lastStart, ok := openWindow(b, date)
	if !ok {
		return []Slot{}
	}
	earliest := -1
	if date.Equal(today()) {
		earliest = nowMinutes()
	}
	need := units(b, partySize)
	slots := []Slot{}
	for start := open; start <= lastStart; start += max(b.SlotMinutes, 15) {
		if start <= earliest {
			continue
		}
		remaining := remainingAt(b, start, reservations)
		slots = append(slots, Slot{Time: hhmm(start), Remaining: remaining, Available: remaining >= need})
	}
	return slots
}

func (s *Store) SlotsFor(b Business, date time.Time, partySize int) ([]Slot, error) {
	d := date.Format(dateLayout)
	res, err := s.ListReservations(b.ID, d, d)
	if err != nil {
		return nil, err
	}
	return Slots(b, date, res, partySize), nil
}

// NextAvailableDay finds the first day (within two weeks, from start) with a free slot.
func (s *Store) NextAvailableDay(b Business, start time.Time, partySize int) (time.Time, []Slot, error) {
	for i := 0; i < 14; i++ {
		d := start.AddDate(0, 0, i)
		slots, err := s.SlotsFor(b, d, partySize)
		if err != nil {
			return d, nil, err
		}
		for _, sl := range slots {
			if sl.Available {
				return d, slots, nil
			}
		}
	}
	return start, []Slot{}, nil
}

func dateLabel(d time.Time) string {
	switch {
	case d.Equal(today()):
		return d.Format("Mon 2 Jan") + " (today)"
	case d.Equal(today().AddDate(0, 0, 1)):
		return d.Format("Mon 2 Jan") + " (tomorrow)"
	}
	return d.Format("Mon 2 Jan")
}

var (
	isoDateRe   = regexp.MustCompile(`\b(\d{4}-\d{2}-\d{2})\b`)
	timeRe      = regexp.MustCompile(`\b([01]?\d|2[0-3])[:h.]([0-5]\d)\b|\b(1[0-2]|0?[1-9])\s*(am|pm)\b`)
	partySizeRe = regexp.MustCompile(`\b(?:for|party of|table for)\s+(\d{1,2})\b|\b(\d{1,2})\s*(?:people|persons|guests|pax|of us)\b`)
	idRe        = regexp.MustCompile(`#?(\d+)`)
)

// findDate looks for "today", "tomorrow", a weekday name or an ISO date in a message.
func findDate(text string) (time.Time, bool) {
	t := today()
	switch {
	case strings.Contains(text, "today") || strings.Contains(text, "tonight"):
		return t, true
	case strings.Contains(text, "day after tomorrow"):
		return t.AddDate(0, 0, 2), true
	case strings.Contains(text, "tomorrow"):
		return t.AddDate(0, 0, 1), true
	}
	if m := isoDateRe.FindStringSubmatch(text); m != nil {
		if d, err := parseDate(m[1]); err == nil {
			return d, true
		}
	}
	for i, name := range weekdayNames {
		lower := strings.ToLower(name)
		if containsWord(text, lower) || containsWord(text, lower[:3]) {
			diff := (i - int(t.Weekday()) + 7) % 7
			return t.AddDate(0, 0, diff), true
		}
	}
	return t, false
}

// findTime returns "HH:MM" for "11:30", "14h00" or "3pm".
func findTime(text string) (string, bool) {
	m := timeRe.FindStringSubmatch(text)
	if m == nil {
		return "", false
	}
	if m[1] != "" {
		h, _ := strconv.Atoi(m[1])
		mm, _ := strconv.Atoi(m[2])
		return hhmm(h*60 + mm), true
	}
	h, _ := strconv.Atoi(m[3])
	if m[4] == "pm" && h != 12 {
		h += 12
	}
	if m[4] == "am" && h == 12 {
		h = 0
	}
	return hhmm(h * 60), true
}

func findPartySize(text string) (int, bool) {
	// Remove times first so "for 11:30" is not read as a party of 11.
	m := partySizeRe.FindStringSubmatch(timeRe.ReplaceAllString(text, " "))
	if m == nil {
		return 0, false
	}
	n, _ := strconv.Atoi(m[1] + m[2])
	if n < 1 {
		return 0, false
	}
	return n, true
}

func containsWord(text, word string) bool {
	return regexp.MustCompile(`\b` + regexp.QuoteMeta(word) + `\b`).MatchString(text)
}

func containsAny(text string, words ...string) bool {
	for _, w := range words {
		if containsWord(text, w) {
			return true
		}
	}
	return false
}

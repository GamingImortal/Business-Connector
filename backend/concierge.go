package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// The Business Connect concierge is the single conversation customers chat in. It finds
// businesses and routes each message to the business being talked to. Messages routed to a
// business are stored in that business's conversation, so owners see them in their inbox;
// the concierge's own messages are stored under a hidden "business-connect" business.
const conciergeID = "business-connect"

func (s *Store) EnsureConcierge() error {
	salt := randomHex(16)
	_, err := s.db.Exec(`INSERT OR IGNORE INTO businesses (id, name, category, type, avatar, color, about, hours, auto_reply, pin_salt, pin_hash, created_at)
		VALUES (?, 'Business Connect', 'Assistant', 'concierge', '💬', '#00a884', 'Find and book local businesses.', '[]', '{}', ?, ?, ?)`,
		conciergeID, salt, hashPIN(salt, randomHex(16)), now())
	return err
}

// BusinessSummary is a business as shown in the concierge's directory cards.
type BusinessSummary struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Avatar    string  `json:"avatar"`
	Color     string  `json:"color"`
	Category  string  `json:"category"`
	Type      string  `json:"type"`
	Rating    float64 `json:"rating"`
	Location  string  `json:"location"`
	HoursText string  `json:"hoursText"`
}

type BusinessesCard struct {
	Type  string            `json:"type"` // "businesses"
	Items []BusinessSummary `json:"items"`
}

func businessesCard(list []Business) BusinessesCard {
	items := make([]BusinessSummary, len(list))
	for i, b := range list {
		items[i] = BusinessSummary{ID: b.ID, Name: b.Name, Avatar: b.Avatar, Color: b.Color, Category: b.Category,
			Type: b.Type, Rating: b.Rating, Location: b.Location, HoursText: b.HoursText}
	}
	return BusinessesCard{Type: "businesses", Items: items}
}

// ConciergeResult is what the chat shows after a customer's message: the stored messages in
// order (the customer's own first), plus the business the conversation is now routed to.
type ConciergeResult struct {
	Messages []Message `json:"messages"`
	Context  string    `json:"context"`
}

var nonLetters = regexp.MustCompile(`[^a-z0-9 ]+`)

// normalize lowercases and strips punctuation, so "Mama's" matches "mamas".
func normalize(s string) string {
	return strings.Join(strings.Fields(nonLetters.ReplaceAllString(strings.ToLower(strings.ReplaceAll(s, "'", "")), " ")), " ")
}

// Words too common to identify a business on their own.
var genericNameWords = map[string]bool{
	"with": true, "the": true, "and": true, "studio": true, "barbershop": true, "plumbing": true, "beauty": true,
	"cleaning": true, "kitchen": true, "maths": true, "salon": true, "shop": true, "restaurant": true, "cafe": true,
	"services": true, "service": true, "company": true, "store": true, "clinic": true, "centre": true, "center": true,
}

// detectBusiness finds a business named in a message ("book at mama's kitchen", "fresh cuts prices").
// It returns several businesses when the message is ambiguous.
func detectBusiness(businesses []Business, text string) (*Business, []Business) {
	norm := " " + normalize(text) + " "
	type hit struct {
		b     Business
		score int
	}
	var hits []hit
	for _, b := range businesses {
		name := normalize(b.Name)
		score := 0
		if strings.Contains(norm, " "+name+" ") {
			score = 100
		} else {
			for _, w := range strings.Fields(name) {
				if len(w) >= 4 && !genericNameWords[w] && strings.Contains(norm, " "+w+" ") {
					score++
				}
			}
		}
		if score > 0 {
			hits = append(hits, hit{b, score})
		}
	}
	if len(hits) == 0 {
		return nil, nil
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].score > hits[j].score })
	if len(hits) > 1 && hits[0].score == hits[1].score {
		var tied []Business
		for _, h := range hits {
			if h.score == hits[0].score {
				tied = append(tied, h.b)
			}
		}
		return nil, tied
	}
	return &hits[0].b, nil
}

// Needs mapped to words found in matching businesses' names, categories, descriptions and offerings.
var needSynonyms = map[string][]string{
	"haircut": {"barber", "hair", "fade"}, "hair": {"hair", "barber", "braid"}, "cut": {"barber", "cut"},
	"fade": {"fade", "barber"}, "beard": {"beard", "barber"}, "shave": {"barber", "beard"}, "barber": {"barber"},
	"hungry": {"@restaurant"}, "eat": {"@restaurant"}, "food": {"@restaurant"}, "lunch": {"@restaurant"},
	"dinner": {"@restaurant"}, "breakfast": {"@restaurant"}, "restaurant": {"@restaurant"}, "restaurants": {"@restaurant"},
	"meal": {"@restaurant"}, "table": {"@restaurant"},
	"leak": {"leak", "plumb"}, "leaking": {"leak", "plumb"}, "pipe": {"plumb"}, "geyser": {"geyser"}, "drain": {"drain"},
	"plumber": {"plumb"}, "toilet": {"plumb", "drain"}, "tap": {"plumb"},
	"nails": {"nail"}, "nail": {"nail"}, "manicure": {"nail", "manicure"}, "pedicure": {"pedicure"}, "braids": {"braid"},
	"wig": {"wig"}, "lashes": {"lash"}, "makeup": {"beauty"}, "beauty": {"beauty"},
	"maths": {"math"}, "math": {"math"}, "tutor": {"tutor"}, "lesson": {"lesson"}, "lessons": {"lesson"},
	"exam": {"math", "tutor"}, "homework": {"tutor"}, "study": {"tutor"},
	"clean": {"clean"}, "cleaning": {"clean"}, "cleaner": {"clean"}, "maid": {"clean"},
}

var searchStopwords = map[string]bool{
	"need": true, "want": true, "looking": true, "find": true, "please": true, "some": true, "where": true,
	"good": true, "best": true, "near": true, "nearby": true, "today": true, "tomorrow": true, "there": true,
	"have": true, "with": true, "from": true, "someone": true, "anyone": true, "recommend": true, "show": true,
	"business": true, "businesses": true, "help": true, "hello": true, "this": true, "that": true, "what": true,
}

// searchBusinesses ranks businesses for a need like "I need a haircut" or "I'm hungry".
func (s *Store) searchBusinesses(businesses []Business, text string) ([]Business, error) {
	var terms []string
	restaurant := false
	for _, w := range strings.Fields(normalize(text)) {
		if syn, ok := needSynonyms[w]; ok {
			for _, t := range syn {
				if t == "@restaurant" {
					restaurant = true
				} else {
					terms = append(terms, t)
				}
			}
		} else if len(w) >= 4 && !searchStopwords[w] {
			terms = append(terms, strings.TrimSuffix(w, "s"))
		}
	}
	if len(terms) == 0 && !restaurant {
		return nil, nil
	}

	type scored struct {
		b     Business
		score int
	}
	var results []scored
	for _, b := range businesses {
		offerings, err := s.ListOfferings(b.ID, true)
		if err != nil {
			return nil, err
		}
		hay := []string{b.Name, b.Category, b.About}
		for _, o := range offerings {
			hay = append(hay, o.Name, o.Category, o.Description)
		}
		haystack := normalize(strings.Join(hay, " "))
		score := 0
		if restaurant && b.Type == "restaurant" {
			score += 3
		}
		for _, t := range terms {
			if strings.Contains(haystack, t) {
				score++
			}
		}
		if score > 0 {
			results = append(results, scored{b, score})
		}
	}
	sort.SliceStable(results, func(i, j int) bool { return results[i].score > results[j].score })
	out := make([]Business, len(results))
	for i, r := range results {
		out[i] = r.b
	}
	return out, nil
}

// Phrases that always leave the current business and go back to the directory.
var exitPhrases = []string{"other business", "another business", "all businesses", "list businesses", "show businesses",
	"switch business", "change business", "start over", "back to business connect", "main menu", "directory"}

// Phrases that start a search for a different kind of business.
var searchPhrases = []string{"find a", "find me", "find another", "looking for", "recommend", "need a", "need an",
	"where can i get", "any good", "is there a"}

func wantsExit(norm string) bool {
	if norm == "back" || norm == "exit" || norm == "businesses" || norm == "leave" {
		return true
	}
	for _, p := range exitPhrases {
		if strings.Contains(norm, p) {
			return true
		}
	}
	return false
}

func wantsSearch(norm string) bool {
	for _, p := range searchPhrases {
		if strings.Contains(norm, p) {
			return true
		}
	}
	return false
}

// Words that, together with a business name, just mean "open this business".
var selectionFiller = map[string]bool{
	"hi": true, "hello": true, "hey": true, "talk": true, "to": true, "chat": true, "with": true, "open": true,
	"connect": true, "me": true, "please": true, "i": true, "want": true, "speak": true, "the": true, "message": true,
	"go": true, "let": true, "lets": true, "us": true, "at": true, "a": true, "can": true, "you": true,
}

func selectionOnly(norm string, b Business) bool {
	nameWords := map[string]bool{}
	for _, w := range strings.Fields(normalize(b.Name)) {
		nameWords[w] = true
	}
	for _, w := range strings.Fields(norm) {
		if !nameWords[w] && !selectionFiller[w] {
			return false
		}
	}
	return true
}

func (s *Store) conciergeSay(customerID string, r Reply) (Message, error) {
	return s.AddReply(conciergeID, customerID, r, true)
}

func conciergeWelcome() string {
	return "Hi 👋 I'm *Business Connect*. Tell me what you need, like *\"I need a haircut\"*, *\"I'm hungry\"* or *\"my geyser is leaking\"*, " +
		"or name a business to chat with it. Here's everyone on Business Connect 👇"
}

// Concierge handles one customer message. context is the business the customer was talking to, if any.
func (s *Store) Concierge(customerID, text, context string) (ConciergeResult, error) {
	businesses, err := s.ListBusinesses()
	if err != nil {
		return ConciergeResult{}, err
	}
	norm := normalize(text)
	target, ambiguous := detectBusiness(businesses, text)

	var searchResults []Business
	if target == nil && len(ambiguous) == 0 && context != "" && !wantsExit(norm) {
		current, err := s.GetBusiness(context)
		if err == nil && current.Type != "concierge" {
			target = &current
			// "I need a plumber" while talking to a restaurant goes back to searching,
			// unless the search points at the current business anyway.
			if wantsSearch(norm) {
				searchResults, err = s.searchBusinesses(businesses, text)
				if err != nil {
					return ConciergeResult{}, err
				}
				if len(searchResults) > 0 && searchResults[0].ID != current.ID {
					target = nil
				}
			}
		}
	}

	if target == nil {
		return s.conciergeReply(businesses, customerID, text, norm, ambiguous, searchResults)
	}
	return s.routeToBusiness(*target, customerID, text, norm, context)
}

func (s *Store) conciergeReply(businesses []Business, customerID, text, norm string, ambiguous, results []Business) (ConciergeResult, error) {
	client := Message{BusinessID: conciergeID, CustomerID: customerID, From: "client", Text: text, Status: "read"}
	if err := s.AddMessage(&client); err != nil {
		return ConciergeResult{}, err
	}
	out := []Message{client}
	say := func(r Reply) error {
		m, err := s.conciergeSay(customerID, r)
		out = append(out, m)
		return err
	}

	var err error
	switch {
	case len(ambiguous) > 0:
		err = say(Reply{Text: "Which business did you mean? Tap *Chat* to talk to one 👇", Card: businessesCard(ambiguous)})
	case strings.Contains(norm, "my booking") || strings.Contains(norm, "my reservation") || strings.Contains(norm, "my appointment"):
		res, qerr := s.UpcomingCustomerReservations(customerID, today().Format(dateLayout))
		if qerr != nil {
			return ConciergeResult{}, qerr
		}
		if len(res) == 0 {
			err = say(Reply{Text: "You don't have any upcoming bookings yet. Tell me what you need and I'll find a business for you."})
		} else {
			err = say(Reply{Text: "📅 Your upcoming bookings:", Card: BookingsCard{Type: "bookings", Title: "Your bookings", Reservations: res}})
		}
	default:
		if results == nil && !wantsExit(norm) {
			if results, err = s.searchBusinesses(businesses, text); err != nil {
				return ConciergeResult{}, err
			}
		}
		switch {
		case len(results) > 0:
			err = say(Reply{Text: "Here are businesses that can help 👇 Tap *Chat*, or just mention one by name.", Card: businessesCard(results)})
		case wantsExit(norm):
			err = say(Reply{Text: "Sure! Who would you like to chat with next? 👇", Card: businessesCard(businesses)})
		case containsAny(norm, "hi", "hello", "hey", "help", "morning", "afternoon", "evening", "start") || norm == "":
			err = say(Reply{Text: conciergeWelcome(), Card: businessesCard(businesses)})
		default:
			err = say(Reply{Text: "I couldn't find a business for that. Here's everyone on Business Connect 👇", Card: businessesCard(businesses)})
		}
	}
	if err != nil {
		return ConciergeResult{}, err
	}
	return ConciergeResult{Messages: out, Context: ""}, nil
}

func (s *Store) routeToBusiness(b Business, customerID, text, norm, context string) (ConciergeResult, error) {
	switched := b.ID != context
	firstContact := !s.HasBusinessReply(b.ID, customerID)

	client := Message{BusinessID: b.ID, CustomerID: customerID, From: "client", Text: text, Status: "sent"}
	if err := s.AddMessage(&client); err != nil {
		return ConciergeResult{}, err
	}
	out := []Message{client}

	if switched {
		note := fmt.Sprintf("🔀 You're now chatting with %s *%s*. Say *back* any time to find another business.", b.Avatar, b.Name)
		if !b.AutoReply.Enabled {
			note += " They reply personally, so their answer will appear here."
		}
		m, err := s.conciergeSay(customerID, Reply{Text: note})
		if err != nil {
			return ConciergeResult{}, err
		}
		out = append(out, m)
	}

	var replies []Reply
	switch {
	case !b.AutoReply.Enabled:
	case selectionOnly(norm, b):
		if firstContact && strings.TrimSpace(b.AutoReply.Greeting) != "" {
			replies = append(replies, Reply{Text: b.AutoReply.Greeting})
		}
		replies = append(replies, Reply{Text: customerHelp(b)})
	default:
		var err error
		if replies, err = s.CustomerReplies(b, customerID, text, firstContact); err != nil {
			return ConciergeResult{}, err
		}
	}

	if len(replies) > 0 {
		s.MarkClientMessagesRead(b.ID, customerID)
		out[0].Status = "read"
	}
	for _, r := range replies {
		m, err := s.AddReply(b.ID, customerID, r, true)
		if err != nil {
			return ConciergeResult{}, err
		}
		out = append(out, m)
	}
	return ConciergeResult{Messages: out, Context: b.ID}, nil
}

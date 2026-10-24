package main

import (
	"strings"
	"testing"
)

func TestConciergeRouting(t *testing.T) {
	s := setup(t)
	if err := s.EnsureConcierge(); err != nil {
		t.Fatal(err)
	}
	s.UpsertCustomer("c1", "Thandi", "")

	// A need with no business named: the concierge suggests businesses.
	res, err := s.Concierge("c1", "I need a haircut", "")
	if err != nil {
		t.Fatal(err)
	}
	last := res.Messages[len(res.Messages)-1]
	if res.Context != "" || last.BusinessID != conciergeID || !strings.Contains(string(last.Card), "fresh-cuts") {
		t.Fatalf("haircut search = %+v", res)
	}

	// Naming a business routes to it and replies with its greeting and help.
	res, _ = s.Concierge("c1", "Mama's Kitchen", "")
	if res.Context != "mamas-kitchen" || res.Messages[0].BusinessID != "mamas-kitchen" || len(res.Messages) != 4 {
		t.Fatalf("select = %+v", res)
	}

	// Follow-ups stay with that business.
	res, _ = s.Concierge("c1", "menu please", "mamas-kitchen")
	if res.Context != "mamas-kitchen" || !strings.Contains(string(res.Messages[1].Card), "Oxtail") {
		t.Fatalf("menu = %+v", res)
	}

	// Mentioning another business switches, and the question is answered by that business.
	res, _ = s.Concierge("c1", "how much is a fade at fresh cuts?", "mamas-kitchen")
	if res.Context != "fresh-cuts" || !strings.Contains(res.Messages[len(res.Messages)-1].Text, "Fade – $6") {
		t.Fatalf("switch = %+v", res)
	}

	// Searching for a different kind of business leaves the current one.
	res, _ = s.Concierge("c1", "I need a plumber", "fresh-cuts")
	if res.Context != "" || !strings.Contains(string(res.Messages[1].Card), "tendai-plumbing") {
		t.Fatalf("search from context = %+v", res)
	}

	// "back" returns to the directory.
	res, _ = s.Concierge("c1", "back", "fresh-cuts")
	if res.Context != "" {
		t.Fatalf("back = %+v", res)
	}

	// The owner only sees messages routed to their business.
	msgs, _ := s.Messages("mamas-kitchen", "c1", 0)
	for _, m := range msgs {
		if strings.Contains(m.Text, "plumber") || strings.Contains(m.Text, "haircut") {
			t.Fatalf("owner saw a message meant for someone else: %q", m.Text)
		}
	}
	thread, _ := s.CustomerThread("c1", 0)
	if len(thread) < 10 {
		t.Fatalf("thread has %d messages", len(thread))
	}
	if list, _ := s.ListBusinesses(); len(list) != 6 {
		t.Fatalf("concierge should be hidden from the business list, got %d", len(list))
	}
}

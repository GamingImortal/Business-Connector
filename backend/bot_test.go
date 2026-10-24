package main

import (
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Fixed "now": Friday 23 Oct 2026, 10:00.
func setup(t *testing.T) *Store {
	t.Helper()
	fixed := time.Date(2026, 10, 23, 10, 0, 0, 0, time.Local)
	clock = func() time.Time { return fixed }
	t.Cleanup(func() { clock = time.Now })

	s, err := openStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Seed(); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestParsing(t *testing.T) {
	setup(t)
	if tm, ok := findTime("book 11:30 please"); !ok || tm != "11:30" {
		t.Errorf("findTime 11:30 = %q %v", tm, ok)
	}
	if tm, _ := findTime("at 7pm"); tm != "19:00" {
		t.Errorf("findTime 7pm = %q", tm)
	}
	if n, ok := findPartySize("table for 4 at 19:00"); !ok || n != 4 {
		t.Errorf("party = %d %v", n, ok)
	}
	if _, ok := findPartySize("book for 11:30"); ok {
		t.Error("a time should not be read as a party size")
	}
	if d, _ := findDate("saturday"); d.Format(dateLayout) != "2026-10-24" {
		t.Errorf("saturday = %s", d.Format(dateLayout))
	}
	if d, _ := findDate("tomorrow"); d.Format(dateLayout) != "2026-10-24" {
		t.Errorf("tomorrow = %s", d.Format(dateLayout))
	}
}

func TestSlotsRespectCapacity(t *testing.T) {
	s := setup(t)
	b, _ := s.GetBusiness("mamas-kitchen") // 40 seats, 90 min slots, opens 11:00
	date := today()
	res := []Reservation{{Time: "11:00", PartySize: 38, Status: "confirmed"}}
	slots := Slots(b, date, res, 4)
	if slots[0].Time != "11:00" || slots[0].Available || slots[0].Remaining != 2 {
		t.Errorf("first slot = %+v, want 2 seats left and unavailable for 4", slots[0])
	}
	if !slots[1].Available {
		t.Errorf("second slot should be free: %+v", slots[1])
	}
}

func TestCustomerBooksFromText(t *testing.T) {
	s := setup(t)
	b, _ := s.GetBusiness("mamas-kitchen")
	s.UpsertCustomer("c1", "Thandi", "")

	replies, err := s.CustomerReplies(b, "c1", "Can I book a table for 4 on Saturday at 18:30?", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(replies) != 1 || !strings.Contains(replies[0].Text, "Table for 4 reserved") {
		t.Fatalf("replies = %+v", replies)
	}
	mine, _ := s.CustomerReservations(b.ID, "c1", "2026-10-23")
	if len(mine) != 1 || mine[0].Date != "2026-10-24" || mine[0].Time != "18:30" || mine[0].PartySize != 4 {
		t.Fatalf("reservations = %+v", mine)
	}
}

func TestCustomerMenuAndFallback(t *testing.T) {
	s := setup(t)
	b, _ := s.GetBusiness("fresh-cuts")

	replies, _ := s.CustomerReplies(b, "c2", "how much is a beard trim?", true)
	if len(replies) != 2 || !strings.Contains(replies[1].Text, "Beard trim – $3") {
		t.Fatalf("replies = %+v", replies)
	}
	mk, _ := s.GetBusiness("mamas-kitchen")
	s.CreateOffering(&Offering{BusinessID: mk.ID, Kind: "menu", Name: "Chicken Stew", Price: 7, Available: true})
	replies, _ = s.CustomerReplies(mk, "c2", "how much is the chicken stew?", false)
	if len(replies) != 1 || replies[0].Text != "• Chicken Stew – $7" {
		t.Fatalf("exact match replies = %+v", replies)
	}

	replies, _ = s.CustomerReplies(b, "c2", "blah", false)
	if len(replies) != 1 || replies[0].Text != b.AutoReply.Fallback {
		t.Fatalf("fallback replies = %+v", replies)
	}

	off, _ := s.GetBusiness("sparkle-clean") // auto-reply disabled
	if replies, _ := s.CustomerReplies(off, "c2", "price", true); len(replies) != 0 {
		t.Fatalf("disabled auto-reply should not answer: %+v", replies)
	}
}

func TestOwnerAssistant(t *testing.T) {
	s := setup(t)
	b, _ := s.GetBusiness("mamas-kitchen")

	replies, _ := s.OwnerReplies(b, "today")
	if !strings.Contains(replies[0].Text, "2 booking(s), 12 guests") {
		t.Fatalf("today = %+v", replies)
	}
	res, _ := s.ListReservations(b.ID, "2026-10-23", "2026-10-23")
	replies, _ = s.OwnerReplies(b, "cancel "+strconv.FormatInt(res[0].ID, 10))
	if !strings.Contains(replies[0].Text, "cancelled") {
		t.Fatalf("cancel = %+v", replies)
	}
	replies, _ = s.OwnerReplies(b, "sold out oxtail")
	if !strings.Contains(replies[0].Text, "Oxtail is now hidden") {
		t.Fatalf("sold out = %+v", replies)
	}
	available, _ := s.ListOfferings(b.ID, true)
	for _, o := range available {
		if o.Name == "Oxtail" {
			t.Fatal("oxtail should be hidden from customers")
		}
	}
}

func TestBookRejectsFullSlot(t *testing.T) {
	s := setup(t)
	b, _ := s.GetBusiness("tendai-plumbing") // capacity 1
	r1 := Reservation{CustomerName: "A", Date: "2026-10-24", Time: "11:00", Source: "chat"}
	if err := s.Book(b, &r1); err != nil {
		t.Fatal(err)
	}
	r2 := Reservation{CustomerName: "B", Date: "2026-10-24", Time: "11:00", Source: "chat"}
	if err := s.Book(b, &r2); err != errSlotTaken {
		t.Fatalf("err = %v, want errSlotTaken", err)
	}
	r3 := Reservation{CustomerName: "C", Date: "2026-10-23", Time: "09:00", Source: "chat"}
	if err := s.Book(b, &r3); err != errPast {
		t.Fatalf("err = %v, want errPast", err)
	}
}

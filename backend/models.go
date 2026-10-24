package main

import "encoding/json"

// AutoReply is a business's automated-response setup. When Enabled is false,
// all replies (including the smart answers built from listed data) are manual.
type AutoReply struct {
	Enabled  bool   `json:"enabled"`
	Greeting string `json:"greeting"`
	Fallback string `json:"fallback"`
	Rules    []Rule `json:"rules"`
}

// Rule matches when any of its comma-separated keywords appears in a customer's message.
type Rule struct {
	ID       string `json:"id"`
	Keywords string `json:"keywords"`
	Reply    string `json:"reply"`
}

// DayHours holds the opening hours for one weekday (0 = Sunday).
type DayHours struct {
	Weekday int    `json:"weekday"`
	Open    string `json:"open"`  // "08:00"
	Close   string `json:"close"` // "18:00"
	Closed  bool   `json:"closed"`
}

type Business struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Category    string     `json:"category"`
	Type        string     `json:"type"` // "service" | "restaurant"
	Avatar      string     `json:"avatar"`
	Color       string     `json:"color"`
	Rating      float64    `json:"rating"`
	Location    string     `json:"location"`
	Phone       string     `json:"phone"`
	About       string     `json:"about"`
	Hours       []DayHours `json:"hours"`
	HoursText   string     `json:"hoursText"`
	SlotMinutes int        `json:"slotMinutes"`
	// Capacity is how many bookings can share a time slot. For restaurants it is
	// counted in guests (seats); for service businesses in appointments.
	Capacity  int       `json:"capacity"`
	AutoReply AutoReply `json:"autoReply"`
}

type Offering struct {
	ID           int64   `json:"id"`
	BusinessID   string  `json:"businessId"`
	Kind         string  `json:"kind"` // "service" | "product" | "menu"
	Category     string  `json:"category"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Price        float64 `json:"price"`
	DurationMins int     `json:"durationMins"`
	Available    bool    `json:"available"`
}

type Reservation struct {
	ID            int64  `json:"id"`
	BusinessID    string `json:"businessId"`
	CustomerID    string `json:"customerId"`
	CustomerName  string `json:"customerName"`
	CustomerPhone string `json:"customerPhone"`
	OfferingID    *int64 `json:"offeringId"`
	OfferingName  string `json:"offeringName"`
	Date          string `json:"date"` // "2026-10-26"
	Time          string `json:"time"` // "11:30"
	PartySize     int    `json:"partySize"`
	Status        string `json:"status"` // "confirmed" | "completed" | "cancelled" | "no_show"
	Notes         string `json:"notes"`
	Source        string `json:"source"` // "chat" | "dashboard"
	CreatedAt     string `json:"createdAt"`
}

type Message struct {
	ID         int64           `json:"id"`
	BusinessID string          `json:"businessId"`
	CustomerID string          `json:"customerId"`
	From       string          `json:"from"` // "client" | "business"
	Text       string          `json:"text"`
	Auto       bool            `json:"auto"`
	Card       json.RawMessage `json:"card,omitempty"`
	Status     string          `json:"status,omitempty"` // client messages: "sent" | "read"
	CreatedAt  string          `json:"createdAt"`
}

// Reply is a bot answer that has not been stored yet.
type Reply struct {
	Text string `json:"text"`
	Card any    `json:"card,omitempty"`
}

type Slot struct {
	Time      string `json:"time"`
	Remaining int    `json:"remaining"`
	Available bool   `json:"available"`
}

var weekdayNames = []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}

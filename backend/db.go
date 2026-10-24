package main

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS businesses (
	id           TEXT PRIMARY KEY,
	name         TEXT NOT NULL,
	category     TEXT NOT NULL DEFAULT '',
	type         TEXT NOT NULL DEFAULT 'service',
	avatar       TEXT NOT NULL DEFAULT '🏪',
	color        TEXT NOT NULL DEFAULT '#00a884',
	rating       REAL NOT NULL DEFAULT 0,
	location     TEXT NOT NULL DEFAULT '',
	phone        TEXT NOT NULL DEFAULT '',
	about        TEXT NOT NULL DEFAULT '',
	hours        TEXT NOT NULL DEFAULT '[]',
	slot_minutes INTEGER NOT NULL DEFAULT 60,
	capacity     INTEGER NOT NULL DEFAULT 1,
	auto_reply   TEXT NOT NULL DEFAULT '{}',
	pin_salt     TEXT NOT NULL,
	pin_hash     TEXT NOT NULL,
	created_at   TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS offerings (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	business_id   TEXT NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
	kind          TEXT NOT NULL DEFAULT 'service',
	category      TEXT NOT NULL DEFAULT '',
	name          TEXT NOT NULL,
	description   TEXT NOT NULL DEFAULT '',
	price         REAL NOT NULL DEFAULT 0,
	duration_mins INTEGER NOT NULL DEFAULT 0,
	available     INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS customers (
	id         TEXT PRIMARY KEY,
	name       TEXT NOT NULL DEFAULT '',
	phone      TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS reservations (
	id             INTEGER PRIMARY KEY AUTOINCREMENT,
	business_id    TEXT NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
	customer_id    TEXT NOT NULL DEFAULT '',
	customer_name  TEXT NOT NULL DEFAULT '',
	customer_phone TEXT NOT NULL DEFAULT '',
	offering_id    INTEGER REFERENCES offerings(id) ON DELETE SET NULL,
	date           TEXT NOT NULL,
	time           TEXT NOT NULL,
	party_size     INTEGER NOT NULL DEFAULT 1,
	status         TEXT NOT NULL DEFAULT 'confirmed',
	notes          TEXT NOT NULL DEFAULT '',
	source         TEXT NOT NULL DEFAULT 'chat',
	created_at     TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_reservations_business_date ON reservations(business_id, date);

CREATE TABLE IF NOT EXISTS messages (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	business_id TEXT NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
	customer_id TEXT NOT NULL,
	sender      TEXT NOT NULL,
	text        TEXT NOT NULL,
	auto        INTEGER NOT NULL DEFAULT 0,
	card        TEXT,
	status      TEXT NOT NULL DEFAULT '',
	created_at  TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_messages_conversation ON messages(business_id, customer_id, id);

CREATE TABLE IF NOT EXISTS sessions (
	token       TEXT PRIMARY KEY,
	business_id TEXT NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
	expires_at  TEXT NOT NULL
);
`

var errNotFound = errors.New("not found")

type Store struct {
	db *sql.DB
}

func openStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}
	// SQLite allows one writer at a time; a single connection avoids "database is locked" errors.
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &Store{db: db}, nil
}

func now() string { return time.Now().Format(time.RFC3339) }

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func hashPIN(salt, pin string) string {
	sum := sha256.Sum256([]byte(salt + ":" + pin))
	return hex.EncodeToString(sum[:])
}

// ---- Businesses ----

const businessColumns = `id, name, category, type, avatar, color, rating, location, phone, about, hours, slot_minutes, capacity, auto_reply`

func scanBusiness(row interface{ Scan(...any) error }) (Business, error) {
	var b Business
	var hours, autoReply string
	err := row.Scan(&b.ID, &b.Name, &b.Category, &b.Type, &b.Avatar, &b.Color, &b.Rating,
		&b.Location, &b.Phone, &b.About, &hours, &b.SlotMinutes, &b.Capacity, &autoReply)
	if errors.Is(err, sql.ErrNoRows) {
		return b, errNotFound
	}
	if err != nil {
		return b, err
	}
	_ = json.Unmarshal([]byte(hours), &b.Hours)
	_ = json.Unmarshal([]byte(autoReply), &b.AutoReply)
	if b.AutoReply.Rules == nil {
		b.AutoReply.Rules = []Rule{}
	}
	b.Hours = normalizeHours(b.Hours)
	b.HoursText = hoursSummary(b.Hours)
	return b, nil
}

func (s *Store) ListBusinesses() ([]Business, error) {
	rows, err := s.db.Query(`SELECT ` + businessColumns + ` FROM businesses WHERE type <> 'concierge' ORDER BY created_at, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []Business{}
	for rows.Next() {
		b, err := scanBusiness(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, b)
	}
	return list, rows.Err()
}

func (s *Store) GetBusiness(id string) (Business, error) {
	return scanBusiness(s.db.QueryRow(`SELECT `+businessColumns+` FROM businesses WHERE id = ?`, id))
}

func (s *Store) CreateBusiness(b Business, pin string) error {
	hours, _ := json.Marshal(normalizeHours(b.Hours))
	autoReply, _ := json.Marshal(b.AutoReply)
	salt := randomHex(16)
	_, err := s.db.Exec(`INSERT INTO businesses (`+businessColumns+`, pin_salt, pin_hash, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		b.ID, b.Name, b.Category, b.Type, b.Avatar, b.Color, b.Rating, b.Location, b.Phone, b.About,
		string(hours), b.SlotMinutes, b.Capacity, string(autoReply), salt, hashPIN(salt, pin), now())
	return err
}

func (s *Store) UpdateBusiness(b Business) error {
	hours, _ := json.Marshal(normalizeHours(b.Hours))
	autoReply, _ := json.Marshal(b.AutoReply)
	_, err := s.db.Exec(`UPDATE businesses SET name = ?, category = ?, type = ?, avatar = ?, color = ?,
		location = ?, phone = ?, about = ?, hours = ?, slot_minutes = ?, capacity = ?, auto_reply = ?
		WHERE id = ?`,
		b.Name, b.Category, b.Type, b.Avatar, b.Color, b.Location, b.Phone, b.About,
		string(hours), b.SlotMinutes, b.Capacity, string(autoReply), b.ID)
	return err
}

func (s *Store) CheckPIN(businessID, pin string) bool {
	var salt, hash string
	err := s.db.QueryRow(`SELECT pin_salt, pin_hash FROM businesses WHERE id = ? AND type <> 'concierge'`, businessID).Scan(&salt, &hash)
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(hashPIN(salt, pin)), []byte(hash)) == 1
}

func (s *Store) SetPIN(businessID, pin string) error {
	salt := randomHex(16)
	_, err := s.db.Exec(`UPDATE businesses SET pin_salt = ?, pin_hash = ? WHERE id = ?`, salt, hashPIN(salt, pin), businessID)
	return err
}

// ---- Sessions ----

const sessionTTL = 7 * 24 * time.Hour

func (s *Store) CreateSession(businessID string) (string, error) {
	token := randomHex(24)
	_, err := s.db.Exec(`INSERT INTO sessions (token, business_id, expires_at) VALUES (?, ?, ?)`,
		token, businessID, time.Now().Add(sessionTTL).Format(time.RFC3339))
	return token, err
}

func (s *Store) SessionBusiness(token string) (string, error) {
	var businessID, expires string
	err := s.db.QueryRow(`SELECT business_id, expires_at FROM sessions WHERE token = ?`, token).Scan(&businessID, &expires)
	if err != nil {
		return "", errNotFound
	}
	if t, err := time.Parse(time.RFC3339, expires); err != nil || time.Now().After(t) {
		s.db.Exec(`DELETE FROM sessions WHERE token = ?`, token)
		return "", errNotFound
	}
	return businessID, nil
}

func (s *Store) DeleteSession(token string) {
	s.db.Exec(`DELETE FROM sessions WHERE token = ?`, token)
}

// ---- Offerings ----

func (s *Store) ListOfferings(businessID string, onlyAvailable bool) ([]Offering, error) {
	q := `SELECT id, business_id, kind, category, name, description, price, duration_mins, available
		FROM offerings WHERE business_id = ?`
	if onlyAvailable {
		q += ` AND available = 1`
	}
	q += ` ORDER BY CASE kind WHEN 'menu' THEN 0 WHEN 'service' THEN 1 ELSE 2 END, category, id`
	rows, err := s.db.Query(q, businessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []Offering{}
	for rows.Next() {
		var o Offering
		if err := rows.Scan(&o.ID, &o.BusinessID, &o.Kind, &o.Category, &o.Name, &o.Description,
			&o.Price, &o.DurationMins, &o.Available); err != nil {
			return nil, err
		}
		list = append(list, o)
	}
	return list, rows.Err()
}

func (s *Store) CreateOffering(o *Offering) error {
	res, err := s.db.Exec(`INSERT INTO offerings (business_id, kind, category, name, description, price, duration_mins, available)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		o.BusinessID, o.Kind, o.Category, o.Name, o.Description, o.Price, o.DurationMins, o.Available)
	if err != nil {
		return err
	}
	o.ID, _ = res.LastInsertId()
	return nil
}

func (s *Store) UpdateOffering(o Offering) error {
	res, err := s.db.Exec(`UPDATE offerings SET kind = ?, category = ?, name = ?, description = ?, price = ?,
		duration_mins = ?, available = ? WHERE id = ? AND business_id = ?`,
		o.Kind, o.Category, o.Name, o.Description, o.Price, o.DurationMins, o.Available, o.ID, o.BusinessID)
	return mustAffect(res, err)
}

func (s *Store) DeleteOffering(businessID string, id int64) error {
	res, err := s.db.Exec(`DELETE FROM offerings WHERE id = ? AND business_id = ?`, id, businessID)
	return mustAffect(res, err)
}

func mustAffect(res sql.Result, err error) error {
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errNotFound
	}
	return nil
}

// ---- Customers ----

func (s *Store) UpsertCustomer(id, name, phone string) error {
	_, err := s.db.Exec(`INSERT INTO customers (id, name, phone, created_at) VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = CASE WHEN excluded.name <> '' THEN excluded.name ELSE customers.name END,
			phone = CASE WHEN excluded.phone <> '' THEN excluded.phone ELSE customers.phone END`,
		id, name, phone, now())
	return err
}

// ---- Reservations ----

const reservationSelect = `SELECT r.id, r.business_id, r.customer_id, r.customer_name, r.customer_phone,
	r.offering_id, COALESCE(o.name, ''), r.date, r.time, r.party_size, r.status, r.notes, r.source, r.created_at
	FROM reservations r LEFT JOIN offerings o ON o.id = r.offering_id`

func (s *Store) queryReservations(where string, args ...any) ([]Reservation, error) {
	rows, err := s.db.Query(reservationSelect+" WHERE "+where+" ORDER BY r.date, r.time, r.id", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []Reservation{}
	for rows.Next() {
		var r Reservation
		var offeringID sql.NullInt64
		if err := rows.Scan(&r.ID, &r.BusinessID, &r.CustomerID, &r.CustomerName, &r.CustomerPhone,
			&offeringID, &r.OfferingName, &r.Date, &r.Time, &r.PartySize, &r.Status, &r.Notes,
			&r.Source, &r.CreatedAt); err != nil {
			return nil, err
		}
		if offeringID.Valid {
			r.OfferingID = &offeringID.Int64
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

// ListReservations returns a business's reservations between two dates (inclusive). Empty bounds are open.
func (s *Store) ListReservations(businessID, from, to string) ([]Reservation, error) {
	where := []string{"r.business_id = ?"}
	args := []any{businessID}
	if from != "" {
		where = append(where, "r.date >= ?")
		args = append(args, from)
	}
	if to != "" {
		where = append(where, "r.date <= ?")
		args = append(args, to)
	}
	return s.queryReservations(strings.Join(where, " AND "), args...)
}

func (s *Store) CustomerReservations(businessID, customerID, fromDate string) ([]Reservation, error) {
	return s.queryReservations(`r.business_id = ? AND r.customer_id = ? AND r.date >= ? AND r.status = 'confirmed'`,
		businessID, customerID, fromDate)
}

func (s *Store) GetReservation(businessID string, id int64) (Reservation, error) {
	list, err := s.queryReservations(`r.business_id = ? AND r.id = ?`, businessID, id)
	if err != nil {
		return Reservation{}, err
	}
	if len(list) == 0 {
		return Reservation{}, errNotFound
	}
	return list[0], nil
}

func (s *Store) CreateReservation(r *Reservation) error {
	r.CreatedAt = now()
	res, err := s.db.Exec(`INSERT INTO reservations (business_id, customer_id, customer_name, customer_phone,
		offering_id, date, time, party_size, status, notes, source, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.BusinessID, r.CustomerID, r.CustomerName, r.CustomerPhone, r.OfferingID, r.Date, r.Time,
		r.PartySize, r.Status, r.Notes, r.Source, r.CreatedAt)
	if err != nil {
		return err
	}
	r.ID, _ = res.LastInsertId()
	return nil
}

func (s *Store) UpdateReservation(r Reservation) error {
	res, err := s.db.Exec(`UPDATE reservations SET customer_name = ?, customer_phone = ?, offering_id = ?,
		date = ?, time = ?, party_size = ?, status = ?, notes = ? WHERE id = ? AND business_id = ?`,
		r.CustomerName, r.CustomerPhone, r.OfferingID, r.Date, r.Time, r.PartySize, r.Status, r.Notes,
		r.ID, r.BusinessID)
	return mustAffect(res, err)
}

func (s *Store) DeleteReservation(businessID string, id int64) error {
	res, err := s.db.Exec(`DELETE FROM reservations WHERE id = ? AND business_id = ?`, id, businessID)
	return mustAffect(res, err)
}

// ---- Messages ----

func (s *Store) AddMessage(m *Message) error {
	m.CreatedAt = now()
	var card any
	if len(m.Card) > 0 {
		card = string(m.Card)
	}
	res, err := s.db.Exec(`INSERT INTO messages (business_id, customer_id, sender, text, auto, card, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		m.BusinessID, m.CustomerID, m.From, m.Text, m.Auto, card, m.Status, m.CreatedAt)
	if err != nil {
		return err
	}
	m.ID, _ = res.LastInsertId()
	return nil
}

// AddReply stores a bot or owner reply in a conversation.
func (s *Store) AddReply(businessID, customerID string, r Reply, auto bool) (Message, error) {
	m := Message{BusinessID: businessID, CustomerID: customerID, From: "business", Text: r.Text, Auto: auto}
	if r.Card != nil {
		m.Card, _ = json.Marshal(r.Card)
	}
	err := s.AddMessage(&m)
	return m, err
}

func (s *Store) Messages(businessID, customerID string, afterID int64) ([]Message, error) {
	return s.queryMessages(`business_id = ? AND customer_id = ? AND id > ?`, businessID, customerID, afterID)
}

// CustomerThread is everything a customer has said and been told, across all businesses
// and the Business Connect concierge, in one timeline.
func (s *Store) CustomerThread(customerID string, afterID int64) ([]Message, error) {
	return s.queryMessages(`customer_id = ? AND id > ?`, customerID, afterID)
}

func (s *Store) queryMessages(where string, args ...any) ([]Message, error) {
	rows, err := s.db.Query(`SELECT id, business_id, customer_id, sender, text, auto, card, status, created_at
		FROM messages WHERE `+where+` ORDER BY id`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []Message{}
	for rows.Next() {
		var m Message
		var card sql.NullString
		if err := rows.Scan(&m.ID, &m.BusinessID, &m.CustomerID, &m.From, &m.Text, &m.Auto, &card,
			&m.Status, &m.CreatedAt); err != nil {
			return nil, err
		}
		if card.Valid {
			m.Card = json.RawMessage(card.String)
		}
		list = append(list, m)
	}
	return list, rows.Err()
}

func (s *Store) UpcomingCustomerReservations(customerID, fromDate string) ([]Reservation, error) {
	return s.queryReservations(`r.customer_id = ? AND r.date >= ? AND r.status = 'confirmed'`, customerID, fromDate)
}

func (s *Store) HasBusinessReply(businessID, customerID string) bool {
	var n int
	s.db.QueryRow(`SELECT COUNT(*) FROM messages WHERE business_id = ? AND customer_id = ? AND sender = 'business'`,
		businessID, customerID).Scan(&n)
	return n > 0
}

func (s *Store) MarkClientMessagesRead(businessID, customerID string) {
	s.db.Exec(`UPDATE messages SET status = 'read' WHERE business_id = ? AND customer_id = ? AND sender = 'client' AND status <> 'read'`,
		businessID, customerID)
}

// Conversation is a chat-list entry: the latest message plus counters.
type Conversation struct {
	BusinessID           string  `json:"businessId"`
	CustomerID           string  `json:"customerId"`
	CustomerName         string  `json:"customerName"`
	LastMessage          Message `json:"lastMessage"`
	Unread               int     `json:"unread"`               // owner side: client messages not yet read
	BusinessMessageCount int     `json:"businessMessageCount"` // customer side: used to work out unread
}

func (s *Store) conversations(where string, arg string) ([]Conversation, error) {
	rows, err := s.db.Query(`
		SELECT m.business_id, m.customer_id, COALESCE(c.name, ''), m.id, m.sender, m.text, m.auto, m.status, m.created_at,
			(SELECT COUNT(*) FROM messages u WHERE u.business_id = m.business_id AND u.customer_id = m.customer_id
				AND u.sender = 'client' AND u.status <> 'read'),
			(SELECT COUNT(*) FROM messages u WHERE u.business_id = m.business_id AND u.customer_id = m.customer_id
				AND u.sender = 'business')
		FROM messages m
		LEFT JOIN customers c ON c.id = m.customer_id
		WHERE m.id IN (SELECT MAX(id) FROM messages WHERE `+where+` GROUP BY business_id, customer_id)
		ORDER BY m.id DESC`, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []Conversation{}
	for rows.Next() {
		var c Conversation
		m := &c.LastMessage
		if err := rows.Scan(&c.BusinessID, &c.CustomerID, &c.CustomerName, &m.ID, &m.From, &m.Text, &m.Auto,
			&m.Status, &m.CreatedAt, &c.Unread, &c.BusinessMessageCount); err != nil {
			return nil, err
		}
		m.BusinessID, m.CustomerID = c.BusinessID, c.CustomerID
		list = append(list, c)
	}
	return list, rows.Err()
}

func (s *Store) BusinessConversations(businessID string) ([]Conversation, error) {
	return s.conversations("business_id = ?", businessID)
}

func (s *Store) CustomerConversations(customerID string) ([]Conversation, error) {
	return s.conversations("customer_id = ?", customerID)
}

func (s *Store) CustomerName(id string) string {
	var name string
	s.db.QueryRow(`SELECT name FROM customers WHERE id = ?`, id).Scan(&name)
	return name
}

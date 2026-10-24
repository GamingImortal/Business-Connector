package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Server struct {
	store   *Store
	limiter *pinLimiter
	origins map[string]bool
}

func NewServer(store *Store, allowedOrigins []string) *Server {
	s := &Server{store: store, limiter: newPinLimiter(), origins: map[string]bool{}}
	for _, o := range allowedOrigins {
		s.origins[o] = true
	}
	return s
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	// Public: used by the chat platform's customer view.
	mux.HandleFunc("GET /api/businesses", s.listBusinesses)
	mux.HandleFunc("POST /api/businesses", s.registerBusiness)
	mux.HandleFunc("GET /api/businesses/{id}", s.getBusiness)
	mux.HandleFunc("GET /api/businesses/{id}/availability", s.availability)
	mux.HandleFunc("GET /api/businesses/{id}/messages", s.customerMessages)
	mux.HandleFunc("POST /api/businesses/{id}/messages", s.customerSend)
	mux.HandleFunc("POST /api/businesses/{id}/reservations", s.customerBook)
	mux.HandleFunc("POST /api/businesses/{id}/reservations/{rid}/cancel", s.customerCancel)
	mux.HandleFunc("GET /api/customers/{cid}/conversations", s.customerConversations)

	// The single Business Connect conversation that routes customers to businesses.
	mux.HandleFunc("POST /api/concierge", s.conciergeSend)
	mux.HandleFunc("GET /api/customers/{cid}/thread", s.customerThread)

	// Owner auth: a business ID plus PIN returns a bearer token.
	mux.HandleFunc("POST /api/auth/login", s.login)
	mux.HandleFunc("POST /api/auth/logout", s.logout)

	// Owner: used by the dashboard and the chat platform's business view.
	owner := func(pattern string, h func(http.ResponseWriter, *http.Request, Business)) {
		mux.HandleFunc(pattern, s.requireOwner(h))
	}
	owner("GET /api/owner/business", s.ownerGetBusiness)
	owner("PUT /api/owner/business", s.ownerUpdateBusiness)
	owner("PUT /api/owner/pin", s.ownerChangePIN)
	owner("GET /api/owner/offerings", s.ownerListOfferings)
	owner("POST /api/owner/offerings", s.ownerCreateOffering)
	owner("PUT /api/owner/offerings/{oid}", s.ownerUpdateOffering)
	owner("DELETE /api/owner/offerings/{oid}", s.ownerDeleteOffering)
	owner("GET /api/owner/reservations", s.ownerListReservations)
	owner("POST /api/owner/reservations", s.ownerCreateReservation)
	owner("PUT /api/owner/reservations/{rid}", s.ownerUpdateReservation)
	owner("DELETE /api/owner/reservations/{rid}", s.ownerDeleteReservation)
	owner("GET /api/owner/availability", s.ownerAvailability)
	owner("GET /api/owner/conversations", s.ownerConversations)
	owner("GET /api/owner/conversations/{cid}", s.ownerMessages)
	owner("POST /api/owner/conversations/{cid}", s.ownerReply)
	owner("POST /api/owner/assistant", s.ownerAssistant)

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	return s.cors(logRequests(mux))
}

// ---- Middleware & helpers ----

func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); s.origins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Vary", "Origin")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		if !(r.Method == http.MethodGet && (strings.Contains(r.URL.Path, "messages") || strings.HasSuffix(r.URL.Path, "/thread"))) &&
			!strings.Contains(r.URL.Path, "conversations") {
			log.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
		}
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func serverError(w http.ResponseWriter, err error) {
	if errors.Is(err, errNotFound) {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	log.Printf("error: %v", err)
	writeError(w, http.StatusInternalServerError, "something went wrong")
}

func readJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return false
	}
	return true
}

func pathInt(r *http.Request, name string) (int64, bool) {
	n, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	return n, err == nil
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if t, ok := strings.CutPrefix(h, "Bearer "); ok {
		return t
	}
	return ""
}

func (s *Server) requireOwner(h func(http.ResponseWriter, *http.Request, Business)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		businessID, err := s.store.SessionBusiness(bearerToken(r))
		if err != nil {
			writeError(w, http.StatusUnauthorized, "please log in with your business PIN")
			return
		}
		b, err := s.store.GetBusiness(businessID)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "business no longer exists")
			return
		}
		h(w, r, b)
	}
}

// pinLimiter locks a business's login for a while after repeated wrong PINs,
// since a 4-digit PIN is easy to guess otherwise.
type pinLimiter struct {
	mu       sync.Mutex
	failures map[string][]time.Time
}

const (
	maxPinFailures = 5
	pinWindow      = 5 * time.Minute
)

func newPinLimiter() *pinLimiter { return &pinLimiter{failures: map[string][]time.Time{}} }

func (l *pinLimiter) recent(id string) []time.Time {
	var keep []time.Time
	for _, t := range l.failures[id] {
		if time.Since(t) < pinWindow {
			keep = append(keep, t)
		}
	}
	l.failures[id] = keep
	return keep
}

func (l *pinLimiter) Locked(id string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.recent(id)) >= maxPinFailures
}

func (l *pinLimiter) Fail(id string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.failures[id] = append(l.recent(id), time.Now())
}

func (l *pinLimiter) Reset(id string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.failures, id)
}

var pinRe = regexp.MustCompile(`^\d{4,6}$`)

// ---- Validation ----

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(name string) string {
	return strings.Trim(slugRe.ReplaceAllString(strings.ToLower(name), "-"), "-")
}

func validateBusiness(b *Business) string {
	b.Name = strings.TrimSpace(b.Name)
	if b.Name == "" {
		return "name is required"
	}
	if b.Type != "restaurant" {
		b.Type = "service"
	}
	if b.SlotMinutes < 15 || b.SlotMinutes > 480 {
		return "slot length must be between 15 and 480 minutes"
	}
	if b.Capacity < 1 || b.Capacity > 1000 {
		return "capacity must be between 1 and 1000"
	}
	for _, h := range b.Hours {
		if !h.Closed && (minutes(h.Open) < 0 || minutes(h.Close) <= minutes(h.Open)) {
			return "opening hours for " + weekdayNames[h.Weekday%7] + " are invalid"
		}
	}
	if b.AutoReply.Rules == nil {
		b.AutoReply.Rules = []Rule{}
	}
	return ""
}

func validateOffering(o *Offering) string {
	o.Name = strings.TrimSpace(o.Name)
	if o.Name == "" {
		return "name is required"
	}
	switch o.Kind {
	case "service", "product", "menu":
	default:
		return "kind must be service, product or menu"
	}
	if o.Price < 0 || o.DurationMins < 0 {
		return "price and duration cannot be negative"
	}
	return ""
}

var validStatus = map[string]bool{"confirmed": true, "completed": true, "cancelled": true, "no_show": true}

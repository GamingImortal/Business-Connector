package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// ---- Public: customers ----

func (s *Server) listBusinesses(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListBusinesses()
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) getBusiness(w http.ResponseWriter, r *http.Request) {
	b, err := s.store.GetBusiness(r.PathValue("id"))
	if err != nil {
		serverError(w, err)
		return
	}
	offerings, err := s.store.ListOfferings(b.ID, true)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"business": b, "offerings": offerings})
}

type registerRequest struct {
	Business
	PIN string `json:"pin"`
}

func (s *Server) registerBusiness(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if !readJSON(w, r, &req) {
		return
	}
	if !pinRe.MatchString(req.PIN) {
		writeError(w, http.StatusBadRequest, "PIN must be 4 to 6 digits")
		return
	}
	b := req.Business
	b.Rating = 0
	if b.Avatar == "" {
		b.Avatar = "🏪"
	}
	if b.Color == "" {
		b.Color = "#00a884"
	}
	if b.SlotMinutes == 0 {
		b.SlotMinutes = 60
	}
	if b.Capacity == 0 {
		b.Capacity = 1
		if b.Type == "restaurant" {
			b.Capacity = 30
		}
	}
	if len(b.Hours) == 0 {
		for d := 1; d <= 6; d++ {
			b.Hours = append(b.Hours, DayHours{Weekday: d, Open: "09:00", Close: "17:00"})
		}
	}
	if b.AutoReply.Greeting == "" && b.AutoReply.Fallback == "" && len(b.AutoReply.Rules) == 0 {
		b.AutoReply = AutoReply{
			Enabled:  true,
			Greeting: fmt.Sprintf("Hi 👋 Welcome to %s! How can we help you today?", strings.TrimSpace(b.Name)),
			Fallback: "Thanks for your message! We'll get back to you shortly.",
		}
	}
	if msg := validateBusiness(&b); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	base := slugify(b.Name)
	if base == "" {
		base = "business"
	}
	b.ID = base
	for i := 2; ; i++ {
		if _, err := s.store.GetBusiness(b.ID); errors.Is(err, errNotFound) {
			break
		}
		b.ID = fmt.Sprintf("%s-%d", base, i)
	}

	if err := s.store.CreateBusiness(b, req.PIN); err != nil {
		serverError(w, err)
		return
	}
	token, err := s.store.CreateSession(b.ID)
	if err != nil {
		serverError(w, err)
		return
	}
	created, _ := s.store.GetBusiness(b.ID)
	writeJSON(w, http.StatusCreated, map[string]any{"token": token, "business": created})
}

func (s *Server) availability(w http.ResponseWriter, r *http.Request) {
	b, err := s.store.GetBusiness(r.PathValue("id"))
	if err != nil {
		serverError(w, err)
		return
	}
	s.writeAvailability(w, r, b)
}

func (s *Server) writeAvailability(w http.ResponseWriter, r *http.Request, b Business) {
	date := today()
	if q := r.URL.Query().Get("date"); q != "" {
		d, err := parseDate(q)
		if err != nil {
			writeError(w, http.StatusBadRequest, "date must look like 2026-10-26")
			return
		}
		date = d
	}
	party, _ := strconv.Atoi(r.URL.Query().Get("party"))
	slots, err := s.store.SlotsFor(b, date, max(party, 1))
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"date": date.Format(dateLayout), "dateLabel": dateLabel(date), "slots": slots})
}

func validCustomerID(id string) bool { return id != "" && len(id) <= 64 }

func (s *Server) customerMessages(w http.ResponseWriter, r *http.Request) {
	customerID := r.URL.Query().Get("customer")
	if !validCustomerID(customerID) {
		writeError(w, http.StatusBadRequest, "customer is required")
		return
	}
	after, _ := strconv.ParseInt(r.URL.Query().Get("after"), 10, 64)
	msgs, err := s.store.Messages(r.PathValue("id"), customerID, after)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, msgs)
}

type customerSendRequest struct {
	CustomerID   string `json:"customerId"`
	CustomerName string `json:"customerName"`
	Text         string `json:"text"`
}

func (s *Server) customerSend(w http.ResponseWriter, r *http.Request) {
	var req customerSendRequest
	if !readJSON(w, r, &req) {
		return
	}
	req.Text = strings.TrimSpace(req.Text)
	if !validCustomerID(req.CustomerID) || req.Text == "" || len(req.Text) > 2000 {
		writeError(w, http.StatusBadRequest, "customerId and a message (up to 2000 characters) are required")
		return
	}
	b, err := s.store.GetBusiness(r.PathValue("id"))
	if err != nil {
		serverError(w, err)
		return
	}
	if err := s.store.UpsertCustomer(req.CustomerID, strings.TrimSpace(req.CustomerName), ""); err != nil {
		serverError(w, err)
		return
	}

	firstContact := !s.store.HasBusinessReply(b.ID, req.CustomerID)
	msg := Message{BusinessID: b.ID, CustomerID: req.CustomerID, From: "client", Text: req.Text, Status: "sent"}
	if err := s.store.AddMessage(&msg); err != nil {
		serverError(w, err)
		return
	}

	replies, err := s.store.CustomerReplies(b, req.CustomerID, req.Text, firstContact)
	if err != nil {
		serverError(w, err)
		return
	}
	stored := []Message{}
	if len(replies) > 0 {
		s.store.MarkClientMessagesRead(b.ID, req.CustomerID)
		msg.Status = "read"
	}
	for _, rep := range replies {
		m, err := s.store.AddReply(b.ID, req.CustomerID, rep, true)
		if err != nil {
			serverError(w, err)
			return
		}
		stored = append(stored, m)
	}
	writeJSON(w, http.StatusOK, map[string]any{"message": msg, "replies": stored})
}

type customerBookRequest struct {
	CustomerID    string `json:"customerId"`
	CustomerName  string `json:"customerName"`
	CustomerPhone string `json:"customerPhone"`
	Date          string `json:"date"`
	Time          string `json:"time"`
	PartySize     int    `json:"partySize"`
	OfferingID    *int64 `json:"offeringId"`
	Notes         string `json:"notes"`
}

func (s *Server) customerBook(w http.ResponseWriter, r *http.Request) {
	var req customerBookRequest
	if !readJSON(w, r, &req) {
		return
	}
	if !validCustomerID(req.CustomerID) {
		writeError(w, http.StatusBadRequest, "customerId is required")
		return
	}
	b, err := s.store.GetBusiness(r.PathValue("id"))
	if err != nil {
		serverError(w, err)
		return
	}
	name := strings.TrimSpace(req.CustomerName)
	if name == "" {
		name = s.store.CustomerName(req.CustomerID)
	}
	if name == "" {
		name = "Customer"
	}
	s.store.UpsertCustomer(req.CustomerID, name, req.CustomerPhone)

	res := Reservation{
		CustomerID: req.CustomerID, CustomerName: name, CustomerPhone: req.CustomerPhone,
		Date: req.Date, Time: req.Time, PartySize: req.PartySize, OfferingID: req.OfferingID,
		Notes: req.Notes, Source: "chat",
	}
	if err := s.store.Book(b, &res); err != nil {
		if isBookingError(err) || strings.HasPrefix(err.Error(), "invalid") {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		serverError(w, err)
		return
	}
	res, _ = s.store.GetReservation(b.ID, res.ID)
	msg, err := s.store.AddReply(b.ID, req.CustomerID,
		Reply{Text: bookingConfirmation(b, res), Card: BookingCard{Type: "booking", Reservation: res}}, true)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"reservation": res, "message": msg})
}

func (s *Server) customerCancel(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CustomerID string `json:"customerId"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	id, ok := pathInt(r, "rid")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid reservation id")
		return
	}
	b, err := s.store.GetBusiness(r.PathValue("id"))
	if err != nil {
		serverError(w, err)
		return
	}
	res, err := s.store.GetReservation(b.ID, id)
	if err != nil || res.CustomerID != req.CustomerID || req.CustomerID == "" {
		writeError(w, http.StatusNotFound, "booking not found")
		return
	}
	if res.Status != "confirmed" {
		writeError(w, http.StatusConflict, "this booking is already "+strings.ReplaceAll(res.Status, "_", " "))
		return
	}
	res.Status = "cancelled"
	if err := s.store.UpdateReservation(res); err != nil {
		serverError(w, err)
		return
	}
	msg, err := s.store.AddReply(b.ID, req.CustomerID, Reply{
		Text: fmt.Sprintf("❌ Your booking #%d on %s at %s has been cancelled. Hope to see you another time!", res.ID, res.Date, res.Time),
	}, true)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"reservation": res, "message": msg})
}

func (s *Server) customerConversations(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.CustomerConversations(r.PathValue("cid"))
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// ---- Auth ----

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BusinessID string `json:"businessId"`
		PIN        string `json:"pin"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	if s.limiter.Locked(req.BusinessID) {
		writeError(w, http.StatusTooManyRequests, "too many wrong PINs, try again in a few minutes")
		return
	}
	if !s.store.CheckPIN(req.BusinessID, req.PIN) {
		s.limiter.Fail(req.BusinessID)
		writeError(w, http.StatusUnauthorized, "wrong business or PIN")
		return
	}
	s.limiter.Reset(req.BusinessID)
	token, err := s.store.CreateSession(req.BusinessID)
	if err != nil {
		serverError(w, err)
		return
	}
	b, _ := s.store.GetBusiness(req.BusinessID)
	writeJSON(w, http.StatusOK, map[string]any{"token": token, "business": b})
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	s.store.DeleteSession(bearerToken(r))
	w.WriteHeader(http.StatusNoContent)
}

// ---- Owner ----

func (s *Server) ownerGetBusiness(w http.ResponseWriter, r *http.Request, b Business) {
	writeJSON(w, http.StatusOK, b)
}

func (s *Server) ownerUpdateBusiness(w http.ResponseWriter, r *http.Request, b Business) {
	var in Business
	if !readJSON(w, r, &in) {
		return
	}
	in.ID, in.Rating = b.ID, b.Rating
	if msg := validateBusiness(&in); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	if err := s.store.UpdateBusiness(in); err != nil {
		serverError(w, err)
		return
	}
	updated, err := s.store.GetBusiness(b.ID)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) ownerChangePIN(w http.ResponseWriter, r *http.Request, b Business) {
	var req struct {
		CurrentPIN string `json:"currentPin"`
		NewPIN     string `json:"newPin"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	if !s.store.CheckPIN(b.ID, req.CurrentPIN) {
		writeError(w, http.StatusUnauthorized, "current PIN is wrong")
		return
	}
	if !pinRe.MatchString(req.NewPIN) {
		writeError(w, http.StatusBadRequest, "new PIN must be 4 to 6 digits")
		return
	}
	if err := s.store.SetPIN(b.ID, req.NewPIN); err != nil {
		serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) ownerListOfferings(w http.ResponseWriter, r *http.Request, b Business) {
	list, err := s.store.ListOfferings(b.ID, false)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) ownerCreateOffering(w http.ResponseWriter, r *http.Request, b Business) {
	var o Offering
	if !readJSON(w, r, &o) {
		return
	}
	o.BusinessID = b.ID
	if msg := validateOffering(&o); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	if err := s.store.CreateOffering(&o); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, o)
}

func (s *Server) ownerUpdateOffering(w http.ResponseWriter, r *http.Request, b Business) {
	var o Offering
	if !readJSON(w, r, &o) {
		return
	}
	id, ok := pathInt(r, "oid")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	o.ID, o.BusinessID = id, b.ID
	if msg := validateOffering(&o); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	if err := s.store.UpdateOffering(o); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, o)
}

func (s *Server) ownerDeleteOffering(w http.ResponseWriter, r *http.Request, b Business) {
	id, ok := pathInt(r, "oid")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := s.store.DeleteOffering(b.ID, id); err != nil {
		serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) ownerListReservations(w http.ResponseWriter, r *http.Request, b Business) {
	q := r.URL.Query()
	list, err := s.store.ListReservations(b.ID, q.Get("from"), q.Get("to"))
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) ownerCreateReservation(w http.ResponseWriter, r *http.Request, b Business) {
	var res Reservation
	if !readJSON(w, r, &res) {
		return
	}
	res.CustomerName = strings.TrimSpace(res.CustomerName)
	if res.CustomerName == "" {
		writeError(w, http.StatusBadRequest, "customer name is required")
		return
	}
	if res.Status != "" && !validStatus[res.Status] {
		writeError(w, http.StatusBadRequest, "invalid status")
		return
	}
	res.ID, res.CustomerID, res.Source = 0, "", "dashboard"
	if err := s.store.Book(b, &res); err != nil {
		if strings.HasPrefix(err.Error(), "invalid") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		serverError(w, err)
		return
	}
	res, _ = s.store.GetReservation(b.ID, res.ID)
	writeJSON(w, http.StatusCreated, res)
}

func (s *Server) ownerUpdateReservation(w http.ResponseWriter, r *http.Request, b Business) {
	id, ok := pathInt(r, "rid")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	current, err := s.store.GetReservation(b.ID, id)
	if err != nil {
		serverError(w, err)
		return
	}
	// Start from the stored booking so clients can send just the fields they change.
	updated := current
	if !readJSON(w, r, &updated) {
		return
	}
	updated.ID, updated.BusinessID, updated.CustomerID, updated.Source = current.ID, b.ID, current.CustomerID, current.Source
	if !validStatus[updated.Status] {
		writeError(w, http.StatusBadRequest, "invalid status")
		return
	}
	if _, err := parseDate(updated.Date); err != nil || minutes(updated.Time) < 0 {
		writeError(w, http.StatusBadRequest, "invalid date or time")
		return
	}
	updated.Time = hhmm(minutes(updated.Time))
	if err := s.store.UpdateReservation(updated); err != nil {
		serverError(w, err)
		return
	}

	// Let chat customers know when the business changes their booking.
	if current.CustomerID != "" {
		var note string
		switch {
		case updated.Status == "cancelled" && current.Status != "cancelled":
			note = fmt.Sprintf("❌ Sorry, %s had to cancel your booking #%d on %s at %s. Reply here to rebook.", b.Name, updated.ID, updated.Date, updated.Time)
		case updated.Status == "completed" && current.Status != "completed":
			note = fmt.Sprintf("🙏 Thanks for visiting %s! We hope to see you again soon.", b.Name)
		case updated.Date != current.Date || updated.Time != current.Time:
			note = fmt.Sprintf("🔁 Your booking #%d has been moved to %s at %s.", updated.ID, updated.Date, updated.Time)
		}
		if note != "" {
			s.store.AddReply(b.ID, current.CustomerID, Reply{Text: note}, true)
		}
	}
	updated, _ = s.store.GetReservation(b.ID, id)
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) ownerDeleteReservation(w http.ResponseWriter, r *http.Request, b Business) {
	id, ok := pathInt(r, "rid")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := s.store.DeleteReservation(b.ID, id); err != nil {
		serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) ownerAvailability(w http.ResponseWriter, r *http.Request, b Business) {
	s.writeAvailability(w, r, b)
}

func (s *Server) ownerConversations(w http.ResponseWriter, r *http.Request, b Business) {
	list, err := s.store.BusinessConversations(b.ID)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) ownerMessages(w http.ResponseWriter, r *http.Request, b Business) {
	customerID := r.PathValue("cid")
	after, _ := strconv.ParseInt(r.URL.Query().Get("after"), 10, 64)
	s.store.MarkClientMessagesRead(b.ID, customerID)
	msgs, err := s.store.Messages(b.ID, customerID, after)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, msgs)
}

func (s *Server) ownerReply(w http.ResponseWriter, r *http.Request, b Business) {
	var req struct {
		Text string `json:"text"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	req.Text = strings.TrimSpace(req.Text)
	if req.Text == "" || len(req.Text) > 2000 {
		writeError(w, http.StatusBadRequest, "message is required (up to 2000 characters)")
		return
	}
	customerID := r.PathValue("cid")
	s.store.MarkClientMessagesRead(b.ID, customerID)
	msg, err := s.store.AddReply(b.ID, customerID, Reply{Text: req.Text}, false)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, msg)
}

func (s *Server) ownerAssistant(w http.ResponseWriter, r *http.Request, b Business) {
	var req struct {
		Text string `json:"text"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	replies, err := s.store.OwnerReplies(b, req.Text)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"replies": replies})
}

// ---- Concierge ----

func (s *Server) conciergeSend(w http.ResponseWriter, r *http.Request) {
	var req struct {
		customerSendRequest
		Context string `json:"context"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	req.Text = strings.TrimSpace(req.Text)
	if !validCustomerID(req.CustomerID) || req.Text == "" || len(req.Text) > 2000 {
		writeError(w, http.StatusBadRequest, "customerId and a message (up to 2000 characters) are required")
		return
	}
	if err := s.store.UpsertCustomer(req.CustomerID, strings.TrimSpace(req.CustomerName), ""); err != nil {
		serverError(w, err)
		return
	}
	result, err := s.store.Concierge(req.CustomerID, req.Text, req.Context)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) customerThread(w http.ResponseWriter, r *http.Request) {
	customerID := r.PathValue("cid")
	if !validCustomerID(customerID) {
		writeError(w, http.StatusBadRequest, "invalid customer")
		return
	}
	after, _ := strconv.ParseInt(r.URL.Query().Get("after"), 10, 64)
	msgs, err := s.store.CustomerThread(customerID, after)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, msgs)
}

package handler

import (
	"net/http"

	"fahrradverleih/internal/domain"
)

// ---- Ausleihe (Mitarbeiter) ----

type rentalView struct {
	domain.Rental
	Bike         domain.Bike `json:"bike"`
	CustomerName string      `json:"customerName"`
	Overdue      bool        `json:"overdue"`
}

func (h *Handler) listRentals(w http.ResponseWriter, r *http.Request) {
	now := h.now()
	names := h.customerNames()
	bikes := map[string]domain.Bike{}
	for _, st := range h.store.BikeStates() {
		bikes[st.Bike.InventoryNo] = st.Bike
	}
	views := []rentalView{}
	for _, rt := range h.store.ActiveRentals() {
		views = append(views, rentalView{
			Rental:       rt,
			Bike:         bikes[rt.BikeNo],
			CustomerName: names[rt.CustomerID],
			Overdue:      rt.Overdue(now),
		})
	}
	writeJSON(w, http.StatusOK, views)
}

func (h *Handler) rentalOptions(w http.ResponseWriter, r *http.Request) {
	id, err := intParam(r.URL.Query().Get("customerId"), "customerId")
	if err != nil {
		writeError(w, err)
		return
	}
	now := h.now()
	options, err := h.store.RentalOptions(id, now)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"start":   domain.HourBlock(now),
		"options": options,
	})
}

type createRentalRequest struct {
	BikeNo     string `json:"bikeNo"`
	CustomerID int    `json:"customerId"`
	ReturnDate string `json:"returnDate"` // JJJJ-MM-TT
	ReturnHour int    `json:"returnHour"` // Rückgabe-Stundenblock 0–23 (7 = bis 07:59)
}

func (h *Handler) createRental(w http.ResponseWriter, r *http.Request) {
	var req createRentalRequest
	if err := readJSON(w, r, &req, 1<<16); err != nil {
		writeError(w, err)
		return
	}
	period, err := domain.NewRentalPeriod(req.ReturnDate, req.ReturnHour, h.now())
	if err != nil {
		writeError(w, err)
		return
	}
	rental, err := h.store.StartRental(req.BikeNo, req.CustomerID, period)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, rental)
}

func (h *Handler) endRental(w http.ResponseWriter, r *http.Request) {
	id, err := intParam(r.PathValue("id"), "id")
	if err != nil {
		writeError(w, err)
		return
	}
	rental, err := h.store.EndRental(id, h.now())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rental)
}

// ---- Reservierung (Kundenseite) ----

type reservationView struct {
	domain.Reservation
	Bike domain.Bike `json:"bike"`
}

func (h *Handler) listReservations(w http.ResponseWriter, r *http.Request) {
	id, err := intParam(r.URL.Query().Get("customerId"), "customerId")
	if err != nil {
		writeError(w, err)
		return
	}
	bikes := map[string]domain.Bike{}
	for _, st := range h.store.BikeStates() {
		bikes[st.Bike.InventoryNo] = st.Bike
	}
	views := []reservationView{}
	for _, res := range h.store.ReservationsOf(id) {
		views = append(views, reservationView{
			Reservation: res,
			Bike:        bikes[res.BikeNo],
		})
	}
	writeJSON(w, http.StatusOK, views)
}

func (h *Handler) reservationOptions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	id, err := intParam(q.Get("customerId"), "customerId")
	if err != nil {
		writeError(w, err)
		return
	}
	fromHour, err := intParam(q.Get("fromHour"), "fromHour")
	if err != nil {
		writeError(w, err)
		return
	}
	toHour, err := intParam(q.Get("toHour"), "toHour")
	if err != nil {
		writeError(w, err)
		return
	}
	now := h.now()
	period, err := domain.NewReservationPeriod(q.Get("from"), fromHour, q.Get("to"), toHour, now)
	if err != nil {
		writeError(w, err)
		return
	}
	bikes, err := h.store.ReservationOptions(id, period, now)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, bikes)
}

type createReservationRequest struct {
	BikeNo     string `json:"bikeNo"`
	CustomerID int    `json:"customerId"`
	From       string `json:"from"`     // JJJJ-MM-TT
	FromHour   int    `json:"fromHour"` // erster Stundenblock 0–23
	To         string `json:"to"`       // JJJJ-MM-TT
	ToHour     int    `json:"toHour"`   // letzter Stundenblock 0–23, inklusive
}

func (h *Handler) createReservation(w http.ResponseWriter, r *http.Request) {
	var req createReservationRequest
	if err := readJSON(w, r, &req, 1<<16); err != nil {
		writeError(w, err)
		return
	}
	now := h.now()
	period, err := domain.NewReservationPeriod(req.From, req.FromHour, req.To, req.ToHour, now)
	if err != nil {
		writeError(w, err)
		return
	}
	res, err := h.store.AddReservation(req.BikeNo, req.CustomerID, period, now)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, res)
}

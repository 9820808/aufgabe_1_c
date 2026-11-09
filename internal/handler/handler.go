// Package handler stellt die JSON-API unter /api bereit.
package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"fahrradverleih/internal/domain"
	"fahrradverleih/internal/store"
)

type Handler struct {
	store *store.Store
	now   func() time.Time // austauschbar für Tests
}

func New(s *store.Store) *Handler {
	return &Handler{store: s, now: time.Now}
}

// Register hängt alle API-Routen an den Mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/bike-types", h.listBikeTypes)
	mux.HandleFunc("GET /api/bikes", h.listBikes)
	mux.HandleFunc("GET /api/bikes/next-number", h.suggestInventoryNo)
	mux.HandleFunc("POST /api/bikes", h.createBike)
	mux.HandleFunc("DELETE /api/bikes/{no}", h.deleteBike)

	mux.HandleFunc("GET /api/customers", h.listCustomers)
	mux.HandleFunc("GET /api/customers/names", h.listCustomerNames)
	mux.HandleFunc("POST /api/customers", h.createCustomer)
	mux.HandleFunc("POST /api/customers/{id}/block", h.blockCustomer(true))
	mux.HandleFunc("POST /api/customers/{id}/unblock", h.blockCustomer(false))
	mux.HandleFunc("GET /api/customers/{id}/photo", h.customerPhoto)

	mux.HandleFunc("GET /api/rentals", h.listRentals)
	mux.HandleFunc("GET /api/rentals/options", h.rentalOptions)
	mux.HandleFunc("POST /api/rentals", h.createRental)
	mux.HandleFunc("POST /api/rentals/{id}/end", h.endRental)

	mux.HandleFunc("GET /api/reservations", h.listReservations)
	mux.HandleFunc("GET /api/reservations/options", h.reservationOptions)
	mux.HandleFunc("POST /api/reservations", h.createReservation)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("JSON schreiben: %v", err)
	}
}

// writeError übersetzt fachliche Fehler in HTTP-Statuscodes.
func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	var de *domain.Error
	if errors.As(err, &de) {
		switch de.Kind {
		case domain.KindInvalid:
			status = http.StatusBadRequest
		case domain.KindNotFound:
			status = http.StatusNotFound
		case domain.KindConflict:
			status = http.StatusConflict
		}
	} else {
		log.Printf("interner Fehler: %v", err)
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func readJSON(w http.ResponseWriter, r *http.Request, v any, maxBytes int64) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return domain.Invalid("ungültige Anfrage: %v", err)
	}
	return nil
}

func intParam(s, name string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, domain.Invalid("Parameter %s fehlt oder ist keine Zahl", name)
	}
	return n, nil
}

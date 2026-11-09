package handler

import (
	"net/http"
	"strconv"
	"time"

	"fahrradverleih/internal/domain"
)

type rentalInfo struct {
	RentalID     int       `json:"rentalId"`
	CustomerID   int       `json:"customerId"`
	CustomerName string    `json:"customerName"`
	Until        time.Time `json:"until"`
	Overdue      bool      `json:"overdue"`
}

type bikeView struct {
	domain.Bike
	RentedTo     *rentalInfo `json:"rentedTo,omitempty"`
	ReservedFrom *time.Time  `json:"reservedFrom,omitempty"` // Beginn der nächsten Reservierung
}

func (h *Handler) listBikeTypes(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, domain.BikeTypes())
}

func (h *Handler) listBikes(w http.ResponseWriter, r *http.Request) {
	now := h.now()
	names := h.customerNames()
	views := []bikeView{}
	for _, st := range h.store.BikeStates() {
		v := bikeView{Bike: st.Bike}
		if st.Rental != nil {
			v.RentedTo = &rentalInfo{
				RentalID:     st.Rental.ID,
				CustomerID:   st.Rental.CustomerID,
				CustomerName: names[st.Rental.CustomerID],
				Until:        st.Rental.Period.End,
				Overdue:      st.Rental.Overdue(now),
			}
		}
		if next := st.NextReservation(now); next != nil {
			v.ReservedFrom = &next.Period.Start
		}
		views = append(views, v)
	}
	writeJSON(w, http.StatusOK, views)
}

func (h *Handler) suggestInventoryNo(w http.ResponseWriter, r *http.Request) {
	year := h.now().Year()
	if y := r.URL.Query().Get("year"); y != "" {
		n, err := strconv.Atoi(y)
		if err != nil || n < 1000 || n > 9999 {
			writeError(w, domain.Invalid("ungültiges Jahr %q", y))
			return
		}
		year = n
	}
	writeJSON(w, http.StatusOK, map[string]string{"inventoryNo": h.store.SuggestInventoryNo(year)})
}

func (h *Handler) createBike(w http.ResponseWriter, r *http.Request) {
	var b domain.Bike
	if err := readJSON(w, r, &b, 1<<16); err != nil {
		writeError(w, err)
		return
	}
	created, err := h.store.AddBike(b)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) deleteBike(w http.ResponseWriter, r *http.Request) {
	if err := h.store.DeleteBike(r.PathValue("no")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// customerNames liefert eine Zuordnung Kunden-ID → Name.
func (h *Handler) customerNames() map[int]string {
	names := map[int]string{}
	for _, c := range h.store.Customers() {
		names[c.ID] = c.Name
	}
	return names
}

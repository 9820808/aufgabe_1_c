package handler

import (
	"encoding/base64"
	"net/http"
	"strings"

	"fahrradverleih/internal/domain"
)

// customerView ist die Mitarbeiter-Sicht auf einen Kunden – ohne Fotodaten,
// das Foto wird separat über /photo geladen.
type customerView struct {
	domain.Customer
	HasPhoto bool `json:"hasPhoto"`
}

// customerName ist die öffentliche Sicht für die Kundenseite (nur Name).
type customerName struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func toCustomerView(c domain.Customer) customerView {
	v := customerView{Customer: c, HasPhoto: c.HasPhoto()}
	v.Photo = ""
	return v
}

func (h *Handler) listCustomers(w http.ResponseWriter, r *http.Request) {
	views := []customerView{}
	for _, c := range h.store.Customers() {
		views = append(views, toCustomerView(c))
	}
	writeJSON(w, http.StatusOK, views)
}

func (h *Handler) listCustomerNames(w http.ResponseWriter, r *http.Request) {
	names := []customerName{}
	for _, c := range h.store.Customers() {
		names = append(names, customerName{ID: c.ID, Name: c.Name})
	}
	writeJSON(w, http.StatusOK, names)
}

func (h *Handler) createCustomer(w http.ResponseWriter, r *http.Request) {
	var c domain.Customer
	if err := readJSON(w, r, &c, 10<<20); err != nil { // Foto max. ~7 MB
		writeError(w, err)
		return
	}
	created, err := h.store.AddCustomer(c)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toCustomerView(created))
}

func (h *Handler) blockCustomer(blocked bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := intParam(r.PathValue("id"), "id")
		if err != nil {
			writeError(w, err)
			return
		}
		c, err := h.store.SetCustomerBlocked(id, blocked)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, toCustomerView(c))
	}
}

// customerPhoto liefert das Ausweisfoto als Bild (aus der gespeicherten Data-URL).
func (h *Handler) customerPhoto(w http.ResponseWriter, r *http.Request) {
	id, err := intParam(r.PathValue("id"), "id")
	if err != nil {
		writeError(w, err)
		return
	}
	c, err := h.store.Customer(id)
	if err != nil {
		writeError(w, err)
		return
	}
	// Format: data:image/png;base64,....
	meta, data, ok := strings.Cut(strings.TrimPrefix(c.Photo, "data:"), ",")
	mime, isBase64 := strings.CutSuffix(meta, ";base64")
	if !ok || !isBase64 {
		writeError(w, domain.NotFound("kein Foto vorhanden"))
		return
	}
	img, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		writeError(w, domain.NotFound("Foto ist beschädigt"))
		return
	}
	w.Header().Set("Content-Type", mime)
	w.Write(img)
}

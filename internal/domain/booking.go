package domain

import "time"

// Rental ist eine Ausleihe eines Fahrrads an einen Kunden.
type Rental struct {
	ID         int        `json:"id"`
	BikeNo     string     `json:"bikeNo"`
	CustomerID int        `json:"customerId"`
	Period     Interval   `json:"period"`
	ReturnedAt *time.Time `json:"returnedAt,omitempty"`
}

func (r Rental) Active() bool { return r.ReturnedAt == nil }

// Overdue ist true, wenn das Fahrrad nach der vereinbarten Rückgabe noch nicht zurück ist.
func (r Rental) Overdue(now time.Time) bool {
	return r.Active() && !now.Before(r.Period.End)
}

// Close beendet die Ausleihe.
func (r *Rental) Close(now time.Time) error {
	if !r.Active() {
		return Conflict("Ausleihe ist bereits beendet")
	}
	r.ReturnedAt = &now
	return nil
}

// Reservation ist ein in der Zukunft liegendes Ausleihen im Stundenraster.
type Reservation struct {
	ID         int      `json:"id"`
	BikeNo     string   `json:"bikeNo"`
	CustomerID int      `json:"customerId"`
	Period     Interval `json:"period"`
}

// BikeState bündelt ein Fahrrad mit seiner aktiven Ausleihe und seinen
// Reservierungen. Hier liegen die Regeln für Verleih und Reservierung.
type BikeState struct {
	Bike         Bike
	Rental       *Rental // aktive Ausleihe oder nil
	Reservations []Reservation
}

// NextReservation liefert die nächste noch nicht abgelaufene Reservierung.
func (s BikeState) NextReservation(now time.Time) *Reservation {
	var next *Reservation
	for i, r := range s.Reservations {
		if r.Period.End.After(now) && (next == nil || r.Period.Start.Before(next.Period.Start)) {
			next = &s.Reservations[i]
		}
	}
	return next
}

// MaxRentalEnd liefert das späteste (exklusive) Ende einer Ausleihe ab start
// durch den Kunden: den Beginn der nächsten Reservierung eines anderen Kunden.
// Reservierung ab 08:00 → Rückgabe bis spätestens 07:59.
// Eigene Reservierungen begrenzen nicht (Abholung der Reservierung).
// nil bedeutet: keine Grenze.
func (s BikeState) MaxRentalEnd(customerID int, start time.Time) *time.Time {
	var limit *time.Time
	for _, r := range s.Reservations {
		if r.CustomerID == customerID || !r.Period.End.After(start) {
			continue
		}
		l := r.Period.Start
		if l.Before(start) {
			l = start
		}
		if limit == nil || l.Before(*limit) {
			limit = &l
		}
	}
	return limit
}

// CanRentTo prüft, ob das Fahrrad dem Kunden ab start überhaupt angeboten werden kann.
func (s BikeState) CanRentTo(c Customer, start time.Time) bool {
	if s.Rental != nil || !s.Bike.Saddle.Overlaps(c.Saddle) {
		return false
	}
	limit := s.MaxRentalEnd(c.ID, start)
	return limit == nil || limit.After(start)
}

// CheckRental prüft, ob das Fahrrad im Zeitraum p an den Kunden verliehen werden darf.
func (s BikeState) CheckRental(c Customer, p Interval) error {
	if err := c.CanBook(); err != nil {
		return err
	}
	if s.Rental != nil {
		return Conflict("Fahrrad %s ist bereits verliehen", s.Bike.InventoryNo)
	}
	if !s.Bike.Saddle.Overlaps(c.Saddle) {
		return Invalid("Sattelhöhe von Fahrrad %s passt nicht zum Kunden", s.Bike.InventoryNo)
	}
	if limit := s.MaxRentalEnd(c.ID, p.Start); limit != nil && p.End.After(*limit) {
		if !limit.After(p.Start) {
			return Conflict("Fahrrad %s ist derzeit für einen anderen Kunden reserviert", s.Bike.InventoryNo)
		}
		return Conflict("Fahrrad %s ist ab %s reserviert – Rückgabe spätestens %s",
			s.Bike.InventoryNo, limit.Format(DisplayFormat), LastMinute(*limit).Format(DisplayFormat))
	}
	return nil
}

// CheckReservation prüft, ob das Fahrrad im Zeitraum p für den Kunden reserviert werden darf.
func (s BikeState) CheckReservation(c Customer, p Interval, now time.Time) error {
	if err := c.CanBook(); err != nil {
		return err
	}
	if !s.Bike.Saddle.Overlaps(c.Saddle) {
		return Invalid("Sattelhöhe von Fahrrad %s passt nicht zum Kunden", s.Bike.InventoryNo)
	}
	if s.Rental != nil {
		if s.Rental.Overdue(now) {
			return Conflict("Fahrrad %s ist überfällig und erst nach Rückgabe reservierbar", s.Bike.InventoryNo)
		}
		if s.Rental.Period.Overlaps(p) {
			return Conflict("Fahrrad %s ist in diesem Zeitraum verliehen", s.Bike.InventoryNo)
		}
	}
	for _, r := range s.Reservations {
		if r.Period.Overlaps(p) {
			return Conflict("Fahrrad %s ist in diesem Zeitraum bereits reserviert", s.Bike.InventoryNo)
		}
	}
	return nil
}

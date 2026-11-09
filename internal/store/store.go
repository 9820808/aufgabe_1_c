// Package store hält alle Daten im Speicher. Alle Methoden sind thread-safe;
// fachliche Prüfungen laufen unter demselben Lock wie die Änderung.
package store

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"fahrradverleih/internal/domain"
)

type Store struct {
	mu           sync.RWMutex
	bikes        map[string]domain.Bike
	customers    map[int]domain.Customer
	rentals      map[int]domain.Rental
	reservations map[int]domain.Reservation
	lastID       int // gemeinsamer Zähler für Kunden-, Ausleih- und Reservierungs-IDs
}

func New() *Store {
	return &Store{
		bikes:        map[string]domain.Bike{},
		customers:    map[int]domain.Customer{},
		rentals:      map[int]domain.Rental{},
		reservations: map[int]domain.Reservation{},
	}
}

func (s *Store) nextID() int {
	s.lastID++
	return s.lastID
}

// ---- Fahrräder ----

func (s *Store) AddBike(b domain.Bike) (domain.Bike, error) {
	if err := b.Validate(); err != nil {
		return domain.Bike{}, fmt.Errorf("Fahrrad anlegen: %w", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.bikes[b.InventoryNo]; ok {
		return domain.Bike{}, domain.Conflict("Inventarnummer %s existiert bereits", b.InventoryNo)
	}
	s.bikes[b.InventoryNo] = b
	return b, nil
}

// DeleteBike löscht ein Fahrrad samt Reservierungen. Verliehene Fahrräder
// können nicht gelöscht werden.
func (s *Store) DeleteBike(no string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.bikes[no]; !ok {
		return domain.NotFound("Fahrrad %s nicht gefunden", no)
	}
	if s.activeRental(no) != nil {
		return domain.Conflict("Fahrrad %s ist verliehen und kann nicht gelöscht werden", no)
	}
	delete(s.bikes, no)
	for id, r := range s.reservations {
		if r.BikeNo == no {
			delete(s.reservations, id)
		}
	}
	return nil
}

// SuggestInventoryNo schlägt die nächste fortlaufende Nummer für das Jahr vor.
func (s *Store) SuggestInventoryNo(year int) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	highest := 0
	for no := range s.bikes {
		if y, n, err := domain.ParseInventoryNo(no); err == nil && y == year && n > highest {
			highest = n
		}
	}
	return domain.FormatInventoryNo(year, highest+1)
}

// BikeStates liefert alle Fahrräder mit Ausleih- und Reservierungsinfo, sortiert nach Inventarnummer.
func (s *Store) BikeStates() []domain.BikeState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	states := make([]domain.BikeState, 0, len(s.bikes))
	for no := range s.bikes {
		states = append(states, s.bikeState(no))
	}
	sort.Slice(states, func(i, j int) bool { return states[i].Bike.InventoryNo < states[j].Bike.InventoryNo })
	return states
}

// bikeState setzt voraus, dass der Aufrufer das Lock hält.
func (s *Store) bikeState(no string) domain.BikeState {
	st := domain.BikeState{Bike: s.bikes[no], Rental: s.activeRental(no)}
	for _, r := range s.reservations {
		if r.BikeNo == no {
			st.Reservations = append(st.Reservations, r)
		}
	}
	return st
}

func (s *Store) activeRental(bikeNo string) *domain.Rental {
	for _, r := range s.rentals {
		if r.BikeNo == bikeNo && r.Active() {
			return &r
		}
	}
	return nil
}

// ---- Kunden ----

func (s *Store) AddCustomer(c domain.Customer) (domain.Customer, error) {
	c.Blocked = false
	if err := c.Validate(); err != nil {
		return domain.Customer{}, fmt.Errorf("Kunde anlegen: %w", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	c.ID = s.nextID()
	s.customers[c.ID] = c
	return c, nil
}

func (s *Store) Customer(id int) (domain.Customer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.customer(id)
}

func (s *Store) customer(id int) (domain.Customer, error) {
	c, ok := s.customers[id]
	if !ok {
		return domain.Customer{}, domain.NotFound("Kunde %d nicht gefunden", id)
	}
	return c, nil
}

// Customers liefert alle Kunden sortiert nach Name.
func (s *Store) Customers() []domain.Customer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]domain.Customer, 0, len(s.customers))
	for _, c := range s.customers {
		list = append(list, c)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
	return list
}

func (s *Store) SetCustomerBlocked(id int, blocked bool) (domain.Customer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, err := s.customer(id)
	if err != nil {
		return domain.Customer{}, err
	}
	c.Blocked = blocked
	s.customers[id] = c
	return c, nil
}

// ---- Ausleihe ----

// RentalOption ist ein Fahrrad, das einem Kunden jetzt verliehen werden kann.
type RentalOption struct {
	Bike      domain.Bike `json:"bike"`
	MaxReturn *time.Time  `json:"maxReturn,omitempty"` // exklusives Ende wegen Reservierung (08:00 = bis 07:59)
}

// RentalOptions listet die freien, passenden Fahrräder für den Kunden.
func (s *Store) RentalOptions(customerID int, now time.Time) ([]RentalOption, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, err := s.customer(customerID)
	if err != nil {
		return nil, err
	}
	start := domain.HourBlock(now)
	options := []RentalOption{}
	for no := range s.bikes {
		st := s.bikeState(no)
		if st.CanRentTo(c, start) {
			options = append(options, RentalOption{Bike: st.Bike, MaxReturn: st.MaxRentalEnd(c.ID, start)})
		}
	}
	sort.Slice(options, func(i, j int) bool { return options[i].Bike.InventoryNo < options[j].Bike.InventoryNo })
	return options, nil
}

func (s *Store) StartRental(bikeNo string, customerID int, p domain.Interval) (domain.Rental, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, err := s.customer(customerID)
	if err != nil {
		return domain.Rental{}, err
	}
	if _, ok := s.bikes[bikeNo]; !ok {
		return domain.Rental{}, domain.NotFound("Fahrrad %s nicht gefunden", bikeNo)
	}
	if err := s.bikeState(bikeNo).CheckRental(c, p); err != nil {
		return domain.Rental{}, fmt.Errorf("Verleih: %w", err)
	}
	r := domain.Rental{ID: s.nextID(), BikeNo: bikeNo, CustomerID: customerID, Period: p}
	s.rentals[r.ID] = r
	return r, nil
}

// ActiveRentals liefert alle laufenden Ausleihen.
func (s *Store) ActiveRentals() []domain.Rental {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := []domain.Rental{}
	for _, r := range s.rentals {
		if r.Active() {
			list = append(list, r)
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].BikeNo < list[j].BikeNo })
	return list
}

func (s *Store) EndRental(id int, now time.Time) (domain.Rental, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.rentals[id]
	if !ok {
		return domain.Rental{}, domain.NotFound("Ausleihe %d nicht gefunden", id)
	}
	if err := r.Close(now); err != nil {
		return domain.Rental{}, fmt.Errorf("Rückgabe: %w", err)
	}
	s.rentals[id] = r
	return r, nil
}

// ---- Reservierung ----

// ReservationOptions listet die Fahrräder, die der Kunde im Zeitraum reservieren kann.
func (s *Store) ReservationOptions(customerID int, p domain.Interval, now time.Time) ([]domain.Bike, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, err := s.customer(customerID)
	if err != nil {
		return nil, err
	}
	if err := c.CanBook(); err != nil {
		return nil, err
	}
	bikes := []domain.Bike{}
	for no := range s.bikes {
		st := s.bikeState(no)
		if st.CheckReservation(c, p, now) == nil {
			bikes = append(bikes, st.Bike)
		}
	}
	sort.Slice(bikes, func(i, j int) bool { return bikes[i].InventoryNo < bikes[j].InventoryNo })
	return bikes, nil
}

func (s *Store) AddReservation(bikeNo string, customerID int, p domain.Interval, now time.Time) (domain.Reservation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, err := s.customer(customerID)
	if err != nil {
		return domain.Reservation{}, err
	}
	if _, ok := s.bikes[bikeNo]; !ok {
		return domain.Reservation{}, domain.NotFound("Fahrrad %s nicht gefunden", bikeNo)
	}
	if err := s.bikeState(bikeNo).CheckReservation(c, p, now); err != nil {
		return domain.Reservation{}, fmt.Errorf("Reservierung: %w", err)
	}
	r := domain.Reservation{ID: s.nextID(), BikeNo: bikeNo, CustomerID: customerID, Period: p}
	s.reservations[r.ID] = r
	return r, nil
}

// ReservationsOf liefert die Reservierungen eines Kunden, nach Beginn sortiert.
func (s *Store) ReservationsOf(customerID int) []domain.Reservation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := []domain.Reservation{}
	for _, r := range s.reservations {
		if r.CustomerID == customerID {
			list = append(list, r)
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Period.Start.Before(list[j].Period.Start) })
	return list
}

package store

import (
	"errors"
	"testing"
	"time"

	"fahrradverleih/internal/domain"
)

// Fester "Jetzt"-Zeitpunkt: 03.11.2026 10:30 lokal.
var now = time.Date(2026, 11, 3, 10, 30, 0, 0, time.Local)

func setup(t *testing.T) (*Store, domain.Customer, domain.Customer) {
	t.Helper()
	s := New()
	if _, err := s.AddBike(domain.Bike{InventoryNo: "2026-001", Brand: "Cube", Type: "Rennrad", Model: "X", Saddle: domain.SaddleRange{Min: 70, Max: 90}, WheelInch: 28}); err != nil {
		t.Fatal(err)
	}
	a, err := s.AddCustomer(domain.Customer{Name: "A", DocType: domain.DocPassport, DocNumber: "P1", Saddle: domain.SaddleRange{Min: 80, Max: 85}})
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.AddCustomer(domain.Customer{Name: "B", DocType: domain.DocPassport, DocNumber: "P2", Saddle: domain.SaddleRange{Min: 80, Max: 85}})
	if err != nil {
		t.Fatal(err)
	}
	return s, a, b
}

func kind(err error) domain.ErrorKind {
	var de *domain.Error
	if errors.As(err, &de) {
		return de.Kind
	}
	return 0
}

func reservation(t *testing.T, from string, fromHour int, to string, toHour int) domain.Interval {
	t.Helper()
	p, err := domain.NewReservationPeriod(from, fromHour, to, toHour, now)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func rental(t *testing.T, date string, hour int) domain.Interval {
	t.Helper()
	p, err := domain.NewRentalPeriod(date, hour, now)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestSuggestInventoryNo(t *testing.T) {
	s, _, _ := setup(t)
	if got := s.SuggestInventoryNo(2026); got != "2026-002" {
		t.Errorf("got %s", got)
	}
	if got := s.SuggestInventoryNo(2025); got != "2025-001" {
		t.Errorf("got %s", got)
	}
}

func TestCustomerNeedsDocOrPhoto(t *testing.T) {
	s := New()
	_, err := s.AddCustomer(domain.Customer{Name: "X", Saddle: domain.SaddleRange{Min: 1, Max: 2}})
	if kind(err) != domain.KindInvalid {
		t.Fatalf("erwartet Validierungsfehler, got %v", err)
	}
	if _, err := s.AddCustomer(domain.Customer{Name: "X", Photo: "data:image/png;base64,AA==", Saddle: domain.SaddleRange{Min: 1, Max: 2}}); err != nil {
		t.Fatalf("nur Foto sollte reichen: %v", err)
	}
}

func TestBlockedCustomerCannotRentOrReserve(t *testing.T) {
	s, a, _ := setup(t)
	s.SetCustomerBlocked(a.ID, true)
	if _, err := s.StartRental("2026-001", a.ID, rental(t, "2026-11-04", 10)); kind(err) != domain.KindConflict {
		t.Errorf("Verleih: got %v", err)
	}
	if _, err := s.AddReservation("2026-001", a.ID, reservation(t, "2026-11-05", 8, "2026-11-05", 9), now); kind(err) != domain.KindConflict {
		t.Errorf("Reservierung: got %v", err)
	}
}

func TestReservationMustBeInFuture(t *testing.T) {
	if _, err := domain.NewReservationPeriod("2026-11-03", 20, "2026-11-04", 8, now); kind(err) != domain.KindInvalid {
		t.Errorf("heute sollte nicht reservierbar sein, got %v", err)
	}
}

func TestRentalLimitedByForeignReservation(t *testing.T) {
	s, a, b := setup(t)
	if _, err := s.AddReservation("2026-001", b.ID, reservation(t, "2026-11-05", 8, "2026-11-06", 17), now); err != nil {
		t.Fatal(err)
	}
	opts, err := s.RentalOptions(a.ID, now)
	if err != nil || len(opts) != 1 || opts[0].MaxReturn == nil {
		t.Fatalf("erwartet Option mit Rückgabegrenze, got %+v %v", opts, err)
	}
	want := time.Date(2026, 11, 5, 8, 0, 0, 0, time.Local) // Rückgabe bis 07:59
	if !opts[0].MaxReturn.Equal(want) {
		t.Errorf("MaxReturn = %v, want %v", opts[0].MaxReturn, want)
	}
	if _, err := s.StartRental("2026-001", a.ID, rental(t, "2026-11-05", 8)); kind(err) != domain.KindConflict {
		t.Errorf("Rückgabe im ersten reservierten Block (08:00–08:59) muss abgelehnt werden, got %v", err)
	}
	if _, err := s.StartRental("2026-001", a.ID, rental(t, "2026-11-05", 7)); err != nil {
		t.Errorf("Rückgabe im Block 07:00–07:59 ist erlaubt: %v", err)
	}
}

func TestOwnReservationDoesNotLimitRental(t *testing.T) {
	s, a, _ := setup(t)
	if _, err := s.AddReservation("2026-001", a.ID, reservation(t, "2026-11-05", 8, "2026-11-06", 17), now); err != nil {
		t.Fatal(err)
	}
	if _, err := s.StartRental("2026-001", a.ID, rental(t, "2026-11-08", 10)); err != nil {
		t.Errorf("Abholung über eigene Reservierung hinweg muss gehen: %v", err)
	}
}

func TestReservationBlockedByRentalAndOverdue(t *testing.T) {
	s, a, b := setup(t)
	r, err := s.StartRental("2026-001", a.ID, rental(t, "2026-11-05", 12))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.AddReservation("2026-001", b.ID, reservation(t, "2026-11-05", 12, "2026-11-05", 14), now); kind(err) != domain.KindConflict {
		t.Errorf("Reservierung im Ausleihzeitraum muss abgelehnt werden, got %v", err)
	}
	if _, err := s.AddReservation("2026-001", b.ID, reservation(t, "2026-11-05", 13, "2026-11-05", 14), now); err != nil {
		t.Errorf("Reservierung direkt nach dem letzten Ausleihblock sollte gehen: %v", err)
	}
	later := time.Date(2026, 11, 5, 13, 0, 0, 0, time.Local) // überfällig
	p, _ := domain.NewReservationPeriod("2026-11-10", 8, "2026-11-10", 9, later)
	if _, err := s.AddReservation("2026-001", b.ID, p, later); kind(err) != domain.KindConflict {
		t.Errorf("überfälliges Fahrrad darf nicht reservierbar sein, got %v", err)
	}
	if _, err := s.EndRental(r.ID, later); err != nil {
		t.Fatal(err)
	}
	if _, err := s.AddReservation("2026-001", b.ID, p, later); err != nil {
		t.Errorf("nach Rückgabe wieder reservierbar: %v", err)
	}
}

func TestNoDoubleReservation(t *testing.T) {
	s, a, b := setup(t)
	if _, err := s.AddReservation("2026-001", a.ID, reservation(t, "2026-11-05", 8, "2026-11-05", 12), now); err != nil {
		t.Fatal(err)
	}
	if _, err := s.AddReservation("2026-001", b.ID, reservation(t, "2026-11-05", 12, "2026-11-05", 15), now); kind(err) != domain.KindConflict {
		t.Errorf("Überschneidung (Bis-Block inklusive) muss abgelehnt werden, got %v", err)
	}
	if _, err := s.AddReservation("2026-001", b.ID, reservation(t, "2026-11-05", 13, "2026-11-05", 15), now); err != nil {
		t.Errorf("direkt anschließend sollte gehen: %v", err)
	}
}

func TestCannotDeleteRentedBike(t *testing.T) {
	s, a, _ := setup(t)
	if _, err := s.StartRental("2026-001", a.ID, rental(t, "2026-11-04", 10)); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteBike("2026-001"); kind(err) != domain.KindConflict {
		t.Errorf("got %v", err)
	}
}

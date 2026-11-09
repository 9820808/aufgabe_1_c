package domain

import (
	"fmt"
	"time"
)

// DisplayFormat ist das Format für Zeitangaben in Fehlermeldungen.
const DisplayFormat = "02.01.2006 15:04"

// Interval ist ein halboffener Zeitraum [Start, End).
// Ausleihen und Reservierungen bestehen immer aus ganzen Stundenblöcken.
type Interval struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Overlaps prüft, ob sich zwei Zeiträume überschneiden.
func (i Interval) Overlaps(o Interval) bool {
	return i.Start.Before(o.End) && o.Start.Before(i.End)
}

// HourBlock schneidet t auf den Beginn der laufenden Stunde ab (Uhrzeit-Stunde).
func HourBlock(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
}

// StartOfDay liefert 00:00 Uhr des Tages von t.
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// ParseDate liest ein Datum im Format JJJJ-MM-TT in lokaler Zeit.
func ParseDate(s string) (time.Time, error) {
	d, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		return time.Time{}, Invalid("ungültiges Datum %q", s)
	}
	return d, nil
}

// LastMinute liefert die letzte Minute vor dem (exklusiven) Ende end,
// z.B. 07:59 für ein Ende um 08:00 – so wird ein Ende angezeigt.
func LastMinute(end time.Time) time.Time {
	return end.Add(-time.Minute)
}

// Block liefert den Beginn des Stundenblocks hour am Tag date (JJJJ-MM-TT).
func Block(date string, hour int) (time.Time, error) {
	if hour < 0 || hour > 23 {
		return time.Time{}, Invalid("Stunde muss zwischen 0 und 23 liegen")
	}
	d, err := ParseDate(date)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(d.Year(), d.Month(), d.Day(), hour, 0, 0, 0, time.Local), nil
}

// NewRentalPeriod bildet den Zeitraum einer Ausleihe: vom laufenden
// Stundenblock bis einschließlich des Rückgabe-Stundenblocks
// (Rückgabestunde 7 = Rückgabe bis spätestens 07:59).
func NewRentalPeriod(returnDate string, returnHour int, now time.Time) (Interval, error) {
	last, err := Block(returnDate, returnHour)
	if err != nil {
		return Interval{}, fmt.Errorf("Rückgabe: %w", err)
	}
	start := HourBlock(now)
	if last.Before(start) {
		return Interval{}, Invalid("Rückgabe darf nicht vor Beginn der Ausleihe (%s) liegen", start.Format(DisplayFormat))
	}
	return Interval{Start: start, End: last.Add(time.Hour)}, nil
}

// NewReservationPeriod bildet den Zeitraum einer Reservierung: vom
// Stundenblock fromHour am Tag from bis einschließlich Stundenblock toHour
// am Tag to. Reservierungen sind frühestens ab morgen möglich.
func NewReservationPeriod(from string, fromHour int, to string, toHour int, now time.Time) (Interval, error) {
	first, err := Block(from, fromHour)
	if err != nil {
		return Interval{}, fmt.Errorf("Von: %w", err)
	}
	last, err := Block(to, toHour)
	if err != nil {
		return Interval{}, fmt.Errorf("Bis: %w", err)
	}
	tomorrow := StartOfDay(now).AddDate(0, 0, 1)
	if first.Before(tomorrow) {
		return Interval{}, Invalid("Reservierungen sind frühestens ab morgen möglich")
	}
	if last.Before(first) {
		return Interval{}, Invalid("Bis liegt vor Von")
	}
	return Interval{Start: first, End: last.Add(time.Hour)}, nil
}

package domain

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// BikeTypes liefert die auswählbaren Fahrradtypen.
func BikeTypes() []string {
	return []string{
		"Rennrad", "Mountain-Bike", "Trekkingrad", "Citybike", "Gravelbike",
		"E-Bike", "E-Mountain-Bike", "Lastenrad", "Kinderrad", "BMX", "Tandem",
	}
}

// SaddleRange ist ein Sattelhöhenbereich in cm.
type SaddleRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

func (r SaddleRange) Validate() error {
	if r.Min <= 0 || r.Max <= 0 {
		return Invalid("Sattelhöhe muss positiv sein")
	}
	if r.Max < r.Min {
		return Invalid("Sattelhöhe: Maximum kleiner als Minimum")
	}
	return nil
}

// Overlaps prüft, ob es eine Höhe gibt, die in beiden Bereichen liegt –
// d.h. ob das Fahrrad auf die Wunschhöhe des Kunden einstellbar ist.
func (r SaddleRange) Overlaps(o SaddleRange) bool {
	return r.Min <= o.Max && o.Min <= r.Max
}

// Bike ist ein Eintrag der Fahrrad-Datenbank.
type Bike struct {
	InventoryNo string      `json:"inventoryNo"`
	Brand       string      `json:"brand"`
	Type        string      `json:"type"`
	Model       string      `json:"model"`
	Saddle      SaddleRange `json:"saddle"`
	WheelInch   float64     `json:"wheelInch"`
}

func (b Bike) Validate() error {
	if _, _, err := ParseInventoryNo(b.InventoryNo); err != nil {
		return err
	}
	if strings.TrimSpace(b.Brand) == "" {
		return Invalid("Marke fehlt")
	}
	if !slices.Contains(BikeTypes(), b.Type) {
		return Invalid("unbekannter Typ %q", b.Type)
	}
	if strings.TrimSpace(b.Model) == "" {
		return Invalid("Modell fehlt")
	}
	if err := b.Saddle.Validate(); err != nil {
		return err
	}
	if b.WheelInch <= 0 {
		return Invalid("Raddurchmesser muss positiv sein")
	}
	return nil
}

// ParseInventoryNo zerlegt eine Inventarnummer der Form <jahr>-<nummer>.
func ParseInventoryNo(s string) (year, num int, err error) {
	y, n, ok := strings.Cut(s, "-")
	year, errY := strconv.Atoi(y)
	num, errN := strconv.Atoi(n)
	if !ok || len(y) != 4 || errY != nil || errN != nil || num <= 0 {
		return 0, 0, Invalid("Inventarnummer %q hat nicht das Format <jahr>-<nummer>", s)
	}
	return year, num, nil
}

// FormatInventoryNo baut eine Inventarnummer, z.B. 2026-007.
func FormatInventoryNo(year, num int) string {
	return fmt.Sprintf("%04d-%03d", year, num)
}

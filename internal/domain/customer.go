package domain

import "strings"

// DocType gibt an, ob die Ausweisnummer ein Führerschein oder ein Pass ist.
type DocType string

const (
	DocNone     DocType = ""
	DocLicense  DocType = "fuehrerschein"
	DocPassport DocType = "pass"
)

// Customer ist ein Eintrag der Kunden-Datenbank.
type Customer struct {
	ID        int         `json:"id"`
	Name      string      `json:"name"`
	Phone     string      `json:"phone"`
	Email     string      `json:"email"`
	Address   string      `json:"address"`
	DocType   DocType     `json:"docType"`
	DocNumber string      `json:"docNumber"`
	Photo     string      `json:"photo,omitempty"` // Data-URL des Ausweisfotos
	Saddle    SaddleRange `json:"saddle"`
	Blocked   bool        `json:"blocked"`
}

func (c Customer) HasPhoto() bool { return c.Photo != "" }

func (c Customer) Validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return Invalid("Name fehlt")
	}
	if c.Photo != "" && !strings.HasPrefix(c.Photo, "data:image/") {
		return Invalid("Ausweisfoto muss ein Bild sein")
	}
	hasNumber := strings.TrimSpace(c.DocNumber) != ""
	if !hasNumber && !c.HasPhoto() {
		return Invalid("Führerschein-/Passnummer oder Ausweisfoto muss angegeben werden")
	}
	if hasNumber && c.DocType != DocLicense && c.DocType != DocPassport {
		return Invalid("bitte auswählen, ob die Nummer ein Führerschein oder ein Pass ist")
	}
	return c.Saddle.Validate()
}

// CanBook prüft, ob der Kunde ausleihen bzw. reservieren darf.
func (c Customer) CanBook() error {
	if c.Blocked {
		return Conflict("Kunde %s ist gesperrt", c.Name)
	}
	return nil
}

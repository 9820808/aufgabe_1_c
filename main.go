package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"fahrradverleih/internal/domain"
	"fahrradverleih/internal/handler"
	"fahrradverleih/internal/store"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP-Adresse")
	demo := flag.Bool("demo", true, "Demodaten anlegen")
	flag.Parse()

	s := store.New()
	if *demo {
		if err := seed(s); err != nil {
			log.Fatalf("Demodaten: %v", err)
		}
	}

	mux := http.NewServeMux()
	handler.New(s).Register(mux)
	mux.Handle("GET /", http.FileServer(http.Dir("web")))

	log.Printf("Fahrradverleih läuft auf http://localhost%s", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

// seed legt ein paar Beispiel-Fahrräder, Kunden und eine Reservierung an.
func seed(s *store.Store) error {
	bikes := []domain.Bike{
		{InventoryNo: "2024-001", Brand: "Cube", Type: "Mountain-Bike", Model: "Aim Pro", Saddle: domain.SaddleRange{Min: 75, Max: 95}, WheelInch: 29},
		{InventoryNo: "2024-002", Brand: "Canyon", Type: "Rennrad", Model: "Endurace 7", Saddle: domain.SaddleRange{Min: 80, Max: 100}, WheelInch: 28},
		{InventoryNo: "2025-001", Brand: "Gazelle", Type: "Citybike", Model: "Esprit", Saddle: domain.SaddleRange{Min: 70, Max: 90}, WheelInch: 28},
		{InventoryNo: "2025-002", Brand: "Riese & Müller", Type: "Lastenrad", Model: "Load 75", Saddle: domain.SaddleRange{Min: 72, Max: 98}, WheelInch: 26},
		{InventoryNo: "2026-001", Brand: "Woom", Type: "Kinderrad", Model: "Woom 4", Saddle: domain.SaddleRange{Min: 50, Max: 65}, WheelInch: 20},
	}
	for _, b := range bikes {
		if _, err := s.AddBike(b); err != nil {
			return err
		}
	}
	customers := []domain.Customer{
		{Name: "Anna Berger", Phone: "+43 660 1234567", Email: "anna@example.org", Address: "Hauptstraße 1, 6020 Innsbruck", DocType: domain.DocLicense, DocNumber: "F1234567", Saddle: domain.SaddleRange{Min: 78, Max: 85}},
		{Name: "Ben Huber", Phone: "+43 664 7654321", Email: "ben@example.org", Address: "Bahnhofweg 5, 6060 Hall", DocType: domain.DocPassport, DocNumber: "P9876543", Saddle: domain.SaddleRange{Min: 88, Max: 94}},
		{Name: "Clara Moser", Phone: "+43 650 1112223", Email: "clara@example.org", Address: "Seeweg 12, 6100 Seefeld", DocType: domain.DocLicense, DocNumber: "F5556667", Saddle: domain.SaddleRange{Min: 55, Max: 60}},
	}
	var ids []int
	for _, c := range customers {
		created, err := s.AddCustomer(c)
		if err != nil {
			return err
		}
		ids = append(ids, created.ID)
	}
	now := time.Now()
	from := now.AddDate(0, 0, 2).Format("2006-01-02")
	to := now.AddDate(0, 0, 3).Format("2006-01-02")
	p, err := domain.NewReservationPeriod(from, 8, to, 17, now) // 08:00 bis 17:59
	if err != nil {
		return err
	}
	_, err = s.AddReservation("2024-001", ids[1], p, now)
	return err
}

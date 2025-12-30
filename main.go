package main

import (
	"log"
	"net/http"
	"os"

	"calsun/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Routes
	http.HandleFunc("/", handlers.WebHandler)
	http.HandleFunc("/calendar.ics", handlers.CalendarHandler)

	log.Printf("CalSun server starting on port %s", port)
	log.Printf("Open http://localhost:%s in your browser", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

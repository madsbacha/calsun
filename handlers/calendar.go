package handlers

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strconv"
	"time"

	ics "github.com/arran4/golang-ical"

	"calsun/services"
)

const (
	defaultDays = 30
	maxDays     = 90
)

// CalendarHandler generates an iCal calendar with sunrise/sunset events
func CalendarHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	latStr := r.URL.Query().Get("lat")
	lngStr := r.URL.Query().Get("lng")
	name := r.URL.Query().Get("name")
	exclude := r.URL.Query().Get("exclude")
	daysStr := r.URL.Query().Get("days")

	// Validate required parameters
	if latStr == "" || lngStr == "" {
		http.Error(w, "lat and lng parameters are required", http.StatusBadRequest)
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil || lat < -90 || lat > 90 {
		http.Error(w, "invalid lat parameter", http.StatusBadRequest)
		return
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil || lng < -180 || lng > 180 {
		http.Error(w, "invalid lng parameter", http.StatusBadRequest)
		return
	}

	// Parse days parameter
	days := defaultDays
	if daysStr != "" {
		days, err = strconv.Atoi(daysStr)
		if err != nil || days < 1 || days > maxDays {
			http.Error(w, fmt.Sprintf("days must be between 1 and %d", maxDays), http.StatusBadRequest)
			return
		}
	}

	// Validate exclude parameter
	includeSunrise := true
	includeSunset := true
	if exclude == "sunrise" {
		includeSunrise = false
	} else if exclude == "sunset" {
		includeSunset = false
	} else if exclude != "" {
		http.Error(w, "exclude must be 'sunrise' or 'sunset'", http.StatusBadRequest)
		return
	}

	// Generate calendar
	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodPublish)
	cal.SetProductId("-//CalSun//Sunrise Sunset Calendar//EN")
	cal.SetName(calendarName(name, includeSunrise, includeSunset))
	cal.SetXWRCalName(calendarName(name, includeSunrise, includeSunset))

	// Get sun times for the date range
	startDate := time.Now().Truncate(24 * time.Hour)
	sunTimes := services.GetSunTimesRange(lat, lng, startDate, days)

	// Location name for descriptions
	locationStr := name
	if locationStr == "" {
		locationStr = fmt.Sprintf("%.4f, %.4f", lat, lng)
	}

	// Add events
	for _, day := range sunTimes {
		if includeSunrise && day.Sunrise != nil {
			event := createSunEvent(day.Sunrise, lat, lng, locationStr)
			cal.AddVEvent(event)
		}
		if includeSunset && day.Sunset != nil {
			event := createSunEvent(day.Sunset, lat, lng, locationStr)
			cal.AddVEvent(event)
		}
	}

	// Set response headers
	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=calsun.ics")

	// Write calendar
	w.Write([]byte(cal.Serialize()))
}

func calendarName(name string, includeSunrise, includeSunset bool) string {
	base := "Sun Times"
	if name != "" {
		base = fmt.Sprintf("Sun Times - %s", name)
	}

	if !includeSunrise {
		return base + " (Sunset only)"
	}
	if !includeSunset {
		return base + " (Sunrise only)"
	}
	return base
}

func createSunEvent(event *services.SunEvent, lat, lng float64, location string) *ics.VEvent {
	// Create a stable UID based on date, location, and event type
	uid := generateUID(event.Time, lat, lng, event.Type)

	e := ics.NewEvent(uid)

	// Set times (1 minute duration)
	e.SetStartAt(event.Time)
	e.SetEndAt(event.Time.Add(1 * time.Minute))

	// Set title with emoji
	if event.Type == "sunrise" {
		e.SetSummary("Sunrise")
	} else {
		e.SetSummary("Sunset")
	}

	// Set description with detailed info
	description := fmt.Sprintf(
		"Time: %s\nLocation: %s\nCoordinates: %.4f, %.4f\nAzimuth: %.1f°",
		event.Time.Format("15:04:05"),
		location,
		lat, lng,
		event.Azimuth,
	)
	e.SetDescription(description)

	// Set location
	e.SetLocation(location)

	return e
}

func generateUID(t time.Time, lat, lng float64, eventType string) string {
	data := fmt.Sprintf("%s-%.4f-%.4f-%s", t.Format("2006-01-02"), lat, lng, eventType)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x@calsun", hash[:8])
}

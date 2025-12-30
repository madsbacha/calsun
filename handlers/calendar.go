package handlers

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"

	"calsun/services"
)

const (
	defaultDays = 30
	maxDays     = 90
)

// calendarParams holds the validated parameters for calendar generation
type calendarParams struct {
	lat            float64
	lng            float64
	name           string
	days           int
	includeSunrise bool
	includeSunset  bool
}

// parseCalendarParams extracts and validates query parameters from the request.
// Returns the parsed params and an error message if validation fails.
func parseCalendarParams(r *http.Request) (*calendarParams, string) {
	q := r.URL.Query()

	// Parse and validate latitude
	latStr := q.Get("lat")
	lngStr := q.Get("lng")
	if latStr == "" || lngStr == "" {
		return nil, "lat and lng parameters are required"
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil || lat < -90 || lat > 90 {
		return nil, "invalid lat parameter"
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil || lng < -180 || lng > 180 {
		return nil, "invalid lng parameter"
	}

	// Parse days parameter with default
	days := defaultDays
	if daysStr := q.Get("days"); daysStr != "" {
		days, err = strconv.Atoi(daysStr)
		if err != nil || days < 1 || days > maxDays {
			return nil, fmt.Sprintf("days must be between 1 and %d", maxDays)
		}
	}

	// Parse exclude parameter
	includeSunrise, includeSunset := true, true
	switch exclude := q.Get("exclude"); exclude {
	case "sunrise":
		includeSunrise = false
	case "sunset":
		includeSunset = false
	case "":
		// No exclusion
	default:
		return nil, "exclude must be 'sunrise' or 'sunset'"
	}

	return &calendarParams{
		lat:            lat,
		lng:            lng,
		name:           q.Get("name"),
		days:           days,
		includeSunrise: includeSunrise,
		includeSunset:  includeSunset,
	}, ""
}

// CalendarHandler generates an iCal calendar with sunrise/sunset events
func CalendarHandler(w http.ResponseWriter, r *http.Request) {
	params, errMsg := parseCalendarParams(r)
	if errMsg != "" {
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	// Generate calendar
	calName := calendarName(params.name, params.includeSunrise, params.includeSunset)
	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodPublish)
	cal.SetProductId("-//CalSun//Sunrise Sunset Calendar//EN")
	cal.SetName(calName)
	cal.SetXWRCalName(calName)

	// Get sun times for the date range
	startDate := time.Now().Truncate(24 * time.Hour)
	sunTimes := services.GetSunTimesRange(params.lat, params.lng, startDate, params.days)

	// Location name for descriptions (use coordinates if no name provided)
	locationStr := params.name
	if locationStr == "" {
		locationStr = fmt.Sprintf("%.4f, %.4f", params.lat, params.lng)
	}

	// Auto-detect timezone from coordinates
	tz := services.GetTimezone(params.lat, params.lng)

	// Add events
	var prevDay *services.DaySunTimes
	for i := range sunTimes {
		day := &sunTimes[i]
		if params.includeSunrise && day.Sunrise != nil {
			cal.AddVEvent(createSunEvent(day.Sunrise, day, prevDay, params.lat, params.lng, locationStr, tz))
		}
		if params.includeSunset && day.Sunset != nil {
			cal.AddVEvent(createSunEvent(day.Sunset, day, prevDay, params.lat, params.lng, locationStr, tz))
		}
		prevDay = day
	}

	// Set response headers and write calendar
	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=calsun.ics")
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

func createSunEvent(event *services.SunEvent, day *services.DaySunTimes, prevDay *services.DaySunTimes, lat, lng float64, location string, tz *time.Location) *ics.VEvent {
	uid := generateUID(event.Time, lat, lng, event.Type)
	e := ics.NewEvent(uid)

	// Set times (1 minute duration)
	e.SetStartAt(event.Time)
	e.SetEndAt(event.Time.Add(time.Minute))

	// Set title with local time (e.g., "Sunrise 06:42")
	localTime := event.Time.In(tz)
	eventTitle := strings.ToUpper(event.Type[:1]) + event.Type[1:]
	e.SetSummary(fmt.Sprintf("%s %s", eventTitle, localTime.Format("15:04")))

	// Build enhanced description
	description := buildDescription(event, day, prevDay, lat, lng, location, tz)
	e.SetDescription(description)
	e.SetLocation(location)

	return e
}

func buildDescription(event *services.SunEvent, day *services.DaySunTimes, prevDay *services.DaySunTimes, lat, lng float64, location string, tz *time.Location) string {
	var lines []string

	// Basic info (show local time)
	localTime := event.Time.In(tz)
	lines = append(lines, fmt.Sprintf("Time: %s", localTime.Format("15:04:05")))
	lines = append(lines, fmt.Sprintf("Location: %s", location))
	lines = append(lines, fmt.Sprintf("Coordinates: %.4f, %.4f", lat, lng))
	lines = append(lines, fmt.Sprintf("Azimuth: %.1f°", event.Azimuth))
	lines = append(lines, "") // blank line

	// Day length (only if both sunrise and sunset exist)
	if day.Sunrise != nil && day.Sunset != nil {
		dayLength := day.Sunset.Time.Sub(day.Sunrise.Time)
		hours := int(dayLength.Hours())
		minutes := int(dayLength.Minutes()) % 60
		lines = append(lines, fmt.Sprintf("Day length: %dh %dm", hours, minutes))
	}

	// Delta from yesterday
	if prevDay != nil {
		var prevEvent *services.SunEvent
		if event.Type == "sunrise" {
			prevEvent = prevDay.Sunrise
		} else {
			prevEvent = prevDay.Sunset
		}

		if prevEvent != nil {
			// Compare times by extracting just hour/minute/second in local timezone
			prevLocalTime := prevEvent.Time.In(tz)
			todaySeconds := localTime.Hour()*3600 + localTime.Minute()*60 + localTime.Second()
			yesterdaySeconds := prevLocalTime.Hour()*3600 + prevLocalTime.Minute()*60 + prevLocalTime.Second()
			deltaSeconds := todaySeconds - yesterdaySeconds
			deltaMinutes := deltaSeconds / 60

			if deltaMinutes != 0 {
				direction := "later"
				if deltaMinutes < 0 {
					direction = "earlier"
					deltaMinutes = -deltaMinutes
				}
				lines = append(lines, fmt.Sprintf("Yesterday: %dm %s", deltaMinutes, direction))
			} else {
				lines = append(lines, "Yesterday: same time")
			}
		}
	}

	// Days until next solstice
	days, solsticeType := services.DaysUntilNextSolstice(event.Time)
	if days == 0 {
		lines = append(lines, fmt.Sprintf("Today is the %s solstice!", solsticeType))
	} else {
		lines = append(lines, fmt.Sprintf("Next solstice: %d days (%s)", days, solsticeType))
	}

	return strings.Join(lines, "\n")
}

func generateUID(t time.Time, lat, lng float64, eventType string) string {
	data := fmt.Sprintf("%s-%.4f-%.4f-%s", t.Format("2006-01-02"), lat, lng, eventType)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x@calsun", hash[:8])
}

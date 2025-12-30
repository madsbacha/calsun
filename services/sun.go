package services

import (
	"math"
	"time"

	"github.com/bradfitz/latlong"
	"github.com/sixdouglas/suncalc"
)

// SunEvent represents a sunrise or sunset event
type SunEvent struct {
	Type      string    // "sunrise" or "sunset"
	Time      time.Time // Exact time of the event
	Azimuth   float64   // Sun's azimuth angle in degrees
	Elevation float64   // Sun's elevation angle in degrees
}

// DaySunTimes holds the sunrise and sunset for a specific day
type DaySunTimes struct {
	Date    time.Time
	Sunrise *SunEvent
	Sunset  *SunEvent
}

// GetSunTimes calculates sunrise and sunset for a given location and date
func GetSunTimes(lat, lng float64, date time.Time) DaySunTimes {
	times := suncalc.GetTimes(date, lat, lng)

	return DaySunTimes{
		Date:    date,
		Sunrise: newSunEvent("sunrise", times[suncalc.Sunrise].Value, lat, lng),
		Sunset:  newSunEvent("sunset", times[suncalc.Sunset].Value, lat, lng),
	}
}

// newSunEvent creates a SunEvent from a time and location, returning nil if the time is zero
func newSunEvent(eventType string, t time.Time, lat, lng float64) *SunEvent {
	if t.IsZero() {
		return nil
	}

	pos := suncalc.GetPosition(t, lat, lng)
	return &SunEvent{
		Type:      eventType,
		Time:      t,
		Azimuth:   radToDeg(pos.Azimuth) + 180, // Convert from [-Pi, Pi] to [0, 360]
		Elevation: radToDeg(pos.Altitude),
	}
}

// GetSunTimesRange calculates sunrise/sunset for a range of days
func GetSunTimesRange(lat, lng float64, startDate time.Time, days int) []DaySunTimes {
	results := make([]DaySunTimes, 0, days)

	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i)
		results = append(results, GetSunTimes(lat, lng, date))
	}

	return results
}

// radToDeg converts radians to degrees
func radToDeg(rad float64) float64 {
	return rad * 180 / math.Pi
}

// DaysUntilNextSolstice calculates the days until the next solstice
// Returns the number of days and the type of solstice ("summer" or "winter")
func DaysUntilNextSolstice(date time.Time) (int, string) {
	year := date.Year()

	// Approximate solstice dates (using UTC)
	summerSolstice := time.Date(year, time.June, 21, 0, 0, 0, 0, time.UTC)
	winterSolstice := time.Date(year, time.December, 21, 0, 0, 0, 0, time.UTC)

	// Normalize date to start of day in UTC for comparison
	dateNorm := time.Date(year, date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	// Calculate days to each solstice
	daysToSummer := int(summerSolstice.Sub(dateNorm).Hours() / 24)
	daysToWinter := int(winterSolstice.Sub(dateNorm).Hours() / 24)

	// If summer solstice has passed this year, use next year's
	if daysToSummer < 0 {
		summerSolstice = time.Date(year+1, time.June, 21, 0, 0, 0, 0, time.UTC)
		daysToSummer = int(summerSolstice.Sub(dateNorm).Hours() / 24)
	}

	// If winter solstice has passed this year, use next year's
	if daysToWinter < 0 {
		winterSolstice = time.Date(year+1, time.December, 21, 0, 0, 0, 0, time.UTC)
		daysToWinter = int(winterSolstice.Sub(dateNorm).Hours() / 24)
	}

	// Return the closest solstice
	if daysToSummer <= daysToWinter {
		return daysToSummer, "summer"
	}
	return daysToWinter, "winter"
}

// GetTimezone returns the timezone for a given latitude and longitude.
// Returns UTC if the timezone cannot be determined.
func GetTimezone(lat, lng float64) *time.Location {
	zoneName := latlong.LookupZoneName(lat, lng)
	if zoneName == "" {
		return time.UTC
	}

	loc, err := time.LoadLocation(zoneName)
	if err != nil {
		return time.UTC
	}

	return loc
}

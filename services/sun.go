package services

import (
	"time"

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
	// Get sun times for the date
	times := suncalc.GetTimes(date, lat, lng)

	result := DaySunTimes{
		Date: date,
	}

	// Get sunrise
	if !times[suncalc.Sunrise].Value.IsZero() {
		sunrisePos := suncalc.GetPosition(times[suncalc.Sunrise].Value, lat, lng)
		result.Sunrise = &SunEvent{
			Type:      "sunrise",
			Time:      times[suncalc.Sunrise].Value,
			Azimuth:   radToDeg(sunrisePos.Azimuth) + 180, // Convert from [-π, π] to [0, 360]
			Elevation: radToDeg(sunrisePos.Altitude),
		}
	}

	// Get sunset
	if !times[suncalc.Sunset].Value.IsZero() {
		sunsetPos := suncalc.GetPosition(times[suncalc.Sunset].Value, lat, lng)
		result.Sunset = &SunEvent{
			Type:      "sunset",
			Time:      times[suncalc.Sunset].Value,
			Azimuth:   radToDeg(sunsetPos.Azimuth) + 180,
			Elevation: radToDeg(sunsetPos.Altitude),
		}
	}

	return result
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
	return rad * 180 / 3.14159265358979323846
}

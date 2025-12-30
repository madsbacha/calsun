package services

import (
	"testing"
	"time"
)

func TestGetSunTimes(t *testing.T) {
	// Test for Copenhagen on a known date
	lat := 55.6761
	lng := 12.5683
	date := time.Date(2024, 6, 21, 12, 0, 0, 0, time.UTC) // Summer solstice

	result := GetSunTimes(lat, lng, date)

	if result.Sunrise == nil {
		t.Fatal("expected sunrise, got nil")
	}
	if result.Sunset == nil {
		t.Fatal("expected sunset, got nil")
	}

	// On summer solstice in Copenhagen, sunrise should be around 2:25 UTC
	// (4:25 local CEST) and sunset around 19:58 UTC (21:58 local CEST)
	if result.Sunrise.Time.Hour() < 2 || result.Sunrise.Time.Hour() > 5 {
		t.Errorf("unexpected sunrise hour: %d", result.Sunrise.Time.Hour())
	}
	if result.Sunset.Time.Hour() < 19 || result.Sunset.Time.Hour() > 22 {
		t.Errorf("unexpected sunset hour: %d", result.Sunset.Time.Hour())
	}

	// Check that sunrise is before sunset
	if !result.Sunrise.Time.Before(result.Sunset.Time) {
		t.Error("sunrise should be before sunset")
	}

	// Check event types
	if result.Sunrise.Type != "sunrise" {
		t.Errorf("expected type 'sunrise', got '%s'", result.Sunrise.Type)
	}
	if result.Sunset.Type != "sunset" {
		t.Errorf("expected type 'sunset', got '%s'", result.Sunset.Type)
	}
}

func TestGetSunTimesRange(t *testing.T) {
	lat := 55.6761
	lng := 12.5683
	startDate := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	days := 7

	results := GetSunTimesRange(lat, lng, startDate, days)

	if len(results) != days {
		t.Errorf("expected %d days, got %d", days, len(results))
	}

	// Each day should have sunrise and sunset
	for i, day := range results {
		if day.Sunrise == nil {
			t.Errorf("day %d: expected sunrise", i)
		}
		if day.Sunset == nil {
			t.Errorf("day %d: expected sunset", i)
		}
	}
}

func TestGetSunTimesAzimuth(t *testing.T) {
	lat := 55.6761
	lng := 12.5683
	date := time.Date(2024, 3, 20, 12, 0, 0, 0, time.UTC) // Spring equinox

	result := GetSunTimes(lat, lng, date)

	// On equinox, sun rises roughly in the east (~90°) and sets in the west (~270°)
	if result.Sunrise.Azimuth < 70 || result.Sunrise.Azimuth > 110 {
		t.Errorf("expected sunrise azimuth near 90°, got %.1f°", result.Sunrise.Azimuth)
	}
	if result.Sunset.Azimuth < 250 || result.Sunset.Azimuth > 290 {
		t.Errorf("expected sunset azimuth near 270°, got %.1f°", result.Sunset.Azimuth)
	}
}

func TestDaysUntilNextSolstice(t *testing.T) {
	tests := []struct {
		name             string
		date             time.Time
		expectedDays     int
		expectedSolstice string
	}{
		{
			name:             "January - winter solstice next",
			date:             time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			expectedDays:     172, // Days to June 21
			expectedSolstice: "summer",
		},
		{
			name:             "March - summer solstice closer",
			date:             time.Date(2024, 3, 21, 0, 0, 0, 0, time.UTC),
			expectedDays:     92, // Days to June 21
			expectedSolstice: "summer",
		},
		{
			name:             "July - winter solstice next",
			date:             time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC),
			expectedDays:     173, // Days to Dec 21
			expectedSolstice: "winter",
		},
		{
			name:             "October - winter solstice closer",
			date:             time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC),
			expectedDays:     81, // Days to Dec 21
			expectedSolstice: "winter",
		},
		{
			name:             "Summer solstice day",
			date:             time.Date(2024, 6, 21, 0, 0, 0, 0, time.UTC),
			expectedDays:     0,
			expectedSolstice: "summer",
		},
		{
			name:             "Winter solstice day",
			date:             time.Date(2024, 12, 21, 0, 0, 0, 0, time.UTC),
			expectedDays:     0,
			expectedSolstice: "winter",
		},
		{
			name:             "Day after winter solstice - summer next",
			date:             time.Date(2024, 12, 22, 0, 0, 0, 0, time.UTC),
			expectedDays:     181, // Days to June 21, 2025
			expectedSolstice: "summer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			days, solstice := DaysUntilNextSolstice(tt.date)
			if days != tt.expectedDays {
				t.Errorf("expected %d days, got %d", tt.expectedDays, days)
			}
			if solstice != tt.expectedSolstice {
				t.Errorf("expected %s solstice, got %s", tt.expectedSolstice, solstice)
			}
		})
	}
}

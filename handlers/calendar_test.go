package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCalendarHandler_ValidRequest(t *testing.T) {
	req := httptest.NewRequest("GET", "/calendar.ics?lat=55.6761&lng=12.5683&name=Copenhagen", nil)
	w := httptest.NewRecorder()

	CalendarHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/calendar") {
		t.Errorf("expected Content-Type text/calendar, got %s", contentType)
	}

	body := w.Body.String()

	// Check for valid iCal structure
	if !strings.HasPrefix(body, "BEGIN:VCALENDAR") {
		t.Error("response should start with BEGIN:VCALENDAR")
	}
	if !strings.Contains(body, "END:VCALENDAR") {
		t.Error("response should contain END:VCALENDAR")
	}
	if !strings.Contains(body, "BEGIN:VEVENT") {
		t.Error("response should contain events")
	}
	if !strings.Contains(body, "SUMMARY:Sunrise") {
		t.Error("response should contain sunrise events")
	}
	if !strings.Contains(body, "SUMMARY:Sunset") {
		t.Error("response should contain sunset events")
	}
}

func TestCalendarHandler_MissingParams(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"missing lat", "/calendar.ics?lng=12.5683"},
		{"missing lng", "/calendar.ics?lat=55.6761"},
		{"missing both", "/calendar.ics"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()

			CalendarHandler(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected status 400, got %d", w.Code)
			}
		})
	}
}

func TestCalendarHandler_InvalidParams(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"invalid lat", "/calendar.ics?lat=invalid&lng=12.5683"},
		{"invalid lng", "/calendar.ics?lat=55.6761&lng=invalid"},
		{"lat out of range", "/calendar.ics?lat=100&lng=12.5683"},
		{"lng out of range", "/calendar.ics?lat=55.6761&lng=200"},
		{"invalid days", "/calendar.ics?lat=55.6761&lng=12.5683&days=abc"},
		{"days too high", "/calendar.ics?lat=55.6761&lng=12.5683&days=100"},
		{"invalid exclude", "/calendar.ics?lat=55.6761&lng=12.5683&exclude=invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()

			CalendarHandler(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected status 400, got %d", w.Code)
			}
		})
	}
}

func TestCalendarHandler_ExcludeSunrise(t *testing.T) {
	req := httptest.NewRequest("GET", "/calendar.ics?lat=55.6761&lng=12.5683&exclude=sunrise", nil)
	w := httptest.NewRecorder()

	CalendarHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	body := w.Body.String()

	if strings.Contains(body, "SUMMARY:Sunrise") {
		t.Error("response should not contain sunrise events when excluded")
	}
	if !strings.Contains(body, "SUMMARY:Sunset") {
		t.Error("response should contain sunset events")
	}
}

func TestCalendarHandler_ExcludeSunset(t *testing.T) {
	req := httptest.NewRequest("GET", "/calendar.ics?lat=55.6761&lng=12.5683&exclude=sunset", nil)
	w := httptest.NewRecorder()

	CalendarHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	body := w.Body.String()

	if !strings.Contains(body, "SUMMARY:Sunrise") {
		t.Error("response should contain sunrise events")
	}
	if strings.Contains(body, "SUMMARY:Sunset") {
		t.Error("response should not contain sunset events when excluded")
	}
}

func TestCalendarHandler_CustomDays(t *testing.T) {
	req := httptest.NewRequest("GET", "/calendar.ics?lat=55.6761&lng=12.5683&days=7", nil)
	w := httptest.NewRecorder()

	CalendarHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	body := w.Body.String()

	// Count events (should be ~14: 7 sunrises + 7 sunsets)
	eventCount := strings.Count(body, "BEGIN:VEVENT")
	if eventCount < 12 || eventCount > 16 {
		t.Errorf("expected ~14 events for 7 days, got %d", eventCount)
	}
}

func TestCalendarHandler_CalendarName(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		expectedName string
	}{
		{"with name", "/calendar.ics?lat=55.6761&lng=12.5683&name=Copenhagen", "Sun Times - Copenhagen"},
		{"without name", "/calendar.ics?lat=55.6761&lng=12.5683", "Sun Times"},
		{"sunset only with name", "/calendar.ics?lat=55.6761&lng=12.5683&name=Test&exclude=sunrise", "Sun Times - Test (Sunset only)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()

			CalendarHandler(w, req)

			body := w.Body.String()
			if !strings.Contains(body, tt.expectedName) {
				t.Errorf("expected calendar name to contain '%s'", tt.expectedName)
			}
		})
	}
}

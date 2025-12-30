package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWebHandler_ServesHTML(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	WebHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("expected Content-Type text/html, got %s", contentType)
	}

	body := w.Body.String()

	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("response should be HTML document")
	}
	if !strings.Contains(body, "CalSun") {
		t.Error("response should contain CalSun title")
	}
}

func TestWebHandler_NotFoundForOtherPaths(t *testing.T) {
	paths := []string{"/foo", "/bar", "/calendar", "/index.html"}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()

			WebHandler(w, req)

			if w.Code != http.StatusNotFound {
				t.Errorf("expected status 404 for %s, got %d", path, w.Code)
			}
		})
	}
}

func TestWebHandler_ContainsRequiredElements(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	WebHandler(w, req)

	body := w.Body.String()

	requiredElements := []string{
		`id="address"`,        // Location input
		`id="calForm"`,        // Form
		`id="result"`,         // Result section
		`id="copyBtn"`,        // Copy button
		`id="subscribeBtn"`,   // Subscribe button
		`name="events"`,       // Radio buttons
		`nominatim`,           // Geocoding reference
	}

	for _, elem := range requiredElements {
		if !strings.Contains(body, elem) {
			t.Errorf("response should contain '%s'", elem)
		}
	}
}

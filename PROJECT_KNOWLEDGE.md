# CalSun - Project Knowledge

## Overview

CalSun is a web service that provides iCal calendar subscriptions for sunrise and sunset times based on geographic location. Users can subscribe to these calendars from their iPhone, Google Calendar, or any calendar app that supports iCal subscriptions.

## Architecture

### Tech Stack
- **Backend**: Go (standard library HTTP server)
- **Frontend**: Single HTML page with vanilla JavaScript
- **Development**: Nix flake for reproducible dev environment
- **Deployment**: Docker container
- **External APIs**: OpenStreetMap Nominatim (client-side only, for geocoding)

### Project Structure
```
calsun/
├── main.go              # Entry point, HTTP server setup
├── handlers/
│   ├── calendar.go      # iCal generation endpoint
│   ├── web.go           # Serve the web UI
│   └── templates/
│       └── index.html   # Single-page web UI (embedded)
├── services/
│   └── sun.go           # Sunrise/sunset calculations
├── flake.nix            # Nix flake for dev environment
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── go.sum
```

### Key Design Decisions

1. **Hybrid Geocoding**: The web UI performs geocoding (address to coordinates) client-side using OpenStreetMap Nominatim. The subscription URL then contains coordinates directly. This means:
   - No external API dependency on the server
   - Calendar refreshes are fast (no geocoding needed)
   - URLs are stable and don't depend on geocoding service availability

2. **Short Event Duration**: Sunrise/sunset events are 1 minute long, just marking the moment. This keeps the calendar clean and non-intrusive.

3. **30-Day Rolling Window**: By default, the calendar generates events for 30 days ahead. This keeps the file small and response fast while still being useful.

4. **Local Calculation**: Sun times are calculated locally using astronomical algorithms, not fetched from an external API. This eliminates rate limits and external dependencies.

## API Reference

### `GET /`
Serves the web UI.

### `GET /calendar.ics`
Returns an iCal calendar file.

**Query Parameters:**
| Parameter | Required | Description |
|-----------|----------|-------------|
| `lat` | Yes | Latitude (-90 to 90) |
| `lng` | Yes | Longitude (-180 to 180) |
| `name` | No | Location name (shown in event details) |
| `exclude` | No | `"sunrise"` or `"sunset"` to exclude one type |
| `days` | No | Days ahead to generate (default: 30, max: 90) |

**Example:**
```
/calendar.ics?lat=55.6761&lng=12.5683&name=Copenhagen&exclude=sunset
```

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/sixdouglas/suncalc` | Astronomical calculations for sun times |
| `github.com/arran4/golang-ical` | RFC 5545 compliant iCal generation |

## Running Locally

```bash
# Using Nix (recommended for development)
nix develop --command go run .

# Using Docker
docker-compose up

# Direct (if Go is installed)
go run .
```

The server runs on port 8080 by default. Open http://localhost:8080 in your browser.

## Timezone & DST Handling

All times are calculated and stored in **UTC**:

1. **suncalc library** calculates sunrise/sunset times in UTC based on coordinates and date
2. **iCal output** uses UTC format with `Z` suffix (e.g., `DTSTART:20251230T074027Z`)
3. **Calendar apps** convert UTC to the user's local timezone for display

This approach means:
- We don't need to know the user's timezone
- DST is handled automatically by the user's calendar app
- Times are always unambiguous (no DST transition edge cases)

Example: Sunrise at `07:40:27Z` (UTC) displays as:
- 08:40:27 in Copenhagen (winter, UTC+1)
- 09:40:27 in Copenhagen (summer, UTC+2)

## Calendar Subscription Notes

- iOS/macOS calendar apps refresh subscriptions automatically (typically every few hours)
- The `webcal://` protocol triggers the native "Add to Calendar" flow on iOS

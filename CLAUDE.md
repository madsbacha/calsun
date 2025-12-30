# Claude Instructions for CalSun

## Project Context

This is a Go web service for generating iCal calendar subscriptions with sunrise/sunset times. See `PROJECT_KNOWLEDGE.md` for full architecture and API documentation.

## Code Conventions

### Go
- Use standard library where possible (net/http, html/template)
- Error handling: return errors, don't panic
- Use `log` package for logging
- Keep handlers thin, business logic in services

### File Organization
- `handlers/` - HTTP handlers only, minimal logic
- `services/` - Business logic and calculations
- `templates/` - HTML templates
- `static/` - CSS, JS, images

## Testing

Run tests with:
```bash
nix develop --command go test ./...

# Verbose output
nix develop --command go test ./... -v
```

Test coverage:
- `services/sun_test.go` - Sun calculation tests (times, azimuth, ranges)
- `handlers/calendar_test.go` - Calendar endpoint tests (params, iCal format, exclusions)
- `handlers/web_test.go` - Web UI handler tests (HTML response, 404 handling)

## Common Tasks

### Adding a new endpoint
1. Create handler in `handlers/`
2. Register route in `main.go`

### Modifying calendar output
- Calendar generation logic is in `handlers/calendar.go`
- iCal format uses `github.com/arran4/golang-ical`

### Changing sun calculations
- All astronomy logic is in `services/sun.go`
- Uses `github.com/sixdouglas/suncalc` library

## Important Notes

- Calendar endpoint must return `Content-Type: text/calendar; charset=utf-8`
- Geocoding happens client-side only (in the web UI JavaScript)
- Event UIDs must be stable across requests (based on date + location + type)

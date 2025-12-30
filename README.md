# CalSun

Subscribe to sunrise and sunset times in your calendar.

CalSun is a web service that generates iCal calendar subscriptions for sunrise and sunset times based on your location. Works with iPhone, Google Calendar, Outlook, and any calendar app that supports iCal subscriptions.

## Usage

1. Open the web interface
2. Enter your location (city, address, or coordinates)
3. Choose which events to include (sunrise, sunset, or both)
4. Click "Add to Calendar" or copy the subscription URL

Your calendar will automatically update with sunrise/sunset times for the next 30 days.

## Running

### Docker

```bash
docker-compose up
```

### From source

```bash
# Using Nix
nix develop --command go run .

# Or with Go installed
go run .
```

Open http://localhost:8080

## API

### `GET /calendar.ics`

Returns an iCal calendar file.

| Parameter | Required | Description |
|-----------|----------|-------------|
| `lat` | Yes | Latitude (-90 to 90) |
| `lng` | Yes | Longitude (-180 to 180) |
| `name` | No | Location name for event details |
| `exclude` | No | `sunrise` or `sunset` to exclude one |
| `days` | No | Days ahead (default: 30, max: 90) |

Example:
```
/calendar.ics?lat=55.6761&lng=12.5683&name=Copenhagen
```

## Development

```bash
# Enter dev environment
nix develop

# Run tests
go test ./...

# Build
go build -o calsun .
```

## License

MIT

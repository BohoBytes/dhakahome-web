# Mock Mode Documentation

## Overview

The application includes a comprehensive mock system that allows you to run the entire application without connecting to a real backend API. This is useful for:

- Frontend development without backend access
- Testing UI changes quickly
- Demos and presentations
- CI/CD environments where backend is not available
- Development when backend is down or under maintenance

## How to Enable Mock Mode

### Method 1: Environment Variable (Recommended)

Set the `MOCK_ENABLED` environment variable to `true`:

```bash
# In your .env or .env.local file
MOCK_ENABLED=true
```

Then start your application normally:

```bash
go run ./cmd/web
```

### Method 2: Inline Environment Variable

Start the application with the environment variable inline:

```bash
MOCK_ENABLED=true go run ./cmd/web
```

### Method 3: Export Environment Variable

Export the variable in your shell session:

```bash
export MOCK_ENABLED=true
go run ./cmd/web
```

## Supported Values

The following values will enable mock mode (case-insensitive):
- `true`
- `1`
- `yes`

Any other value (including empty string) will disable mock mode and use the real API.

## Mock Data

When mock mode is enabled, the application serves a curated dataset of **25 properties** that mirrors the live API shape.

### Coverage
- Residential, commercial, hostel, and short-term listings
- Neighborhoods across Dhaka (Uttara, Gulshan, Banani, Dhanmondi, Mirpur, Bashundhara, Mohammadpur, etc.)
- Rentals and sales spanning budget to luxury price bands
- Images and metadata shaped to match the live card/detail templates

## Supported Features

All API operations work seamlessly in mock mode:

### 1. Property Search (`SearchProperties`)
- Full text search (title, address, badges)
- Location filtering
- Area/Neighborhood filtering
- Property type filtering (Residential, Commercial, etc.)
- Status filtering (To-let, For Sale)
- Price range filtering (min/max)
- Bedroom/Bathroom filtering
- Furnished status filtering
- **Pagination support** (page & limit parameters)

### 2. Get Single Property (`GetProperty`)
- Retrieve any property by ID
- Returns complete property details

### 3. Lead Submission (`SubmitLead`)
- Accepts lead submissions
- Logs to console for verification
- Always returns success

## Testing Examples

Use a browser for the full experience; the `curl` calls below return HTML you can skim for matching cards/headings.

1) Default search page  
```bash
curl -I "http://localhost:5173/search"
```

2) City + area filter  
```bash
curl "http://localhost:5173/search?city=Dhaka&neighborhood=Gulshan"
```

3) Commercial listings  
```bash
curl "http://localhost:5173/search?type=Commercial"
```

4) Price band  
```bash
curl "http://localhost:5173/search?price_min=20000&price_max=60000"
```

5) Bedrooms + listing type  
```bash
curl "http://localhost:5173/search?bedrooms=3&listing_type=listed_rental"
```

6) Pagination + ordering  
```bash
curl "http://localhost:5173/search?page=2&limit=5&sort_by=price&order=desc"
```

7) Property details  
```bash
curl "http://localhost:5173/properties/mock-res-uttara-01"
```

## Log Output

When mock mode is enabled, you'll see clear indicators in the logs:

```
🎭 API Client: MOCK MODE ENABLED - All API calls will use mock data
🎭 Mock: Searching properties with params: location=Gulshan
🎭 Mock: Found 4 properties after filtering
```

The 🎭 (theater mask) emoji indicates mock operations.

## Implementation Details

### Architecture

The mock system is implemented directly in the API client ([internal/api/client.go](../internal/api/client.go)) using a conditional flag:

1. On initialization, the client checks the `MOCK_ENABLED` environment variable
2. If enabled, the `mockEnabled` flag is set to `true`
3. All API methods check this flag and route to mock implementations
4. Mock data is generated inline with comprehensive filtering logic

### Key Functions

- `New()` - Detects mock mode from environment variable
- `SearchProperties()` - Checks `mockEnabled` flag and routes accordingly
- `GetProperty()` - Checks `mockEnabled` flag and searches mock data
- `SubmitLead()` - Logs mock lead submissions
- `getMockSearchResults()` - Main mock data generator with filtering
- `matchesMockFilters()` - Applies all search filters to mock data
- `getAllMockProperties()` - Returns complete mock dataset

### Filter Support

All standard search parameters are fully supported:
- `q` (or `location`/`area`) for text search seeds
- `city`
- `area` / `neighborhood`
- `types` / `type`
- `listing_type` / `listingType` (rent vs sale)
- `status` (overrides the default `listed_rental,listed_sale`)
- `price_min` / `price_max`
- `bedrooms`, `bathrooms`
- `parking`
- `serviced`, `shared_room`, `furnished`
- `area_min`, `area_max`
- `sort_by` (e.g., `price`) and `order` (`asc`/`desc`)
- `page` (default 1) and `limit` (default 9)

## Benefits

1. **Zero Backend Dependency** - Develop frontend features without backend
2. **Fast Iteration** - No network latency, instant responses
3. **Consistent Data** - Same mock data every time for reliable testing
4. **Full Feature Coverage** - All API endpoints work identically
5. **Easy Toggle** - Switch between mock and real API with one environment variable
6. **Production-Safe** - Mock mode defaults to OFF, must be explicitly enabled

## Limitations

1. **Static Data** - Mock data doesn't change unless code is updated
2. **No Persistence** - Lead submissions are logged but not saved
3. **Simplified Logic** - Some complex backend behaviors may not be replicated
4. **Fixed Dataset** - Always returns the same dataset (currently 25 properties)

## Extending Mock Data

To add more mock properties, edit the `getAllMockProperties()` function in [internal/api/client.go](../internal/api/client.go):

```go
func getAllMockProperties() []Property {
    return []Property{
        // Add your new properties here
        {
            ID:        "mock-custom-01",
            Title:     "Your Property Title",
            Address:   "Your Address",
            Price:     50000,
            Currency:  "৳",
            Images:    []string{"/assets/your-image.png"},
            Badges:    []string{"To-let", "Verified", "Residential"},
            Bedrooms:  3,
            Bathrooms: 2,
            Area:      1500,
            Parking:   1,
        },
        // ... existing properties
    }
}
```

## Troubleshooting

### Mock Mode Not Working

1. **Check environment variable**:
   ```bash
   echo $MOCK_ENABLED
   ```

2. **Check .env file**:
   ```bash
   cat .env | grep MOCK_ENABLED
   cat .env.local | grep MOCK_ENABLED
   ```

3. **Check logs for mock indicator**:
   Look for `🎭 API Client: MOCK MODE ENABLED` in startup logs

### Still Calling Real API

If you see real API calls despite setting `MOCK_ENABLED=true`:
- Ensure the environment variable is set BEFORE starting the app
- Restart the application after changing .env files
- Check for typos in the variable name (it's case-sensitive)

### No Properties Returned

If mock search returns no results:
- Check your filter parameters - they may be too restrictive
- Review the mock dataset to ensure it matches your filters
- Check logs for "Found X properties after filtering" message

## Production Usage

**IMPORTANT**: Mock mode should NEVER be enabled in production. Always ensure:

```bash
# Production .env
MOCK_ENABLED=false
```

Or simply omit the variable entirely (defaults to false).

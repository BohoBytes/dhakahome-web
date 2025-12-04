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
go run cmd/web/main.go
```

### Method 2: Inline Environment Variable

Start the application with the environment variable inline:

```bash
MOCK_ENABLED=true go run cmd/web/main.go
```

### Method 3: Export Environment Variable

Export the variable in your shell session:

```bash
export MOCK_ENABLED=true
go run cmd/web/main.go
```

## Supported Values

The following values will enable mock mode (case-insensitive):
- `true`
- `1`
- `yes`

Any other value (including empty string) will disable mock mode and use the real API.

## Mock Data

When mock mode is enabled, the application uses a comprehensive dataset of **23 properties** including:

### Property Types
- **Residential Properties** (17 properties)
  - Apartments in Uttara, Gulshan, Banani, Dhanmondi, Mirpur, Bashundhara, Mohammadpur
  - Ranging from budget-friendly to luxury
  - Studio to 6-bedroom options

- **Commercial Properties** (4 properties)
  - Office spaces
  - Retail shops
  - Various locations and sizes

- **Hostels** (2 properties)
  - Student hostels
  - Professional hostels
  - Shared accommodations

- **Short Term Rentals** (2 properties)
  - Serviced apartments
  - Daily/monthly rentals

### Areas Covered
- Uttara (North & South)
- Gulshan (1 & 2)
- Banani (including DOHS)
- Dhanmondi
- Mirpur (10 & 11)
- Bashundhara R/A
- Mohammadpur

### Price Range
- Budget: ৳3,500 - ৳22,000/month
- Mid-range: ৳35,000 - ৳75,000/month
- Luxury: ৳95,000 - ৳250,000/month
- Sale: ৳3.5M - ৳25M

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

### 1. Basic Search (All Properties)
```bash
curl "http://localhost:5173/"
# Returns: First 9 properties (default pagination)
```

### 2. Search by Location
```bash
curl "http://localhost:5173/search?location=Gulshan"
# Returns: 4 properties in Gulshan area
```

### 3. Search by Property Type
```bash
curl "http://localhost:5173/search?type=Commercial"
# Returns: 4 commercial properties
```

### 4. Search with Multiple Filters
```bash
curl "http://localhost:5173/search?location=Uttara&type=Residential&price_max=50000"
# Returns: Residential properties in Uttara under ৳50,000
```

### 5. Pagination
```bash
# Page 1 with 5 items per page
curl "http://localhost:5173/search?page=1&limit=5"

# Page 2 with 5 items per page
curl "http://localhost:5173/search?page=2&limit=5"
```

### 6. Price Range
```bash
curl "http://localhost:5173/search?price_min=20000&price_max=60000"
# Returns: Properties between ৳20,000 and ৳60,000
```

### 7. Bedroom Filter
```bash
curl "http://localhost:5173/search?bedrooms=3"
# Returns: Properties with 3 or more bedrooms
```

### 8. Furnished Properties
```bash
curl "http://localhost:5173/search?furnished=yes"
# Returns: Only furnished properties
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
- `q` - Text search
- `location` - Location filter
- `area` / `neighborhood` - Area filter
- `types` / `type` - Property type filter
- `status` - Status filter (ready_for_listing, active, for_sale, etc.)
- `price_min` / `price_max` - Price range
- `bedrooms` - Minimum bedrooms
- `bathrooms` - Minimum bathrooms
- `furnished` - Furnished status (yes/no)
- `page` - Page number (default: 1)
- `limit` - Items per page (default: 9)

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
4. **Fixed Dataset** - Always returns the same 23 properties

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

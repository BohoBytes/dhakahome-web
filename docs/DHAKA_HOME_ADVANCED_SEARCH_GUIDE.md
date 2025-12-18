# Dhaka Home Site - Advanced Search Feature Guide

**Version:** 2.0 (Updated with Advanced Search Filters)
**Date:** December 16, 2024
**Status:** ✅ Production Ready

---

## Table of Contents

1. [Feature Overview](#feature-overview)
2. [API Endpoints](#api-endpoints)
3. [Advanced Search Filters](#advanced-search-filters)
4. [Implementation Guides](#implementation-guides)
5. [API Examples](#api-examples)
6. [Frontend Integration](#frontend-integration)
7. [Data Structures](#data-structures)
8. [Troubleshooting](#troubleshooting)

---

## Feature Overview

The advanced search feature provides comprehensive filtering capabilities for property listings on the Dhaka Home site. Users can filter by multiple criteria including:

- **Basic Filters:** City, Area/Neighborhood, Property Type
- **Property Characteristics:** Bedrooms, Bathrooms, Square Footage, Parking Spaces
- **Amenities & Features:** Furnished Status, Serviced Status, Shared Rooms
- **Pricing:** Dynamic price ranges based on current results
- **Geographic:** Bounding box and radius-based searches

### What's New (v2.0)

✨ **New Filter Options:**
- **Parking Spaces** - Support for 0, 1, 2, 3+ parking spaces
- **Serviced Apartments** - Filter residential properties by serviced status
- **Shared Rooms** - Filter hostel properties by shared vs private rooms

✨ **Smart Price Slider**
- Dynamic min/max based on current search results
- Frontend calculates bounds from search results
- Real-time updates as filters change

---

## API Endpoints

### Core Search Endpoints

#### 1. **List Properties with Filters**
```
GET /api/v1/assets
```

**Purpose:** Search and filter properties with multiple criteria

**Required Parameters:**
- `status` - Property listing status (required for public listings)
  - Values: `listed_rental,listed_sale`

**Optional Parameters:**

| Parameter | Type | Example | Description |
|-----------|------|---------|-------------|
| `city` | string | `Dhaka` | Filter by city |
| `neighborhood` | string | `Gulshan` | Filter by area/neighborhood |
| `types` | string | `Apartment,Hostel` | Comma-separated property types |
| `bedrooms` | integer | `3` | Minimum bedrooms |
| `bathrooms` | integer | `2` | Minimum bathrooms |
| `price_min` | float | `50000` | Minimum price |
| `price_max` | float | `150000` | Maximum price |
| `parking` | integer | `2` | Parking spaces (0, 1, 2, 3+) |
| `serviced` | boolean | `true` | Serviced apartment (for residential) |
| `shared_room` | boolean | `false` | Shared room (for hostels) |
| `furnished` | boolean | `true` | Furnished status |
| `limit` | integer | `20` | Results per page (max 100) |
| `page` | integer | `1` | Page number |

**Response:** Array of property objects with pricing, photos, and details

---

#### 2. **Get Cities Dropdown**
```
GET /api/v1/assets/cities
```

**Parameters:**
- `status` - optional, filter cities by listing status

**Response:** Array of city names available

---

#### 3. **Get Neighborhoods**
```
GET /api/v1/assets/neighborhoods
```

**Parameters:**
- `city` - required, filter by specific city
- `status` - optional

**Response:** Array of neighborhood names for the city

---

#### 4. **Get Property Type Configuration**
```
GET /api/v1/config/property-types
```

**Parameters:**
- `listingType` - optional, `rental` or `sale`

**Response:** Property type hierarchy with features

---

#### 5. **Get Property Details**
```
GET /api/v1/assets/{id}
```

**Response:** Complete property information including photos, documents, pricing, amenities

---

#### 6. **Get Similar Properties**
```
GET /api/v1/assets/{id}/similar
```

**Parameters:**
- `limit` - number of similar properties (default 6, max 10)

**Response:** Array of similar properties

---

## Advanced Search Filters

### Parking Spaces Filter

**Field:** `parking`
**Type:** Integer
**Supported Values:**
- `0` - No parking
- `1` - 1 parking space
- `2` - 2 parking spaces
- `3` - 3 or more parking spaces

**Example:**
```bash
# Apartments with 2 parking spaces
GET /api/v1/assets?status=listed_rental&types=Apartment&parking=2

# Commercial spaces with 3+ parking
GET /api/v1/assets?status=listed_sale&types=Office&parking=3
```

**UI Implementation:**
```html
<select name="parking">
  <option value="">Any</option>
  <option value="0">No Parking</option>
  <option value="1">1 Space</option>
  <option value="2">2 Spaces</option>
  <option value="3">3 + Spaces</option>
</select>
```

---

### Serviced Apartment Filter

**Field:** `serviced`
**Type:** Boolean
**Applies to:** Residential properties (Apartment, House, Villa)

**Values:**
- `true` - Serviced apartments
- `false` - Non-serviced apartments
- omit - Any (default)

**Example:**
```bash
# Serviced apartments in Gulshan
GET /api/v1/assets?status=listed_rental&types=Apartment&neighborhood=Gulshan&serviced=true

# Non-serviced apartments
GET /api/v1/assets?status=listed_rental&serviced=false
```

**UI Implementation:**
```html
<!-- Show only for Apartment/House types -->
<label>
  <input type="checkbox" name="serviced" value="true">
  Serviced Apartment
</label>
```

---

### Shared Room Filter

**Field:** `shared_room`
**Type:** Boolean
**Applies to:** Hostel properties only

**Values:**
- `true` - Shared rooms (multiple beds)
- `false` - Private rooms
- omit - Any (default)

**Example:**
```bash
# Shared room hostels
GET /api/v1/assets?status=listed_rental&types=Hostel&shared_room=true

# Private room hostels
GET /api/v1/assets?status=listed_rental&types=Hostel&shared_room=false
```

**UI Implementation:**
```html
<!-- Show only for Hostel type -->
<label>
  <input type="checkbox" name="shared_room" value="true">
  Shared Room
</label>
```

---

### Price Range Slider (Smart/Dynamic)

**Implementation Strategy:** Frontend-Computed

**Flow:**
1. User enters search criteria and clicks search
2. API returns all matching properties with prices
3. Frontend extracts all prices from results
4. Frontend calculates min and max price
5. Frontend rounds max to nearest 10,000
6. Set slider range: [minPrice, roundedMaxPrice]
7. User adjusts slider and searches again

**Frontend Pseudocode:**
```javascript
function setDynamicPriceSlider(searchResults) {
  const prices = searchResults
    .filter(item => item.details?.pricing?.monthly_rent || item.details?.pricing?.sale_price)
    .map(item => item.details.pricing.monthly_rent || item.details.pricing.sale_price);

  if (prices.length === 0) {
    return; // No prices available
  }

  const minPrice = Math.min(...prices);
  const maxPrice = Math.max(...prices);

  // Round max up to nearest 10,000 for nice slider ticks
  const roundedMax = Math.ceil(maxPrice / 10000) * 10000;

  setSliderRange(minPrice, roundedMax);
}
```

---

## Implementation Guides

### Step 1: Basic Search Form Setup

```html
<form id="searchForm">
  <!-- Basic Filters -->
  <select name="city" id="citySelect">
    <option value="">Select City</option>
  </select>

  <select name="neighborhood" id="neighborhoodSelect">
    <option value="">Select Neighborhood</option>
  </select>

  <select name="types" id="typeSelect">
    <option value="">Select Property Type</option>
  </select>

  <!-- Advanced Search Toggle -->
  <button type="button" id="advancedSearchBtn">
    <i class="icon-filter"></i> Advanced Search
  </button>

  <button type="submit">Search</button>
  <button type="button" id="clearBtn">Clear All</button>
</form>
```

---

### Step 2: Advanced Search Section

```html
<div id="advancedSearchSection" style="display: none;">
  <h3>Advanced Filters</h3>

  <!-- Listing Type -->
  <div>
    <label>Listing Type</label>
    <select name="listingType">
      <option value="listed_rental">Rent</option>
      <option value="listed_sale">Sale</option>
      <option value="">Both</option>
    </select>
  </div>

  <!-- Bedrooms -->
  <div>
    <label>Bedrooms</label>
    <select name="bedrooms">
      <option value="">Any</option>
      <option value="1">1</option>
      <option value="2">2</option>
      <option value="3">3</option>
      <option value="4">4</option>
      <option value="5">5+</option>
    </select>
  </div>

  <!-- Bathrooms -->
  <div>
    <label>Bathrooms</label>
    <select name="bathrooms">
      <option value="">Any</option>
      <option value="1">1</option>
      <option value="2">2</option>
      <option value="3">3</option>
      <option value="4">4+</option>
    </select>
  </div>

  <!-- Parking -->
  <div>
    <label>Parking Spaces</label>
    <select name="parking">
      <option value="">Any</option>
      <option value="0">No Parking</option>
      <option value="1">1 Space</option>
      <option value="2">2 Spaces</option>
      <option value="3">3+ Spaces</option>
    </select>
  </div>

  <!-- Serviced (for residential) -->
  <div id="servicedSection" style="display: none;">
    <label>
      <input type="checkbox" name="serviced" value="true">
      Serviced Apartment
    </label>
  </div>

  <!-- Shared Room (for hostel) -->
  <div id="sharedRoomSection" style="display: none;">
    <label>
      <input type="checkbox" name="shared_room" value="true">
      Shared Room
    </label>
  </div>

  <!-- Price Range Slider -->
  <div>
    <label>Price Range</label>
    <input type="range" id="priceSlider" min="0" max="200000" step="1000">
    <div>
      <span id="priceDisplay">0 - 200,000</span>
    </div>
  </div>

  <!-- Furnished -->
  <div>
    <label>
      <input type="checkbox" name="furnished" value="true">
      Furnished
    </label>
  </div>
</div>
```

---

### Step 3: JavaScript Implementation

```javascript
// Load cities on page load
document.addEventListener('DOMContentLoaded', async () => {
  const cities = await fetch(
    'https://api.nestlo.com/api/v1/assets/cities?status=listed_rental,listed_sale',
    { headers: { 'Authorization': `Bearer ${TOKEN}` } }
  ).then(r => r.json());

  populateDropdown('#citySelect', cities);
});

// Load neighborhoods when city changes
document.getElementById('citySelect').addEventListener('change', async (e) => {
  const city = e.target.value;
  if (!city) return;

  const neighborhoods = await fetch(
    `https://api.nestlo.com/api/v1/assets/neighborhoods?city=${city}&status=listed_rental`,
    { headers: { 'Authorization': `Bearer ${TOKEN}` } }
  ).then(r => r.json());

  populateDropdown('#neighborhoodSelect', neighborhoods);
});

// Show/hide advanced sections based on property type
document.getElementById('typeSelect').addEventListener('change', (e) => {
  const type = e.target.value;

  // Show serviced for residential
  document.getElementById('servicedSection').style.display =
    (type === 'Apartment' || type === 'House') ? 'block' : 'none';

  // Show shared room for hostels
  document.getElementById('sharedRoomSection').style.display =
    (type === 'Hostel') ? 'block' : 'none';
});

// Handle form submission
document.getElementById('searchForm').addEventListener('submit', async (e) => {
  e.preventDefault();
  await performSearch();
});

async function performSearch() {
  const formData = new FormData(document.getElementById('searchForm'));

  // Build query params
  const params = new URLSearchParams();

  // Always include status
  const listingType = formData.get('listingType') || 'listed_rental,listed_sale';
  params.append('status', listingType);

  // Add other filters
  ['city', 'neighborhood', 'types', 'bedrooms', 'bathrooms', 'parking', 'serviced', 'shared_room', 'furnished'].forEach(key => {
    const value = formData.get(key);
    if (value) params.append(key, value);
  });

  params.append('limit', '20');

  // Perform search
  const results = await fetch(
    `https://api.nestlo.com/api/v1/assets?${params}`,
    { headers: { 'Authorization': `Bearer ${TOKEN}` } }
  ).then(r => r.json());

  // Update price slider based on results
  updateDynamicPriceSlider(results.data);

  // Display results
  displayResults(results.data);
}

function updateDynamicPriceSlider(results) {
  const prices = results
    .filter(item => item.Details?.pricing?.monthly_rent || item.Details?.pricing?.sale_price)
    .map(item => item.Details.pricing.monthly_rent || item.Details.pricing.sale_price);

  if (prices.length === 0) return;

  const minPrice = Math.min(...prices);
  const maxPrice = Math.max(...prices);
  const roundedMax = Math.ceil(maxPrice / 10000) * 10000;

  const slider = document.getElementById('priceSlider');
  slider.min = minPrice;
  slider.max = roundedMax;
  slider.value = roundedMax;

  document.getElementById('priceDisplay').textContent =
    `${formatCurrency(minPrice)} - ${formatCurrency(roundedMax)}`;
}

// Clear all filters
document.getElementById('clearBtn').addEventListener('click', () => {
  document.getElementById('searchForm').reset();
  document.getElementById('advancedSearchSection').style.display = 'none';
});

// Toggle advanced search
document.getElementById('advancedSearchBtn').addEventListener('click', () => {
  const section = document.getElementById('advancedSearchSection');
  section.style.display = section.style.display === 'none' ? 'block' : 'none';
});
```

---

## API Examples

### Complete Search Example (Residential)

```bash
# Search for 3BR serviced apartments with parking in Gulshan
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "https://api.nestlo.com/api/v1/assets?
    status=listed_rental \
    &city=Dhaka \
    &neighborhood=Gulshan \
    &types=Apartment \
    &bedrooms=3 \
    &bathrooms=2 \
    &parking=2 \
    &serviced=true \
    &furnished=true \
    &price_min=80000 \
    &price_max=150000 \
    &limit=20"
```

### JavaScript Example

```javascript
async function searchProperties(filters) {
  const params = new URLSearchParams({
    status: filters.listingType || 'listed_rental,listed_sale',
    city: filters.city,
    neighborhood: filters.neighborhood,
    types: filters.types,
    bedrooms: filters.bedrooms,
    bathrooms: filters.bathrooms,
    parking: filters.parking,
    serviced: filters.serviced,
    shared_room: filters.sharedRoom,
    furnished: filters.furnished,
    price_min: filters.priceMin,
    price_max: filters.priceMax,
    limit: 20
  });

  // Remove empty params
  for (let [key, value] of params) {
    if (!value) params.delete(key);
  }

  const response = await fetch(
    `https://api.nestlo.com/api/v1/assets?${params}`,
    { headers: { 'Authorization': `Bearer ${TOKEN}` } }
  );

  return response.json();
}
```

---

## Frontend Integration

### React Component Example

```jsx
import { useState, useEffect } from 'react';

export function AdvancedSearch() {
  const [filters, setFilters] = useState({
    city: '',
    neighborhood: '',
    types: '',
    bedrooms: '',
    bathrooms: '',
    parking: '',
    serviced: false,
    sharedRoom: false,
    priceMin: 0,
    priceMax: 200000,
    furnished: false
  });

  const [results, setResults] = useState([]);
  const [cities, setCities] = useState([]);
  const [neighborhoods, setNeighborhoods] = useState([]);

  useEffect(() => {
    loadCities();
  }, []);

  const loadCities = async () => {
    const response = await fetch(
      'https://api.nestlo.com/api/v1/assets/cities?status=listed_rental,listed_sale',
      { headers: { 'Authorization': `Bearer ${TOKEN}` } }
    );
    const data = await response.json();
    setCities(data);
  };

  const loadNeighborhoods = async (city) => {
    if (!city) return;
    const response = await fetch(
      `https://api.nestlo.com/api/v1/assets/neighborhoods?city=${city}`,
      { headers: { 'Authorization': `Bearer ${TOKEN}` } }
    );
    const data = await response.json();
    setNeighborhoods(data);
  };

  const handleSearch = async () => {
    const params = new URLSearchParams();

    // Add status
    params.append('status', 'listed_rental,listed_sale');

    // Add other filters
    Object.entries(filters).forEach(([key, value]) => {
      if (value && value !== '' && value !== false) {
        params.append(key, value);
      }
    });

    const response = await fetch(
      `https://api.nestlo.com/api/v1/assets?${params}`,
      { headers: { 'Authorization': `Bearer ${TOKEN}` } }
    );
    const data = await response.json();

    // Update price slider
    const prices = data.data.map(item =>
      item.Details.pricing.monthly_rent || item.Details.pricing.sale_price
    );
    if (prices.length > 0) {
      const maxPrice = Math.ceil(Math.max(...prices) / 10000) * 10000;
      setFilters(prev => ({ ...prev, priceMax: maxPrice }));
    }

    setResults(data.data);
  };

  return (
    <div className="advanced-search">
      {/* City Select */}
      <select
        value={filters.city}
        onChange={(e) => {
          const city = e.target.value;
          setFilters(prev => ({ ...prev, city }));
          loadNeighborhoods(city);
        }}
      >
        <option value="">Select City</option>
        {cities.map(city => <option key={city}>{city}</option>)}
      </select>

      {/* Parking Select */}
      <select
        value={filters.parking}
        onChange={(e) => setFilters(prev => ({ ...prev, parking: e.target.value }))}
      >
        <option value="">Any</option>
        <option value="0">No Parking</option>
        <option value="1">1 Space</option>
        <option value="2">2 Spaces</option>
        <option value="3">3+ Spaces</option>
      </select>

      {/* Serviced Checkbox (for residential) */}
      {['Apartment', 'House'].includes(filters.types) && (
        <label>
          <input
            type="checkbox"
            checked={filters.serviced}
            onChange={(e) => setFilters(prev => ({ ...prev, serviced: e.target.checked }))}
          />
          Serviced Apartment
        </label>
      )}

      {/* Shared Room Checkbox (for hostel) */}
      {filters.types === 'Hostel' && (
        <label>
          <input
            type="checkbox"
            checked={filters.sharedRoom}
            onChange={(e) => setFilters(prev => ({ ...prev, sharedRoom: e.target.checked }))}
          />
          Shared Room
        </label>
      )}

      <button onClick={handleSearch}>Search</button>

      {/* Display Results */}
      <div className="results">
        {results.map(property => (
          <PropertyCard key={property.ID} property={property} />
        ))}
      </div>
    </div>
  );
}
```

---

## Data Structures

### Property Response Object

```json
{
  "ID": "uuid",
  "Name": "Spacious 3BR Apartment",
  "Type": "Apartment",
  "Status": "listed_rental",
  "Location": {
    "city": "Dhaka",
    "neighborhood": "Gulshan",
    "address": "House 12, Street 5, Gulshan 2",
    "lat": 23.7809,
    "lng": 90.4217
  },
  "Details": {
    "listing_title": "Luxury 3BR Apartment in Gulshan 2",
    "description": "Modern apartment with all amenities",
    "bedrooms": 3,
    "bathrooms": 2,
    "sizeSqft": 1850,
    "parkingSpaces": 2,
    "isServiced": true,
    "isSharedRoom": null,
    "furnishingStatus": "furnished",
    "amenities": ["Gym", "Pool", "Generator"],
    "pricing": {
      "monthly_rent": 95000,
      "security_deposit": 190000
    }
  },
  "Photos": [
    {
      "ViewURL": "https://cdn.nestlo.com/...",
      "IsCover": true
    }
  ]
}
```

---

## Troubleshooting

### Common Issues

**Issue: "Invalid status" error**
```
Solution: Always use status=listed_rental or status=listed_sale
❌ Wrong: /assets?city=Dhaka
✅ Correct: /assets?status=listed_rental,listed_sale&city=Dhaka
```

**Issue: No results from search**
```
Solution:
1. Verify status parameter is included
2. Check spelling of city/neighborhood names
3. Try with fewer filters (remove optional ones)
4. Verify price range is reasonable
```

**Issue: Parking filter not working**
```
Solution:
- parking value must be 0, 1, 2, or 3
- 3 means "3 or more"
- Ensure your property has parkingSpaces set in the database
```

**Issue: Serviced filter not returning results**
```
Solution:
- Only applies to Apartment, House types
- Property must have isServiced field set
- Boolean must be "true" or "false" (not 1/0)
```

**Issue: Price slider not updating dynamically**
```
Solution:
- JavaScript must calculate min/max from search results
- Ensure pricing data exists in response
- Check that monthly_rent or sale_price fields are populated
```

---

## API Response Codes

| Code | Meaning | Solution |
|------|---------|----------|
| 200 | Success | Results will be in the response |
| 400 | Bad Request | Check query parameters, especially `status` |
| 401 | Unauthorized | Verify OAuth token is valid and not expired |
| 429 | Rate Limited | Wait before retrying (check `X-RateLimit-Reset` header) |
| 500 | Server Error | Contact support with request details |

---

## Rate Limiting

**Search Endpoints:** 100 requests/minute
**Detail Endpoints:** 500 requests/minute
**Configuration Endpoints:** 50 requests/minute

**Monitor these headers:**
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1702737600
```

---

## Support & Questions

For technical support and questions about the API:
- 📧 Email: api-support@nestlo.com
- 📚 Documentation: This guide
- 🐛 Report Issues: Include request URL, error message, and response

---

**End of Advanced Search Feature Guide**

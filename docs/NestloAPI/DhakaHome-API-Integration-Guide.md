# Dhaka Home - Nestlo API Integration Guide

**Version:** 3.0 (Consolidated & Complete)
**Date:** December 19, 2024
**Status:** ✅ Production Ready
**Audience:** DhakaHome Development Team

---

## Table of Contents

1. [Quick Start](#quick-start)
2. [Authentication](#authentication)
3. [API Endpoints Reference](#api-endpoints-reference)
4. [Advanced Search Implementation](#advanced-search-implementation)
5. [Frontend Integration](#frontend-integration)
6. [Data Structures](#data-structures)
7. [Implementation Examples](#implementation-examples)
8. [Troubleshooting & Support](#troubleshooting--support)

---

## Quick Start

### Step 1: Get OAuth Credentials

Contact your Nestlo administrator to provision OAuth credentials. You will receive:

- **Client ID**: Unique identifier for your application
- **Client Secret**: Secret key for authentication (keep this secure!)
- **Scope**: Usually `assets.read` for public property browsing

### Step 2: Generate Access Token

```bash
curl -X POST https://api.nestlo.com/oauth/token \
  -H "Content-Type: application/json" \
  -d '{
    "grant_type": "client_credentials",
    "client_id": "YOUR_CLIENT_ID",
    "client_secret": "YOUR_CLIENT_SECRET",
    "scope": "assets.read"
  }'
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsImtpZCI6IiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 900,
  "scope": "assets.read"
}
```

### Step 3: Use Token in Requests

```bash
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  "https://api.nestlo.com/api/v1/assets?status=listed_rental,listed_sale&city=Dhaka"
```

---

## Authentication

### OAuth 2.0 Client Credentials Flow

Nestlo API uses **OAuth 2.0 Client Credentials** for machine-to-machine (M2M) authentication. This is ideal for:

- Public-facing property search websites
- Anonymous property browsing
- Service-to-service integrations
- Mobile applications without user login

### Token Details

| Property | Value |
|----------|-------|
| **Type** | JWT (JSON Web Token) |
| **Signing Algorithm** | RS256 (RSA with SHA-256) |
| **Expiration** | 15 minutes (900 seconds) |
| **Refresh Strategy** | Request new token before expiration |

### How It Works

```
┌─────────────┐                              ┌─────────────┐
│ DhakaHome   │                              │ Nestlo API  │
└──────┬──────┘                              └──────┬──────┘
       │                                            │
       │  1. POST /oauth/token                      │
       │     (client_id + client_secret)            │
       ├───────────────────────────────────────────>│
       │                                            │
       │  2. Response: access_token                 │
       │<───────────────────────────────────────────┤
       │                                            │
       │  3. GET /api/v1/assets                     │
       │     Authorization: Bearer {token}          │
       ├───────────────────────────────────────────>│
       │                                            │
       │  4. Response: Property data                │
       │<───────────────────────────────────────────┤
```

### API Base URLs

| Environment | Base URL |
|-------------|----------|
| **Production** | `https://api.nestlo.com/api/v1` |
| **Staging** | `https://stage-nestlo-api.onrender.com/api/v1` |
| **Local Development** | `http://localhost:3000/api/v1` |

---

## API Endpoints Reference

### Core Search Endpoints

#### 1. **Search Properties with Filters**
```
GET /api/v1/assets
```

**Purpose:** Search and filter properties with comprehensive criteria

**Required Parameters:**
- `status` - Property listing status (required)
  - Values: `listed_rental,listed_sale` (or use both)

**Optional Filter Parameters:**

| Parameter | Type | Example | Description |
|-----------|------|---------|-------------|
| `city` | string | `Dhaka` | Filter by city |
| `neighborhood` | string | `Gulshan` | Filter by area/neighborhood |
| `types` | string | `Apartment,Hostel` | Comma-separated property types |
| `bedrooms` | integer | `3` | Minimum number of bedrooms |
| `bathrooms` | integer | `2` | Minimum number of bathrooms |
| `price_min` | float | `50000` | Minimum price |
| `price_max` | float | `150000` | Maximum price |
| `parking` | integer | `2` | Parking spaces (0, 1, 2, 3+) |
| `serviced` | boolean | `true` | Serviced apartment (residential only) |
| `shared_room` | boolean | `false` | Shared room (hostel only) |
| `furnished` | boolean | `true` | Furnished status |
| `limit` | integer | `20` | Results per page (max 100) |
| `page` | integer | `1` | Page number |

**Example Request:**
```bash
GET /api/v1/assets?status=listed_rental&city=Dhaka&neighborhood=Gulshan&types=Apartment&bedrooms=3&parking=2&limit=20
```

**Response:**
```json
{
  "data": [
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
        "bedrooms": 3,
        "bathrooms": 2,
        "sizeSqft": 1850,
        "parkingSpaces": 2,
        "isServiced": true,
        "furnishingStatus": "furnished",
        "amenities": ["Lift", "Gas Supply", "Generator Backup"],
        "pricing": {
          "monthly_rent": 95000,
          "security_deposit": 190000
        }
      },
      "Photos": [
        {
          "ViewURL": "https://cdn.nestlo.com/signed-urls/...",
          "IsCover": true
        }
      ]
    }
  ],
  "total": 45,
  "page": 1,
  "limit": 20
}
```

---

#### 2. **Get Cities Dropdown**
```
GET /api/v1/assets/cities
```

**Parameters:**
- `status` (optional) - Filter cities by listing status

**Example Request:**
```bash
GET /api/v1/assets/cities?status=listed_rental,listed_sale
```

**Response:**
```json
[
  "Dhaka",
  "Chittagong",
  "Sylhet",
  "Khulna"
]
```

---

#### 3. **Get Neighborhoods by City**
```
GET /api/v1/assets/neighborhoods
```

**Parameters:**
- `city` (required) - Filter by specific city
- `status` (optional) - Filter by listing status

**Example Request:**
```bash
GET /api/v1/assets/neighborhoods?city=Dhaka&status=listed_rental,listed_sale
```

**Response:**
```json
[
  "Gulshan",
  "Banani",
  "Dhanmondi",
  "Uttara",
  "Mirpur",
  "Bashundhara"
]
```

---

#### 4. **Get Top Neighborhoods by Property Count** ⭐
```
GET /api/v1/assets/neighborhoods/top
```

**Purpose:** Fetch neighborhoods ranked by property count (perfect for homepage featured areas)

**Parameters:**
- `limit` (optional) - Number of top neighborhoods to return (default: 10, max: 100)
- `city` (optional) - Filter results by specific city
- `status` (optional) - Comma-separated listing statuses

**Example Request:**
```bash
# Get top 10 neighborhoods across all listings
GET /api/v1/assets/neighborhoods/top?limit=10&status=listed_rental,listed_sale

# Get top 8 areas in Dhaka city only
GET /api/v1/assets/neighborhoods/top?limit=8&city=Dhaka&status=listed_rental
```

**Response:**
```json
[
  {
    "neighborhood": "Gulshan",
    "city": "Dhaka",
    "count": 145
  },
  {
    "neighborhood": "Banani",
    "city": "Dhaka",
    "count": 132
  },
  {
    "neighborhood": "Dhanmondi",
    "city": "Dhaka",
    "count": 98
  }
]
```

---

#### 5. **Get Property Type Configuration**
```
GET /api/v1/config/property-types
```

**Parameters:**
- `listingType` (optional) - `rental` or `sale`

**Response:** Property type hierarchy with features

---

#### 6. **Get Property Details**
```
GET /api/v1/assets/{id}
```

**Response:** Complete property information including photos, documents, pricing, amenities

---

#### 7. **Get Similar Properties**
```
GET /api/v1/assets/{id}/similar
```

**Parameters:**
- `limit` (optional) - Number of similar properties (default: 6, max: 10)

**Response:** Array of similar properties

---

## Advanced Search Implementation

### Filter Options

#### Parking Spaces Filter

**Field:** `parking`
**Type:** Integer
**Supported Values:**
- `0` - No parking
- `1` - 1 parking space
- `2` - 2 parking spaces
- `3` - 3 or more parking spaces

```bash
# Apartments with 2 parking spaces
GET /api/v1/assets?status=listed_rental&types=Apartment&parking=2
```

#### Serviced Apartments Filter

**Field:** `serviced`
**Type:** Boolean
**Applies to:** Residential properties (Apartment, House, Villa)

**Values:**
- `true` - Serviced apartments
- `false` - Non-serviced apartments
- omit - Show all

```bash
# Serviced apartments in Gulshan
GET /api/v1/assets?status=listed_rental&types=Apartment&neighborhood=Gulshan&serviced=true
```

#### Shared Rooms Filter

**Field:** `shared_room`
**Type:** Boolean
**Applies to:** Hostel properties only

**Values:**
- `true` - Shared rooms (multiple beds)
- `false` - Private rooms
- omit - Show all

```bash
# Shared room hostels
GET /api/v1/assets?status=listed_rental&types=Hostel&shared_room=true
```

#### Price Range Slider (Dynamic)

**Implementation Strategy:** Frontend-Computed

**Flow:**
1. User enters search criteria and submits
2. API returns all matching properties with prices
3. Frontend extracts all prices from results
4. Frontend calculates min and max price
5. Frontend rounds max to nearest 10,000
6. Set slider range: [minPrice, roundedMaxPrice]
7. User can adjust slider and search again

```javascript
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
}
```

---

## Frontend Integration

### HTML Form Template

```html
<form id="searchForm" onsubmit="handleSearch(event)">
  <!-- Basic Filters -->
  <select name="city" id="citySelect">
    <option value="">Select City</option>
  </select>

  <select name="neighborhood" id="neighborhoodSelect">
    <option value="">Select Neighborhood</option>
  </select>

  <select name="types" id="typeSelect">
    <option value="">Select Property Type</option>
    <option value="Apartment">Apartment</option>
    <option value="House">House</option>
    <option value="Hostel">Hostel</option>
    <option value="Office">Office</option>
  </select>

  <!-- Advanced Filters Button -->
  <button type="button" onclick="toggleAdvanced()">⚙️ Advanced Search</button>

  <!-- Advanced Filters (hidden by default) -->
  <div id="advancedFilters" style="display:none;">
    <select name="bedrooms">
      <option value="">Any Bedrooms</option>
      <option value="1">1+</option>
      <option value="2">2+</option>
      <option value="3">3+</option>
    </select>

    <select name="parking">
      <option value="">Any Parking</option>
      <option value="0">No Parking</option>
      <option value="1">1 Space</option>
      <option value="2">2 Spaces</option>
      <option value="3">3+ Spaces</option>
    </select>

    <label>
      <input type="checkbox" name="serviced">
      Serviced Apartment
    </label>

    <label>
      <input type="checkbox" name="shared_room">
      Shared Room
    </label>

    <label>
      <input type="checkbox" name="furnished">
      Furnished
    </label>

    <label>Price Max:
      <input type="range" name="price_max" id="price_max" min="0" max="500000" step="10000">
      <span id="priceDisplay">500,000</span>
    </label>
  </div>

  <button type="submit">🔍 Search</button>
  <button type="button" onclick="resetForm()">✕ Clear</button>
</form>

<!-- Top Areas Section -->
<section id="topAreasSection">
  <h2>Explore Popular Areas</h2>
  <div id="topAreasContainer" class="areas-grid">
    <!-- Area cards will be inserted here -->
  </div>
</section>

<!-- Search Results -->
<div id="searchResults"></div>
```

---

### JavaScript Implementation (Vanilla)

```javascript
const API_BASE = "https://api.nestlo.com/api/v1";
const TOKEN = "YOUR_ACCESS_TOKEN"; // Set from OAuth response

// 1. Initialize: Load cities on page load
document.addEventListener('DOMContentLoaded', async () => {
  await loadCities();
  await loadTopAreas();
});

// Load cities dropdown
async function loadCities() {
  const cities = await fetch(
    `${API_BASE}/assets/cities?status=listed_rental,listed_sale`,
    { headers: { 'Authorization': `Bearer ${TOKEN}` } }
  ).then(r => r.json());

  const select = document.getElementById('citySelect');
  cities.forEach(city => {
    const option = document.createElement('option');
    option.value = city;
    option.textContent = city;
    select.appendChild(option);
  });
}

// Load neighborhoods when city changes
document.getElementById('citySelect').addEventListener('change', async (e) => {
  const city = e.target.value;
  if (!city) return;

  const neighborhoods = await fetch(
    `${API_BASE}/assets/neighborhoods?city=${city}&status=listed_rental,listed_sale`,
    { headers: { 'Authorization': `Bearer ${TOKEN}` } }
  ).then(r => r.json());

  const select = document.getElementById('neighborhoodSelect');
  select.innerHTML = '<option value="">Select Neighborhood</option>';
  neighborhoods.forEach(neighborhood => {
    const option = document.createElement('option');
    option.value = neighborhood;
    option.textContent = neighborhood;
    select.appendChild(option);
  });
});

// Handle form submission
document.getElementById('searchForm').addEventListener('submit', async (e) => {
  e.preventDefault();
  await performSearch();
});

async function performSearch() {
  const formData = new FormData(document.getElementById('searchForm'));
  const params = new URLSearchParams();

  // Always include status
  params.append('status', 'listed_rental,listed_sale');

  // Add other filters
  ['city', 'neighborhood', 'types', 'bedrooms', 'parking', 'serviced', 'shared_room', 'furnished'].forEach(key => {
    const value = formData.get(key);
    if (value) params.append(key, value);
  });

  params.append('limit', '20');

  const results = await fetch(
    `${API_BASE}/assets?${params}`,
    { headers: { 'Authorization': `Bearer ${TOKEN}` } }
  ).then(r => r.json());

  // Update price slider based on results
  updateDynamicPriceSlider(results.data);

  // Display results
  displayResults(results.data);
}

// Load and display top neighborhoods
async function loadTopAreas() {
  try {
    const response = await fetch(
      `${API_BASE}/assets/neighborhoods/top?limit=10&status=listed_rental,listed_sale`,
      { headers: { 'Authorization': `Bearer ${TOKEN}` } }
    );

    const topAreas = await response.json();

    // Randomly select 4 areas from top 10
    const shuffled = topAreas.sort(() => Math.random() - 0.5);
    const selectedAreas = shuffled.slice(0, 4);

    const container = document.getElementById('topAreasContainer');
    container.innerHTML = selectedAreas
      .map(area => `
        <div class="area-card" onclick="searchByArea('${area.neighborhood}', '${area.city}')">
          <h3>${area.neighborhood}</h3>
          <p class="count">${area.count} properties</p>
          <p class="city">${area.city}</p>
          <a href="#" class="explore-link">Explore →</a>
        </div>
      `)
      .join('');
  } catch (error) {
    console.error('Error loading top areas:', error);
  }
}

// Search by area when user clicks area card
function searchByArea(neighborhood, city) {
  document.getElementById('citySelect').value = city;
  document.getElementById('neighborhoodSelect').value = neighborhood;
  document.getElementById('searchForm').dispatchEvent(new Event('submit'));
}

// Toggle advanced filters
function toggleAdvanced() {
  const section = document.getElementById('advancedFilters');
  section.style.display = section.style.display === 'none' ? 'block' : 'none';
}

// Clear all filters
function resetForm() {
  document.getElementById('searchForm').reset();
  document.getElementById('advancedFilters').style.display = 'none';
}

// Update price slider dynamically
function updateDynamicPriceSlider(results) {
  const prices = results
    .filter(item => item.Details?.pricing?.monthly_rent || item.Details?.pricing?.sale_price)
    .map(item => item.Details.pricing.monthly_rent || item.Details.pricing.sale_price);

  if (prices.length === 0) return;

  const minPrice = Math.min(...prices);
  const maxPrice = Math.max(...prices);
  const roundedMax = Math.ceil(maxPrice / 10000) * 10000;

  const slider = document.getElementById('price_max');
  slider.min = minPrice;
  slider.max = roundedMax;
  slider.value = roundedMax;

  document.getElementById('priceDisplay').textContent =
    `${formatCurrency(minPrice)} - ${formatCurrency(roundedMax)}`;
}

// Display search results
function displayResults(properties) {
  const container = document.getElementById('searchResults');

  if (properties.length === 0) {
    container.innerHTML = '<p>No properties found. Try adjusting your filters.</p>';
    return;
  }

  container.innerHTML = properties
    .map(property => `
      <div class="property-card">
        <img src="${property.Photos[0]?.ViewURL}" alt="${property.Name}">
        <h3>${property.Name}</h3>
        <p>${property.Details.bedrooms} BR | ${property.Details.bathrooms} BA | ${property.Details.sizeSqft} sqft</p>
        <p>📍 ${property.Location.neighborhood}, ${property.Location.city}</p>
        <p class="price">৳ ${property.Details.pricing.monthly_rent || property.Details.pricing.sale_price}</p>
        <a href="/property/${property.ID}">View Details</a>
      </div>
    `)
    .join('');
}

function formatCurrency(amount) {
  return new Intl.NumberFormat('en-BD', { style: 'currency', currency: 'BDT' }).format(amount);
}
```

---

### React Component Example

```jsx
import { useState, useEffect } from 'react';

const API_BASE = "https://api.nestlo.com/api/v1";
const TOKEN = "YOUR_ACCESS_TOKEN";

export function DhakaHomeSearch() {
  const [filters, setFilters] = useState({
    city: '',
    neighborhood: '',
    types: '',
    bedrooms: '',
    bathrooms: '',
    parking: '',
    serviced: false,
    sharedRoom: false,
    furnished: false,
  });

  const [results, setResults] = useState([]);
  const [cities, setCities] = useState([]);
  const [neighborhoods, setNeighborhoods] = useState([]);
  const [topAreas, setTopAreas] = useState([]);
  const [showAdvanced, setShowAdvanced] = useState(false);

  useEffect(() => {
    loadCities();
    loadTopAreas();
  }, []);

  const loadCities = async () => {
    const data = await fetch(
      `${API_BASE}/assets/cities?status=listed_rental,listed_sale`,
      { headers: { 'Authorization': `Bearer ${TOKEN}` } }
    ).then(r => r.json());
    setCities(data);
  };

  const loadNeighborhoods = async (city) => {
    if (!city) return;
    const data = await fetch(
      `${API_BASE}/assets/neighborhoods?city=${city}&status=listed_rental,listed_sale`,
      { headers: { 'Authorization': `Bearer ${TOKEN}` } }
    ).then(r => r.json());
    setNeighborhoods(data);
  };

  const loadTopAreas = async () => {
    const data = await fetch(
      `${API_BASE}/assets/neighborhoods/top?limit=10&status=listed_rental,listed_sale`,
      { headers: { 'Authorization': `Bearer ${TOKEN}` } }
    ).then(r => r.json());

    const shuffled = [...data].sort(() => Math.random() - 0.5);
    setTopAreas(shuffled.slice(0, 4));
  };

  const handleSearch = async () => {
    const params = new URLSearchParams();
    params.append('status', 'listed_rental,listed_sale');

    Object.entries(filters).forEach(([key, value]) => {
      if (value && value !== '' && value !== false) {
        params.append(key, value);
      }
    });

    const data = await fetch(
      `${API_BASE}/assets?${params}`,
      { headers: { 'Authorization': `Bearer ${TOKEN}` } }
    ).then(r => r.json());

    setResults(data.data);
  };

  const handleAreaClick = (neighborhood, city) => {
    setFilters(prev => ({ ...prev, city, neighborhood }));
  };

  return (
    <div className="dhaka-home-search">
      {/* Top Areas Section */}
      <section className="top-areas">
        <h2>Explore Popular Areas</h2>
        <div className="areas-grid">
          {topAreas.map(area => (
            <div
              key={`${area.neighborhood}-${area.city}`}
              className="area-card"
              onClick={() => handleAreaClick(area.neighborhood, area.city)}
            >
              <h3>{area.neighborhood}</h3>
              <p>{area.count} properties</p>
              <p className="city">{area.city}</p>
            </div>
          ))}
        </div>
      </section>

      {/* Search Form */}
      <form onSubmit={(e) => { e.preventDefault(); handleSearch(); }}>
        <select
          value={filters.city}
          onChange={(e) => {
            setFilters(prev => ({ ...prev, city: e.target.value }));
            loadNeighborhoods(e.target.value);
          }}
        >
          <option value="">Select City</option>
          {cities.map(city => <option key={city} value={city}>{city}</option>)}
        </select>

        <select
          value={filters.neighborhood}
          onChange={(e) => setFilters(prev => ({ ...prev, neighborhood: e.target.value }))}
        >
          <option value="">Select Neighborhood</option>
          {neighborhoods.map(n => <option key={n} value={n}>{n}</option>)}
        </select>

        <button type="button" onClick={() => setShowAdvanced(!showAdvanced)}>
          Advanced Search
        </button>

        {showAdvanced && (
          <div className="advanced-filters">
            <select value={filters.parking} onChange={(e) => setFilters(prev => ({ ...prev, parking: e.target.value }))}>
              <option value="">Any Parking</option>
              <option value="0">No Parking</option>
              <option value="1">1 Space</option>
              <option value="2">2 Spaces</option>
              <option value="3">3+ Spaces</option>
            </select>

            <label>
              <input
                type="checkbox"
                checked={filters.serviced}
                onChange={(e) => setFilters(prev => ({ ...prev, serviced: e.target.checked }))}
              />
              Serviced
            </label>

            <label>
              <input
                type="checkbox"
                checked={filters.furnished}
                onChange={(e) => setFilters(prev => ({ ...prev, furnished: e.target.checked }))}
              />
              Furnished
            </label>
          </div>
        )}

        <button type="submit">Search</button>
      </form>

      {/* Results */}
      <div className="results">
        {results.map(property => (
          <PropertyCard key={property.ID} property={property} />
        ))}
      </div>
    </div>
  );
}

function PropertyCard({ property }) {
  return (
    <div className="property-card">
      <img src={property.Photos[0]?.ViewURL} alt={property.Name} />
      <h3>{property.Name}</h3>
      <p>{property.Details.bedrooms} BR • {property.Details.bathrooms} BA</p>
      <p>{property.Location.neighborhood}, {property.Location.city}</p>
      <p className="price">৳ {property.Details.pricing.monthly_rent || property.Details.pricing.sale_price}</p>
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

### Neighborhood Statistics Response

```json
[
  {
    "neighborhood": "Gulshan",
    "city": "Dhaka",
    "count": 145
  },
  {
    "neighborhood": "Banani",
    "city": "Dhaka",
    "count": 132
  },
  {
    "neighborhood": "Dhanmondi",
    "city": "Dhaka",
    "count": 98
  }
]
```

---

## Implementation Examples

### Example 1: Basic Search by City

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "https://api.nestlo.com/api/v1/assets?status=listed_rental,listed_sale&city=Dhaka&limit=20"
```

### Example 2: Advanced Search with Filters

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "https://api.nestlo.com/api/v1/assets?status=listed_rental&city=Dhaka&neighborhood=Gulshan&types=Apartment&bedrooms=3&parking=2&serviced=true&furnished=true&limit=20"
```

### Example 3: Get Top Areas for Homepage

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "https://api.nestlo.com/api/v1/assets/neighborhoods/top?limit=10&status=listed_rental,listed_sale"
```

### Example 4: JavaScript - Token Management

```javascript
class NestloAPI {
  constructor(clientId, clientSecret) {
    this.clientId = clientId;
    this.clientSecret = clientSecret;
    this.token = null;
    this.tokenExpiry = null;
  }

  async getToken() {
    // Return cached token if still valid
    if (this.token && new Date() < this.tokenExpiry) {
      return this.token;
    }

    // Request new token
    const response = await fetch('https://api.nestlo.com/oauth/token', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        grant_type: 'client_credentials',
        client_id: this.clientId,
        client_secret: this.clientSecret,
        scope: 'assets.read'
      })
    });

    const data = await response.json();
    this.token = data.access_token;
    this.tokenExpiry = new Date(Date.now() + (data.expires_in * 1000) - 60000); // 1 min buffer
    return this.token;
  }

  async searchProperties(filters) {
    const token = await this.getToken();
    const params = new URLSearchParams({ status: 'listed_rental,listed_sale', ...filters });

    const response = await fetch(
      `https://api.nestlo.com/api/v1/assets?${params}`,
      { headers: { 'Authorization': `Bearer ${token}` } }
    );

    return response.json();
  }

  async getTopAreas(limit = 10) {
    const token = await this.getToken();
    const response = await fetch(
      `https://api.nestlo.com/api/v1/assets/neighborhoods/top?limit=${limit}&status=listed_rental,listed_sale`,
      { headers: { 'Authorization': `Bearer ${token}` } }
    );
    return response.json();
  }
}

// Usage
const api = new NestloAPI('YOUR_CLIENT_ID', 'YOUR_CLIENT_SECRET');
const results = await api.searchProperties({ city: 'Dhaka', bedrooms: 3 });
```

---

## Troubleshooting & Support

### Common Issues

**Issue: "Invalid status" error**
```
❌ Wrong: /assets?city=Dhaka
✅ Correct: /assets?status=listed_rental,listed_sale&city=Dhaka
```
**Solution:** Status parameter is REQUIRED in all searches.

---

**Issue: "Invalid client credentials" (401 Unauthorized)**
```
Solution:
- Verify Client ID and Client Secret are correct
- Check that credentials haven't expired
- Ensure you're sending them in the OAuth token request
- Verify the OAuth endpoint is correct for your environment
```

---

**Issue: No results from search**
```
Solution:
1. Verify status parameter includes valid values
2. Check spelling of city/neighborhood names
3. Try with fewer filters (remove optional ones)
4. Verify price range is reasonable
5. Check that properties exist with those filters
```

---

**Issue: Parking filter not working**
```
Solution:
- Parking value must be 0, 1, 2, or 3 only
- 3 means "3 or more"
- Ensure property has parkingSpaces set in database
```

---

**Issue: Serviced filter not returning results**
```
Solution:
- Only applies to: Apartment, House property types
- Property must have isServiced field set to true/false
- Boolean must be "true" or "false" (not 1/0)
```

---

**Issue: Top neighborhoods endpoint returns empty results**
```
Solution:
- Verify status parameter is included (e.g., status=listed_rental,listed_sale)
- Ensure properties have valid city and neighborhood data
- Check that limit parameter is between 1-100
- Verify OAuth token has "assets.read" scope
```

---

**Issue: Token expired errors (401)**
```
Solution:
- Tokens expire every 15 minutes
- Implement token caching with expiry check
- Request new token before using expired one
- See "Token Management" example above
```

---

### Error Response Codes

| Code | Error | Solution |
|------|-------|----------|
| 200 | Success | Results returned as expected |
| 400 | Bad Request | Check query parameters, especially `status` |
| 401 | Unauthorized | Verify OAuth token is valid and not expired |
| 403 | Forbidden | Token doesn't have required scope (needs `assets.read`) |
| 404 | Not Found | Resource doesn't exist |
| 429 | Rate Limited | Too many requests; wait before retrying |
| 500 | Server Error | Contact Nestlo support |

---

### Rate Limiting

**Limits per minute:**
- Search endpoints: 100 requests/minute
- Detail endpoints: 500 requests/minute

**Monitor these response headers:**
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1702737600
```

---

### Support Resources

For technical support and questions:

- **Email:** api-support@nestlo.com
- **Response Time:** 24 hours
- **Required Information for Support:**
  - Request URL
  - Error message
  - Response code
  - Request timestamp
  - OAuth Client ID (not secret)

---

## Critical Implementation Rules

1. ✅ **ALWAYS include `status` parameter** in every search request
2. ✅ **Store OAuth tokens securely** - never expose client_secret to frontend
3. ✅ **Implement token caching** - tokens expire every 15 minutes
4. ✅ **Use ViewURL not FileURL** for photos - ViewURL includes signed access
5. ✅ **Parking values:** Only 0, 1, 2, or 3 (3 = 3 or more)
6. ✅ **Serviced/Shared:** Only boolean true/false or omit for any
7. ✅ **Price slider:** Calculate min/max from frontend search results
8. ✅ **Validate environment URLs** before deploying to production

---

## Implementation Checklist

- [ ] OAuth credentials obtained from Nestlo
- [ ] Token management implemented with caching
- [ ] Basic search form created
- [ ] Cities dropdown populated
- [ ] Neighborhoods dropdown populated (city-dependent)
- [ ] Property search working with status parameter
- [ ] Advanced search filters implemented (parking, serviced, etc.)
- [ ] Dynamic price slider from results
- [ ] Top areas section on homepage
- [ ] Search results displayed with photos/pricing
- [ ] Error handling for API failures
- [ ] Rate limiting handled gracefully
- [ ] Token refresh before expiry
- [ ] Testing in staging environment
- [ ] Production deployment

---

**Last Updated:** December 19, 2024
**Status:** ✅ Production Ready
**Version:** 3.0

For questions or feedback, contact: api-support@nestlo.com

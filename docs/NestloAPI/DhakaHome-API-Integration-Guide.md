# Dhaka Home - Nestlo API Integration Guide

**Version:** 3.0 (Consolidated & Complete)
**Date:** December 19, 2024
**Status:** ✅ Production Ready
**Audience:** DhakaHome Development Team

---

## Table of Contents

1. [Quick Start](#quick-start)
2. [Authentication](#authentication)
   - [User Login (Email/Password)](#user-login-emailpassword--new)
   - [OAuth 2.0 (Service-to-Service)](#oauth-20-client-credentials-flow)
3. [API Endpoints Reference](#api-endpoints-reference)
   - [Lead Creation (Contact Form)](#0-create-a-new-lead-from-contact-form--new)
   - [Property Search](#1-search-properties-with-filters)
   - [Cities & Neighborhoods](#2-get-cities-dropdown)
   - [Shortlist Management](#7-shortlist-management--new)
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

### User Login (Email/Password) ⭐ NEW

If you want to enable users registered in Nestlo to login directly from your DhakaHome website, use the user login endpoint.

**✅ YES, This is Fully Supported!**

All users registered in the Nestlo system with DhakaHome tenant can login via their email and password from your website. There are NO restrictions preventing this.

#### How It Works

1. User registered in Nestlo with DhakaHome can login from your site
2. You call the login endpoint with their email + password
3. You receive a JWT token valid for 24 hours
4. You use that token for subsequent API calls

#### Endpoint

```
POST /api/v1/auth/login
```

**Authentication:** Not required (no token needed for login)

#### Request Body

```json
{
  "email": "string",    // Required: User's registered email
  "password": "string"  // Required: User's password
}
```

#### Example Request

```bash
curl -X POST https://api.nestlo.com/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePassword123!"
  }'
```

#### Response (200 OK)

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Ahmed Hassan",
    "email": "ahmed@example.com",
    "phone_number": "+880123456789",
    "role": "tenant",
    "status": "active",
    "created_at": "2025-12-20T10:30:00Z",
    "image_url": "https://cdn.nestlo.com/user-123.jpg",
    "selected_asset_id": "550e8400-e29b-41d4-a716-446655440100"
  }
}
```

#### Token Details

| Property | Value |
|----------|-------|
| **Type** | JWT (JSON Web Token) |
| **Algorithm** | HS256 (HMAC with SHA-256) |
| **Expiration** | 24 hours |
| **Format** | Bearer token |
| **Contains** | user_id, role, expiration |

#### Using the Token in Requests

Once you have the token, include it in all subsequent API requests:

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "https://api.nestlo.com/api/v1/assets?status=listed_rental"
```

#### JavaScript Example

```javascript
async function loginUser(email, password) {
  const response = await fetch('https://api.nestlo.com/api/v1/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
  });

  if (response.status === 200) {
    const { token, user } = await response.json();

    // Store token securely (httpOnly cookie recommended)
    localStorage.setItem('nestlo_token', token);
    localStorage.setItem('nestlo_user', JSON.stringify(user));

    console.log('Login successful:', user.name);
    return { token, user };
  } else {
    const error = await response.json();
    console.error('Login failed:', error);
    throw new Error('Invalid credentials');
  }
}
```

#### Security Considerations

1. **Store Token Securely**: Use httpOnly cookies (preferred) or secure session storage, NOT localStorage for sensitive operations
2. **Token Expiration**: Token expires after 24 hours. Implement refresh logic or require re-login
3. **HTTPS Only**: Always use HTTPS in production to protect credentials in transit
4. **Client Secret Safety**: Never expose client_secret or user password in frontend code
5. **Password Requirements**:
   - Minimum 10 characters
   - At least one uppercase letter
   - At least one lowercase letter
   - At least one number
   - At least one special character (!@#~$%^&*()+|_.,<>?/\-)

#### Error Responses

**Invalid Credentials (401):**
```json
{
  "error": "Invalid credentials"
}
```

**Unverified Email (401):**
```json
{
  "error": "The account is not verified, please verify first."
}
```

**Pending Role Selection (200 but status = pending_role_selection):**
```json
{
  "token": "...",
  "user": {
    "status": "pending_role_selection",
    "role": ""
  }
}
```

Users with pending role selection CAN login but need to complete role selection first.

---

### OAuth 2.0 Client Credentials Flow

For server-to-server (backend) communication, Nestlo API uses **OAuth 2.0 Client Credentials**. This is ideal for:

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

### Lead Creation Endpoint

#### 0. **Create a New Lead from Contact Form** ⭐ NEW
```
POST /api/v1/admin/leads
```

**Purpose:** Create a new lead when a visitor submits your contact form with property inquiry details

**Authentication:** Required (Bearer token)

**Request Body:**

```json
{
  "lead_type": "string",                          // Required: tenant, buyer, rental, sale
  "source": "string",                             // Required: web, digital_media, physical_ad, referral, internal
  "client_info": {                                // Required
    "name": "string",                             // Required: visitor name (from contact form)
    "email": "string",                            // Optional: visitor email
    "phone": "string",                            // Optional: visitor phone number
    "preferred_contact_method": "string"          // Optional: phone, email, whatsapp, sms, any
  },
  "requirements": {                               // Optional: property requirements
    "property_types": ["string"],                 // Array: apartment, house, commercial, etc.
    "locations": ["string"],                      // Array: preferred areas (Gulshan, Banani, etc.)
    "budget_min": "number",                       // Min budget
    "budget_max": "number",                       // Max budget
    "bedrooms": "integer",                        // Preferred number of bedrooms
    "bathrooms": "integer",                       // Preferred number of bathrooms
    "amenities": ["string"],                      // Array: parking, gym, pool, etc.
    "move_in_date": "string",                     // YYYY-MM-DD format
    "notes": "string"                             // Specific requirements notes
  },
  "notes": "string",                              // Optional: general message from visitor
  "asset_id": "UUID"                              // Optional: specific property ID if they inquired about a property
}
```

**Example Request (Contact Form Submission):**

```bash
curl -X POST https://api.nestlo.com/api/v1/admin/leads \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "lead_type": "tenant",
    "source": "web",
    "client_info": {
      "name": "Ahmed Hassan",
      "email": "ahmed@example.com",
      "phone": "+880123456789",
      "preferred_contact_method": "phone"
    },
    "requirements": {
      "property_types": ["apartment"],
      "locations": ["Gulshan", "Banani"],
      "budget_min": 25000,
      "budget_max": 40000,
      "bedrooms": 2,
      "bathrooms": 1,
      "move_in_date": "2025-02-01"
    },
    "notes": "Interested in furnished apartment with parking. Looking to move in February 2025.",
    "asset_id": "550e8400-e29b-41d4-a716-446655440000"
  }'
```

**Response (201 Created):**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440001",
  "tenant_id": 1,
  "created_at": "2025-12-20T10:30:00Z",
  "updated_at": "2025-12-20T10:30:00Z",
  "lead_type": "tenant",
  "stage": "lead_capture",
  "source": "web",
  "priority": "medium",
  "status": "unassigned",
  "asset_id": "550e8400-e29b-41d4-a716-446655440000",
  "client_info": {
    "name": "Ahmed Hassan",
    "email": "ahmed@example.com",
    "phone": "+880123456789",
    "preferred_contact_method": "phone"
  },
  "requirements": {
    "property_types": ["apartment"],
    "locations": ["Gulshan", "Banani"],
    "budget_min": 25000,
    "budget_max": 40000,
    "bedrooms": 2,
    "bathrooms": 1,
    "move_in_date": "2025-02-01"
  },
  "interested_asset_ids": [],
  "assigned_agent_id": null,
  "assigned_at": null,
  "notes": "Interested in furnished apartment with parking. Looking to move in February 2025.",
  "version": 1
}
```

**Field Validation Rules:**

| Field | Required | Format | Examples |
|-------|----------|--------|----------|
| `lead_type` | ✅ Yes | Enum | tenant, buyer, rental, sale |
| `source` | ✅ Yes | Enum | web, digital_media, physical_ad, referral, internal |
| `client_info.name` | ✅ Yes | Non-empty string | "Ahmed Hassan" |
| `client_info.email` | ⚠️ Either email or phone | Email format | "ahmed@example.com" |
| `client_info.phone` | ⚠️ Either email or phone | E.164 format | "+880123456789" |
| `preferred_contact_method` | ❌ No | phone, email, whatsapp, sms, in_person, any | "phone" |

**Mapping from Your Contact Form to Lead Fields:**

Your form has: `name`, `phone`, `email`, `message`

Here's how to map them:

| Form Field | Nestlo Field | Notes |
|-----------|-------------|-------|
| name | `client_info.name` | Required - visitor's full name |
| email | `client_info.email` | Optional - visitor's email |
| phone | `client_info.phone` | Optional - visitor's phone number (use E.164 format with +880) |
| message | `notes` | Optional - customer's inquiry message |
| Property URL/ID | `asset_id` | Optional - if they're inquiring about specific property |

**Property Details from Contact Form (Optional):**

If your contact form also captures property preferences, map them to `requirements`:

| Form Input | Nestlo Field | Example |
|-----------|-------------|---------|
| "What type of property?" | `requirements.property_types` | ["apartment", "house"] |
| "Preferred areas?" | `requirements.locations` | ["Gulshan", "Banani"] |
| "Budget range?" | `requirements.budget_min/max` | 25000, 40000 |
| "How many bedrooms?" | `requirements.bedrooms` | 2 |
| "How many bathrooms?" | `requirements.bathrooms` | 1 |
| "Move-in date?" | `requirements.move_in_date` | "2025-02-01" |
| "Special features?" | `requirements.amenities` | ["parking", "gym"] |

**JavaScript Example (Contact Form Submission):**

```javascript
async function submitContactForm(formData) {
  const token = await getAccessToken(); // Get your OAuth token

  const leadData = {
    lead_type: "tenant", // or "buyer", "rental", "sale" based on form
    source: "web",       // Since form is on website
    client_info: {
      name: formData.name,
      email: formData.email,
      phone: formData.phone,
      preferred_contact_method: formData.phone ? "phone" : "email"
    },
    notes: formData.message,
    requirements: {
      property_types: formData.propertyTypes ? [formData.propertyTypes] : undefined,
      locations: formData.locations ? [formData.locations] : undefined,
      budget_min: formData.budgetMin,
      budget_max: formData.budgetMax,
      bedrooms: formData.bedrooms,
      bathrooms: formData.bathrooms,
      move_in_date: formData.moveInDate
    },
    asset_id: formData.propertyId // If they clicked "Inquire" on a specific property
  };

  const response = await fetch('https://api.nestlo.com/api/v1/admin/leads', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(leadData)
  });

  if (response.status === 201) {
    const lead = await response.json();
    console.log('Lead created successfully:', lead.id);
    // Show success message to user
    showSuccessMessage('Your inquiry has been received. We will contact you shortly.');
  } else {
    const error = await response.json();
    console.error('Failed to create lead:', error);
    showErrorMessage('There was an error submitting your inquiry. Please try again.');
  }
}
```

**Important Notes:**

1. **Email & Phone Format**: Phone should be in E.164 format (international format with +880 prefix for Bangladesh)
2. **At least one contact method required**: Either email OR phone must be provided
3. **Optional property link**: If user submitted the form from a specific property page, include the property ID in `asset_id`
4. **Status tracking**: Leads created via this endpoint are automatically marked as "unassigned" until an admin assigns an agent
5. **Admin visibility**: All submitted leads appear in the Nestlo Admin Dashboard under "Leads" section

**Error Responses:**

```json
{
  "error": "client_name_required"
}
```

**Common Error Codes:**

| Error | Cause | Solution |
|-------|-------|----------|
| `client_name_required` | Visitor name is empty | Validate form - name field is required |
| `contact_method_required` | No email or phone provided | Require at least email OR phone in form |
| `invalid_email_format` | Email format is incorrect | Validate email before submission |
| `invalid_phone_format` | Phone not in E.164 format | Format phone as +880XXXXXXXXX |
| `invalid_lead_type` | Lead type not in allowed values | Use: tenant, buyer, rental, sale |
| `invalid_source` | Source not in allowed values | Use: web, digital_media, physical_ad, referral, internal |

---

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

#### 7. **Shortlist Management** ⭐ NEW

Users can save (shortlist) properties they're interested in for later review. All shortlist operations require authentication.

##### 7a. Get User's Shortlists
```
GET /api/v1/shortlists
```

**Authentication:** Required (Bearer token - user JWT only)

**Purpose:** Retrieve all shortlists for the authenticated user with item counts

**Response (200 OK):**
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440050",
    "name": "Favorites",
    "description": "My favorite properties",
    "is_default": true,
    "item_count": 5,
    "created_at": "2025-12-20T10:30:00Z",
    "updated_at": "2025-12-20T14:45:00Z"
  },
  {
    "id": "550e8400-e29b-41d4-a716-446655440051",
    "name": "Investment Properties",
    "description": "For long-term investment",
    "is_default": false,
    "item_count": 3,
    "created_at": "2025-12-19T08:15:00Z",
    "updated_at": "2025-12-20T12:00:00Z"
  }
]
```

---

##### 7b. Get Shortlist with Full Property Details
```
GET /api/v1/shortlists/{shortlist_id}
```

**Authentication:** Required (Bearer token - user JWT only)

**Parameters:**
- `shortlist_id` (path) - UUID of the shortlist

**Purpose:** Retrieve a specific shortlist with all its items and complete asset details

**Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440050",
  "name": "Favorites",
  "description": "My favorite properties",
  "is_default": true,
  "item_count": 2,
  "created_at": "2025-12-20T10:30:00Z",
  "updated_at": "2025-12-20T14:45:00Z",
  "items": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "added_at": "2025-12-20T14:00:00Z",
      "notes": "Great location with good amenities",
      "asset": {
        "ID": "550e8400-e29b-41d4-a716-446655440100",
        "Name": "Spacious 3BR Apartment in Gulshan",
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
    }
  ]
}
```

---

##### 7c. Add Property to Shortlist (Favorites)
```
POST /api/v1/shortlists/items
```

**Authentication:** Required (Bearer token - user JWT only)

**Purpose:** Add a property to the user's default "Favorites" shortlist. If the property is already shortlisted, updates the notes.

**Request Body:**
```json
{
  "asset_id": "550e8400-e29b-41d4-a716-446655440100",
  "notes": "Great location for long-term investment"
}
```

**Field Details:**
| Field | Required | Type | Description |
|-------|----------|------|-------------|
| `asset_id` | ✅ Yes | UUID string | ID of the property to shortlist |
| `notes` | ❌ No | string | Optional personal notes about the property |

**Response (201 Created):**
```json
{
  "message": "Property added to shortlist",
  "shortlist_id": "550e8400-e29b-41d4-a716-446655440050",
  "item_id": "660e8400-e29b-41d4-a716-446655440001",
  "asset_id": "550e8400-e29b-41d4-a716-446655440100"
}
```

**JavaScript Example:**
```javascript
async function shortlistProperty(assetId, notes = '') {
  const token = sessionStorage.getItem('nestlo_token');

  const response = await fetch(
    'https://api.nestlo.com/api/v1/shortlists/items',
    {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        asset_id: assetId,
        notes: notes
      })
    }
  );

  if (response.status === 201) {
    const result = await response.json();
    console.log('Property shortlisted:', result.shortlist_id);
    // Update UI - show "Remove from Favorites" button
    showSuccessMessage('Property added to your Favorites!');
  } else if (response.status === 404) {
    alert('Property not found');
  } else {
    alert('Error adding to shortlist');
  }
}
```

---

##### 7d. Remove Property from Shortlist
```
DELETE /api/v1/shortlists/items/{asset_id}
```

**Authentication:** Required (Bearer token - user JWT only)

**Parameters:**
- `asset_id` (path) - UUID of the property to remove

**Purpose:** Remove a property from all of the user's shortlists

**Response (200 OK):**
```json
{
  "message": "Property removed from shortlist"
}
```

**JavaScript Example:**
```javascript
async function removeFromShortlist(assetId) {
  const token = sessionStorage.getItem('nestlo_token');

  const response = await fetch(
    `https://api.nestlo.com/api/v1/shortlists/items/${assetId}`,
    {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    }
  );

  if (response.status === 200) {
    console.log('Property removed from shortlist');
    // Update UI - show "Add to Favorites" button
    showSuccessMessage('Property removed from your Favorites');
  } else {
    alert('Error removing from shortlist');
  }
}
```

---

##### 7e. Check if Property is Shortlisted
```
GET /api/v1/shortlists/check/{asset_id}
```

**Authentication:** Required (Bearer token - user JWT only)

**Parameters:**
- `asset_id` (path) - UUID of the property to check

**Purpose:** Check if a property is in any of the user's shortlists (useful for showing favorite badge on property cards)

**Response (200 OK):**
```json
{
  "is_shortlisted": true,
  "asset_id": "550e8400-e29b-41d4-a716-446655440100",
  "shortlist_id": "550e8400-e29b-41d4-a716-446655440050"
}
```

**JavaScript Example (with caching):**
```javascript
// Cache to avoid repeated API calls
const shortlistCache = new Map();

async function isPropertyShortlisted(assetId) {
  // Check cache first
  if (shortlistCache.has(assetId)) {
    return shortlistCache.get(assetId);
  }

  const token = sessionStorage.getItem('nestlo_token');

  const response = await fetch(
    `https://api.nestlo.com/api/v1/shortlists/check/${assetId}`,
    {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    }
  );

  if (response.status === 200) {
    const result = await response.json();
    shortlistCache.set(assetId, result.is_shortlisted);
    return result.is_shortlisted;
  }

  return false;
}

// Update property card UI with favorite status
async function updatePropertyCardUI(propertyId) {
  const isFavorited = await isPropertyShortlisted(propertyId);
  const favoriteBtn = document.querySelector(`[data-property-id="${propertyId}"] .favorite-btn`);

  if (isFavorited) {
    favoriteBtn.textContent = '❤️ Remove from Favorites';
    favoriteBtn.classList.add('favorited');
    favoriteBtn.onclick = () => removeFromShortlist(propertyId);
  } else {
    favoriteBtn.textContent = '🤍 Add to Favorites';
    favoriteBtn.classList.remove('favorited');
    favoriteBtn.onclick = () => shortlistProperty(propertyId);
  }
}
```

---

##### 7f. Toggle Favorite (Helper Function)
```javascript
// Convenience function to toggle favorite status
async function togglePropertyFavorite(assetId) {
  const isShortlisted = await isPropertyShortlisted(assetId);

  if (isShortlisted) {
    // Remove from favorites
    await removeFromShortlist(assetId);
    shortlistCache.set(assetId, false);
  } else {
    // Add to favorites
    await shortlistProperty(assetId, '');
    shortlistCache.set(assetId, true);
  }

  // Invalidate cache and update UI
  await updatePropertyCardUI(assetId);
}
```

---

#### 8. **Get Similar Properties**
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

  async createLead(leadData) {
    const token = await this.getToken();
    const response = await fetch(
      'https://api.nestlo.com/api/v1/admin/leads',
      {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(leadData)
      }
    );

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to create lead');
    }

    return response.json();
  }
}

// Usage
const api = new NestloAPI('YOUR_CLIENT_ID', 'YOUR_CLIENT_SECRET');
const results = await api.searchProperties({ city: 'Dhaka', bedrooms: 3 });
```

### Example 5: User Login Form (Email/Password)

This example shows how to implement a login form on DhakaHome for users already registered in Nestlo.

```html
<!-- HTML Login Form -->
<form id="loginForm" onsubmit="handleLogin(event)">
  <h2>Login to Your Account</h2>

  <div class="form-group">
    <label for="email">Email Address</label>
    <input
      type="email"
      id="email"
      name="email"
      placeholder="your@email.com"
      required
    >
  </div>

  <div class="form-group">
    <label for="password">Password</label>
    <input
      type="password"
      id="password"
      name="password"
      placeholder="Your Password"
      required
    >
  </div>

  <button type="submit" class="btn-primary">Sign In</button>
  <a href="/forgot-password">Forgot Password?</a>
  <a href="/register">Don't have an account? Register here</a>

  <div id="errorMessage" class="error-message"></div>
  <div id="successMessage" class="success-message"></div>
</form>
```

```javascript
// JavaScript to handle login
async function handleLogin(event) {
  event.preventDefault();

  const form = event.target;
  const email = form.querySelector('[name="email"]').value;
  const password = form.querySelector('[name="password"]').value;
  const errorDiv = document.getElementById('errorMessage');
  const successDiv = document.getElementById('successMessage');

  // Clear previous messages
  errorDiv.textContent = '';
  successDiv.textContent = '';

  try {
    // Call login endpoint
    const response = await fetch('https://api.nestlo.com/api/v1/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password })
    });

    if (response.status === 200) {
      const { token, user } = await response.json();

      // Store token securely
      // Option 1: HTTP-Only Cookie (preferred, handled by backend)
      // Option 2: Session Storage (more secure than localStorage)
      sessionStorage.setItem('nestlo_token', token);
      sessionStorage.setItem('nestlo_user', JSON.stringify(user));

      // Update UI
      successDiv.textContent = `Welcome, ${user.name}!`;

      // Redirect to dashboard after short delay
      setTimeout(() => {
        window.location.href = '/dashboard';
      }, 1500);
    } else if (response.status === 401) {
      const error = await response.json();
      errorDiv.textContent = error.error || 'Invalid email or password';
    } else {
      errorDiv.textContent = 'An error occurred. Please try again later.';
    }
  } catch (error) {
    console.error('Login error:', error);
    errorDiv.textContent = 'Network error. Please check your connection.';
  }
}

// Get token from storage for API calls
function getAuthToken() {
  return sessionStorage.getItem('nestlo_token');
}

// Make authenticated API call
async function fetchWithAuth(url, options = {}) {
  const token = getAuthToken();

  if (!token) {
    window.location.href = '/login';
    return;
  }

  const headers = {
    ...options.headers,
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  };

  const response = await fetch(url, { ...options, headers });

  // If token expired (401), redirect to login
  if (response.status === 401) {
    sessionStorage.removeItem('nestlo_token');
    sessionStorage.removeItem('nestlo_user');
    window.location.href = '/login';
    return;
  }

  return response;
}
```

---

### Example 6: Authenticated Property Search (After Login)

Once a user is logged in, use their token to search for properties.

```javascript
// Search properties with user's token
async function searchPropertiesForLoggedInUser(filters) {
  try {
    const params = new URLSearchParams({
      status: 'listed_rental,listed_sale',
      ...filters
    });

    const response = await fetchWithAuth(
      `https://api.nestlo.com/api/v1/assets?${params}`
    );

    if (!response.ok) {
      throw new Error('Failed to fetch properties');
    }

    const results = await response.json();
    displayProperties(results.data);
  } catch (error) {
    console.error('Search error:', error);
    alert('Error searching properties. Please try again.');
  }
}

// Example usage after login
async function loadDashboard() {
  const user = JSON.parse(sessionStorage.getItem('nestlo_user'));

  if (!user) {
    window.location.href = '/login';
    return;
  }

  // Show welcome message
  document.getElementById('welcomeMessage').textContent = `Welcome, ${user.name}!`;

  // Load properties based on user's preferences
  await searchPropertiesForLoggedInUser({
    city: 'Dhaka',
    types: 'Apartment',
    bedrooms: 2
  });
}
```

---

### Example 7: Contact Form Submission to Create Lead

This example shows how to submit an inquiry as either a logged-in user or anonymous visitor.

```html
<!-- HTML Contact Form -->
<form id="contactForm" onsubmit="handleContactSubmit(event)">
  <input type="text" name="name" placeholder="Your Name" required>
  <input type="email" name="email" placeholder="Your Email">
  <input type="tel" name="phone" placeholder="Your Phone (with +880)">

  <select name="propertyType">
    <option value="">Select Property Type</option>
    <option value="apartment">Apartment</option>
    <option value="house">House</option>
    <option value="commercial">Commercial</option>
  </select>

  <select name="area">
    <option value="">Preferred Area</option>
    <option value="Gulshan">Gulshan</option>
    <option value="Banani">Banani</option>
    <option value="Dhanmondi">Dhanmondi</option>
  </select>

  <input type="number" name="budgetMin" placeholder="Min Budget">
  <input type="number" name="budgetMax" placeholder="Max Budget">

  <textarea name="message" placeholder="Your Message"></textarea>

  <button type="submit">Submit Inquiry</button>
</form>
```

```javascript
// For lead creation, you'll need OAuth token (client_credentials)
// This is done on your BACKEND, not frontend
const api = new NestloAPI('YOUR_CLIENT_ID', 'YOUR_CLIENT_SECRET');

async function handleContactSubmit(event) {
  event.preventDefault();

  const form = event.target;
  const formData = new FormData(form);

  try {
    // Build lead data from form
    const leadData = {
      lead_type: "tenant",
      source: "web",
      client_info: {
        name: formData.get('name'),
        email: formData.get('email') || undefined,
        phone: formData.get('phone') || undefined,
        preferred_contact_method: formData.get('phone') ? 'phone' : 'email'
      },
      requirements: {
        property_types: formData.get('propertyType') ? [formData.get('propertyType')] : undefined,
        locations: formData.get('area') ? [formData.get('area')] : undefined,
        budget_min: formData.get('budgetMin') ? parseInt(formData.get('budgetMin')) : undefined,
        budget_max: formData.get('budgetMax') ? parseInt(formData.get('budgetMax')) : undefined
      },
      notes: formData.get('message') || undefined
    };

    // Remove undefined fields
    Object.keys(leadData).forEach(key => {
      if (leadData[key] === undefined) delete leadData[key];
    });

    // Create the lead
    const result = await api.createLead(leadData);

    // Success response
    console.log('Lead created:', result.id);
    alert('Thank you! Your inquiry has been submitted. We will contact you shortly.');
    form.reset();
  } catch (error) {
    console.error('Error submitting form:', error);
    alert('Error submitting inquiry: ' + error.message);
  }
}
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
9. ✅ **Shortlist cache:** Cache shortlist status locally to reduce API calls
10. ✅ **Shortlist auth:** All shortlist endpoints require user JWT, NOT OAuth token
11. ✅ **Favorite toggle:** Implement optimistic UI updates for favorite button clicks
12. ✅ **Error handling:** Handle 404 (property not found) and 401 (auth expired) gracefully

---

## Implementation Checklist

### Authentication & Search
- [ ] OAuth credentials obtained from Nestlo
- [ ] Token management implemented with caching
- [ ] User login form implemented (email/password)
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

### Shortlist/Favorites (NEW)
- [ ] User registration for creating accounts
- [ ] User login and session management
- [ ] "Add to Favorites" button on property cards
- [ ] "Remove from Favorites" button on favorited properties
- [ ] Favorites counter badge showing number of saved properties
- [ ] Shortlist page to view all favorites with full details
- [ ] Cache shortlist status to reduce API calls
- [ ] Toggle favorite button with optimistic UI updates
- [ ] Display personal notes on favorited properties
- [ ] Clear favorites or batch operations (future)

### General
- [ ] Testing in staging environment
- [ ] Production deployment

---

**Last Updated:** December 21, 2024 (Added Shortlist/Favorites Feature)
**Status:** ✅ Production Ready
**Version:** 3.1 (Shortlist Feature Added)

For questions or feedback, contact: api-support@nestlo.com

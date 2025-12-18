# Dhaka Home Advanced Search - Quick Reference

**Last Updated:** December 16, 2024

---

## ⚡ Quick Start (Copy-Paste Ready)

### Authentication
```javascript
const TOKEN = "your-oauth-token-here";
const API_BASE = "https://api.nestlo.com/api/v1";
```

### Load Cities
```javascript
const cities = await fetch(
  `${API_BASE}/assets/cities?status=listed_rental,listed_sale`,
  { headers: { 'Authorization': `Bearer ${TOKEN}` } }
).then(r => r.json());
```

### Search Properties
```javascript
const results = await fetch(
  `${API_BASE}/assets?status=listed_rental,listed_sale&city=Dhaka&parking=2&serviced=true`,
  { headers: { 'Authorization': `Bearer ${TOKEN}` } }
).then(r => r.json());
```

---

## 🎯 Filter Parameters at a Glance

| Parameter | Values | Example |
|-----------|--------|---------|
| `status` | `listed_rental,listed_sale` | Both (REQUIRED) |
| `city` | City name | `Dhaka` |
| `neighborhood` | Area name | `Gulshan` |
| `types` | Property type | `Apartment,Hostel` |
| `bedrooms` | 1-5 | `3` |
| `bathrooms` | 1-4 | `2` |
| `parking` | 0, 1, 2, 3 | `2` (3 = 3+) |
| `serviced` | true/false | `true` |
| `shared_room` | true/false | `true` |
| `furnished` | true/false | `true` |
| `price_min` | amount | `50000` |
| `price_max` | amount | `150000` |
| `limit` | 1-100 | `20` |
| `page` | integer | `1` |

---

## 📌 **NEW** Filters (v2.0)

### Parking Spaces
```
?parking=0    → No Parking
?parking=1    → 1 Space
?parking=2    → 2 Spaces
?parking=3    → 3 or More
```

### Serviced Apartments (Residential Only)
```
?serviced=true   → Show serviced
?serviced=false  → Show non-serviced
(omit)           → Show all
```

### Shared Rooms (Hostel Only)
```
?shared_room=true   → Show shared rooms
?shared_room=false  → Show private rooms
(omit)              → Show all
```

---

## 🔗 Endpoint Reference

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/assets` | GET | Search properties |
| `/assets/cities` | GET | Get cities list |
| `/assets/neighborhoods` | GET | Get neighborhoods |
| `/assets/{id}` | GET | Get property details |
| `/assets/{id}/similar` | GET | Get similar properties |
| `/config/property-types` | GET | Get property types |

---

## 💻 Common Queries

### "3BR serviced apartments with 2 parking in Gulshan"
```
/assets?status=listed_rental&city=Dhaka&neighborhood=Gulshan&types=Apartment&bedrooms=3&parking=2&serviced=true
```

### "Commercial office space with 3+ parking"
```
/assets?status=listed_sale&types=Office&parking=3
```

### "Affordable shared room hostels"
```
/assets?status=listed_rental&types=Hostel&shared_room=true&price_max=20000
```

### "Furnished apartments in price range"
```
/assets?status=listed_rental&furnished=true&price_min=50000&price_max=150000
```

---

## 🎨 HTML Form Template

```html
<form onsubmit="handleSearch(event)">
  <!-- Basic Filters -->
  <select name="city" id="city">
    <option value="">Select City</option>
  </select>

  <select name="neighborhood" id="neighborhood">
    <option value="">Select Area</option>
  </select>

  <select name="types" id="types">
    <option value="">Select Type</option>
    <option value="Apartment">Apartment</option>
    <option value="House">House</option>
    <option value="Hostel">Hostel</option>
    <option value="Office">Office</option>
  </select>

  <!-- Advanced Filters Button -->
  <button type="button" onclick="toggleAdvanced()">⚙️ Advanced</button>

  <!-- Advanced Filters (hidden by default) -->
  <div id="advancedFilters" style="display:none;">
    <select name="bedrooms" id="bedrooms">
      <option value="">Any Bedrooms</option>
      <option value="1">1+</option>
      <option value="2">2+</option>
      <option value="3">3+</option>
    </select>

    <select name="parking" id="parking">
      <option value="">Any Parking</option>
      <option value="0">No Parking</option>
      <option value="1">1 Space</option>
      <option value="2">2 Spaces</option>
      <option value="3">3+ Spaces</option>
    </select>

    <label>
      <input type="checkbox" name="serviced" id="serviced">
      Serviced Apartment
    </label>

    <label>
      <input type="checkbox" name="shared_room" id="shared_room">
      Shared Room
    </label>

    <label>
      <input type="checkbox" name="furnished" id="furnished">
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
```

---

## 📊 Response Structure

```javascript
{
  "data": [
    {
      "ID": "uuid",
      "Name": "Property Name",
      "Type": "Apartment",
      "Details": {
        "bedrooms": 3,
        "bathrooms": 2,
        "parkingSpaces": 2,
        "isServiced": true,
        "pricing": {
          "monthly_rent": 95000,
          "sale_price": 5000000
        }
      },
      "Location": {
        "city": "Dhaka",
        "neighborhood": "Gulshan",
        "lat": 23.7809,
        "lng": 90.4217
      },
      "Photos": [{ "ViewURL": "https://...", "IsCover": true }]
    }
  ],
  "total": 45,
  "page": 1,
  "limit": 20
}
```

---

## ⚠️ Critical Rules

1. **ALWAYS include `status=listed_rental,listed_sale`** in every search
2. **Parking values:** Only 0, 1, 2, or 3 (3 = 3 or more)
3. **Serviced/Shared:** Only boolean true/false or omit for any
4. **Price slider:** Calculate min/max from frontend search results
5. **Photos:** Use `ViewURL` not `FileURL` (ViewURL is signed public URL)

---

## ❌ Common Mistakes

```javascript
// WRONG ❌
GET /assets?city=Dhaka  // Missing status!
GET /assets?parking=true  // Should be 0-3
GET /assets?serviced="yes"  // Should be boolean

// CORRECT ✅
GET /assets?status=listed_rental,listed_sale&city=Dhaka
GET /assets?status=listed_rental&parking=2
GET /assets?status=listed_rental&serviced=true
```

---

## 🚀 Frontend Implementation (Minimal)

```javascript
// 1. Load initial data
async function init() {
  const cities = await fetch(
    'https://api.nestlo.com/api/v1/assets/cities?status=listed_rental,listed_sale',
    { headers: { 'Authorization': `Bearer ${TOKEN}` } }
  ).then(r => r.json());

  document.getElementById('city').innerHTML = cities
    .map(city => `<option value="${city}">${city}</option>`)
    .join('');
}

// 2. Load neighborhoods on city change
async function onCityChange(city) {
  if (!city) return;
  const neighborhoods = await fetch(
    `https://api.nestlo.com/api/v1/assets/neighborhoods?city=${city}`,
    { headers: { 'Authorization': `Bearer ${TOKEN}` } }
  ).then(r => r.json());

  document.getElementById('neighborhood').innerHTML = neighborhoods
    .map(n => `<option value="${n}">${n}</option>`)
    .join('');
}

// 3. Perform search
async function handleSearch(e) {
  e.preventDefault();
  const formData = new FormData(e.target);

  const params = new URLSearchParams({
    status: 'listed_rental,listed_sale'
  });

  // Add filters from form
  ['city', 'neighborhood', 'types', 'bedrooms', 'parking', 'serviced', 'shared_room', 'furnished'].forEach(key => {
    const value = formData.get(key);
    if (value) params.append(key, value);
  });

  const results = await fetch(
    `https://api.nestlo.com/api/v1/assets?${params}`,
    { headers: { 'Authorization': `Bearer ${TOKEN}` } }
  ).then(r => r.json());

  // Update price slider based on results
  const prices = results.data
    .map(item => item.Details.pricing.monthly_rent || item.Details.pricing.sale_price)
    .filter(p => p);

  if (prices.length > 0) {
    const max = Math.ceil(Math.max(...prices) / 10000) * 10000;
    document.getElementById('price_max').max = max;
    document.getElementById('price_max').value = max;
  }

  displayResults(results.data);
}

// Initialize on page load
init();
document.getElementById('city').addEventListener('change', (e) => onCityChange(e.target.value));
```

---

## 📞 Support

**Issue: Status filter error?**
→ Always use: `status=listed_rental,listed_sale`

**Issue: No parking results?**
→ Check property has `parkingSpaces` set and use 0-3 only

**Issue: Serviced filter not working?**
→ Only works for: Apartment, House; property must have `isServiced` set

**Issue: Slider not updating?**
→ Calculate min/max from frontend results, not from backend

---

## 🎯 Feature Checklist

- [ ] Status filter always included
- [ ] Parking dropdown (0-3 values)
- [ ] Serviced checkbox (residential only)
- [ ] Shared room checkbox (hostel only)
- [ ] Dynamic price slider from results
- [ ] Clear/Reset button
- [ ] Advanced search toggle
- [ ] Results display with photos
- [ ] Similar properties link

---

**Version:** 2.0
**Status:** ✅ Production Ready
**Last Updated:** December 16, 2024

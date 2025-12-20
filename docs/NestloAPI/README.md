# DhakaHome - Nestlo API Integration

Welcome to the official Nestlo API Integration Guide for DhakaHome!

## 📖 Documentation Structure

### 📦 For DhakaHome Development Team

**Main Deliverable:**
- **[DhakaHome-API-Integration-Guide.md](./DhakaHome-API-Integration-Guide.md)** ⭐

  This is your **single, comprehensive guide** that includes:
  - OAuth authentication setup
  - All API endpoints reference
  - Advanced search implementation
  - Frontend integration examples (Vanilla JS + React)
  - Data structures and response formats
  - Complete implementation examples
  - Troubleshooting guide
  - Implementation checklist

  **Start here!** This document contains everything you need to integrate Nestlo APIs into DhakaHome.

---

### 📚 Internal Reference (For Nestlo Team)

These documents are for internal use and reference:
- `_internal/DHAKA_SITE_AUDIT.md` - Implementation audit and gap analysis
- `_internal/DHAKA_SITE_IMPLEMENTATION_CHECKLIST.md` - Internal tracking checklist

---

### 📂 Archived Documentation

Previous versions and consolidated documents are stored in `_archive/` folder for historical reference.

---

## 🚀 Quick Start for DhakaHome

### 1. **Get OAuth Credentials**
Contact your Nestlo administrator to provision OAuth credentials:
- Client ID
- Client Secret
- Scope: `assets.read`

### 2. **Read the Main Guide**
Open [DhakaHome-API-Integration-Guide.md](./DhakaHome-API-Integration-Guide.md) and follow the sections in order:
1. Quick Start
2. Authentication
3. API Endpoints Reference
4. Advanced Search Implementation
5. Frontend Integration (choose Vanilla JS or React)

### 3. **Implement**
Use the provided examples and follow the implementation checklist at the end of the guide.

### 4. **Test**
Use the provided curl examples to test endpoints before integrating into your frontend.

---

## 📋 Key Features Supported

✅ Property search with advanced filters
✅ City and neighborhood selection
✅ Top areas ranking (for homepage featured section)
✅ Dynamic price slider
✅ Parking, serviced apartments, shared rooms filters
✅ Complete property details with photos
✅ Similar properties recommendation

---

## 🔑 Critical Information

⚠️ **IMPORTANT:** The `status` parameter is **REQUIRED** in all API requests
- Use: `?status=listed_rental,listed_sale`

---

## 📞 Support

For technical support, contact:
- **Email:** api-support@nestlo.com
- **Response Time:** 24 hours
- **Required Info:** Request URL, error message, response code, timestamp

---

## 📅 Version History

| Version | Date | Changes |
|---------|------|---------|
| 3.0 | Dec 19, 2024 | Consolidated into single comprehensive guide |
| 2.1 | Dec 19, 2024 | Added Top Neighborhoods endpoint |
| 2.0 | Dec 16, 2024 | Added advanced search filters (parking, serviced, shared_room) |
| 1.0 | Dec 2024 | Initial API documentation |

---

**Last Updated:** December 19, 2024
**Status:** ✅ Production Ready
**Audience:** DhakaHome Development Team

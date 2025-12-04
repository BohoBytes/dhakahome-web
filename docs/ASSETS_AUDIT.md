# Assets Audit Report

**Date**: 2025-12-03
**Status**: ✅ Cleaned - All orphaned files removed

## Summary

- **Total image files**: 53 (after cleanup)
- **Orphaned files found**: 4 SVG files
- **Files removed**: 4 (19.5 KB total)
- **All remaining files**: ✅ In use

## Removed Files (Orphaned)

The following files were not referenced anywhere in the codebase and have been removed:

| File | Size | Status |
|------|------|--------|
| `34b5887a41be868fb5d38dc0839772fc66e1b33d.svg` | 2.1 KB | ✅ Removed |
| `515653c85bc1a4759e916797ee6e5c45dc844a55.svg` | 1.7 KB | ✅ Removed |
| `76e85f7d9da20feebf4735c2e7e9c2d6374067f3.svg` | 12 KB | ✅ Removed |
| `8f0842e42c8e13b6d459dfcb01b4addb15bc2142.svg` | 3.7 KB | ✅ Removed |

**Total space saved**: ~19.5 KB

## Verified Assets (All In Use)

### PNG Files - Property Images (Mock Data)

Used in: [internal/api/client.go](../internal/api/client.go) (mock property dataset)

| File | Usage Count | Purpose |
|------|-------------|---------|
| `1f002be890c252fab41bc52a14801210d4fa2535.png` | 7 occurrences | Mock property images |
| `2f8fe8dfbde9fb83f633da9c0e8bdff775034700.png` | 6 occurrences | Mock property images |
| `8abeccd3fd2f4096a7b4a66a184c5ae36074637a.png` | 6 occurrences | Mock property images |
| `d466fbc3c6a3829176f4bf45c88ed96204288a39.png` | 5 occurrences | Mock property images |
| `db6726f48a0bae50917980327e8ff5eb40ae871e.png` | 6 occurrences | Mock property images |

### PNG Files - Hero & Search Backgrounds

| File | Used In | Purpose |
|------|---------|---------|
| `hero-bg.png` | home.html, search-results.html, new-search-results.html | Hero section background |
| `hero-image.png` | hero.html, property-card.html (fallback) | Hero image display |
| `search-bg.png` | search-box.html, new-search.html | Search box background texture |

### PNG Files - Properties by Area Section

| File | Used In | Purpose |
|------|---------|---------|
| `images/area1.png` | properties-by-area.html | Area showcase card |
| `images/area2.png` | properties-by-area.html | Area showcase card |
| `images/area3.png` | properties-by-area.html | Area showcase card |
| `images/area4.png` | properties-by-area.html | Area showcase card |

### PNG Files - Property Details Pages

| File | Used In | Purpose |
|------|---------|---------|
| `images/property-interior.png` | property.html, property-details.html | Interior photo example |
| `images/property-office-1.png` | property-details.html | Similar properties section |
| `images/property-office-2.png` | property-details.html | Similar properties section |
| `images/footer-bg.png` | base.html | Footer background |

### SVG Files - Icons (37 icons, all in use)

#### Navigation & Branding
- `logo.svg` - Header logo
- `icons/logo-white.svg` - Footer logo
- `icons/logo-footer.svg` - Alternative footer logo
- `icons/logo-red.svg` - Brand variations

#### Property Features
- `icons/bedroom.svg` - Bedroom count indicator
- `icons/bathroom.svg` - Bathroom count indicator
- `icons/area.svg` - Area/size indicator
- `icons/parking.svg` - Parking availability

#### User Interface
- `icons/location.svg` - Location pins
- `icons/location-grey.svg` - Secondary location icons
- `icons/arrow_down.svg` - Dropdown indicators
- `icons/arrow-back.svg` - Back navigation
- `icons/chevron-down.svg` - Accordion toggles
- `icons/check-circle.svg` - Feature checkmarks

#### Services & Features
- `icons/key-ring.svg` - Property rental service
- `icons/home-gear.svg` - Property management
- `icons/hand-shake.svg` - Partnership/deals
- `icons/hotel-gear.svg` - Hostel management

#### Why DhakaHome Section
- `icons/badge-stars.svg` - Quality badge
- `icons/home-sold.svg` - Homes sold metric
- `icons/plot-sold.svg` - Plots sold metric
- `icons/client-man.svg` - Client count

#### Testimonials
- `icons/avatar-1.svg` - Testimonial avatar
- `icons/avatar-2.svg` - Testimonial avatar
- `icons/avatar-3.svg` - Testimonial avatar
- `icons/quote-start.svg` - Quote decoration
- `icons/quote-end.svg` - Quote decoration

#### Contact & Social
- `icons/phone.svg` - Phone contact
- `icons/phone-call.svg` - Call to action
- `icons/facebook.svg` - Social link
- `icons/google.svg` - Social link
- `icons/youtube.svg` - Social link
- `icons/social-media.svg` - Social media general

#### Property Details
- `icons/calendar.svg` - Year built
- `icons/date.svg` - Available date

#### Placeholder
- `property-placeholder.svg` - Fallback for missing property images

## Verification Commands

To verify no orphaned files remain:

```bash
# Search for all image references in code
grep -r "/assets/.*\.(png|svg|jpg|jpeg)" internal/ public/ --include="*.go" --include="*.html" --include="*.css"

# List all image files
find public/assets -type f \( -name "*.png" -o -name "*.svg" -o -name "*.jpg" -o -name "*.jpeg" \)
```

## Audit Methodology

1. **Scanned all code files** for image references:
   - Go files (`internal/api/*.go`)
   - HTML templates (`internal/views/**/*.html`)
   - CSS files (`public/assets/tailwind.css`)
   - Documentation (`docs/*.md`)

2. **Listed all image files** in public/assets directory

3. **Cross-referenced** each file against usage

4. **Identified orphans**: Files not referenced anywhere

5. **Verified and removed** orphaned files

## Maintenance Notes

### When Adding New Assets

1. Use descriptive names instead of hash names when possible
2. Organize by type:
   - `images/` - Photos and backgrounds
   - `icons/` - Icon graphics
   - Root - Logos and common images

### When Removing Features

Always check for orphaned assets after removing features:

```bash
# Find potential orphans (files not referenced in code)
find public/assets -name "*.png" -o -name "*.svg" | while read file; do
  basename=$(basename "$file")
  if ! grep -r "$basename" internal/ >/dev/null 2>&1; then
    echo "Potential orphan: $file"
  fi
done
```

## Conclusion

All PNG files are actively used. The codebase is clean with no orphaned image assets remaining after removing 4 unused SVG files.

**Next audit recommended**: After any major feature removal or UI redesign.

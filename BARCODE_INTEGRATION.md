# Barcode Scanning Integration Guide

This document provides comprehensive guidance for integrating the barcode scanning feature into your frontend application.

## Overview

The barcode scanning feature allows users to scan product barcodes (EAN-13, UPC, etc.) to automatically retrieve nutritional information from the Open Food Facts database and create food entries in the Ultra-Bis system.

**Key Features:**
- Scan any product barcode (EAN-13, UPC, etc.)
- Automatic retrieval of product data from Open Food Facts
- Instant creation of food entries with nutritional values per 100g
- No manual data entry required for common products

## API Endpoint

### Scan Barcode and Create Food

**Endpoint:** `POST /foods/barcode/{code}`

**Authentication:** Required (JWT Bearer token)

**Description:** Scans a barcode using the Open Food Facts API, retrieves product nutritional information, and automatically creates a food entry in the database.

**URL Parameters:**
- `code` (string, required) - The barcode number (e.g., "3017620422003")

**Request:**
```http
POST /foods/barcode/3017620422003
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Success Response (201 Created):**
```json
{
  "id": 42,
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T10:30:00Z",
  "name": "Nutella",
  "description": "Ferrero - Hazelnut cocoa spread",
  "calories": 539,
  "protein": 6.3,
  "carbs": 57.5,
  "fat": 30.9,
  "fiber": 0
}
```

**Error Responses:**

| Status Code | Error Message | Description |
|------------|---------------|-------------|
| 400 | "Barcode is required" | Empty or missing barcode |
| 400 | "Product name is missing from barcode data" | Invalid product data from API |
| 401 | "Unauthorized" | Missing or invalid JWT token |
| 404 | "Product not found for barcode: {code}" | Product doesn't exist in Open Food Facts database |
| 500 | "Failed to scan barcode: {error}" | API error or network issue |
| 500 | "Failed to create food: {error}" | Database error |

**Example Error Response (404):**
```json
{
  "error": "Product not found for barcode: 1234567890123"
}
```

## Frontend Implementation Guide

### 1. Prerequisites

You'll need a barcode scanning library for your frontend framework. Recommended libraries:

**React/React Native:**
- `react-native-camera` with barcode scanning
- `react-zxing` (web)
- `react-barcode-reader` (web)
- `expo-barcode-scanner` (Expo)

**Vue.js:**
- `vue-barcode-reader`
- `@zxing/library`

**Angular:**
- `@zxing/ngx-scanner`

**Flutter:**
- `mobile_scanner`
- `qr_code_scanner`

### 2. Basic Integration Flow

```javascript
// Example using React with react-zxing
import { useZxing } from "react-zxing";
import { useState } from "react";

function BarcodeScanner() {
  const [result, setResult] = useState(null);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(false);

  const { ref } = useZxing({
    onDecodeResult(result) {
      const barcode = result.getText();
      scanBarcode(barcode);
    },
  });

  const scanBarcode = async (barcode) => {
    setLoading(true);
    setError(null);

    try {
      const response = await fetch(`https://api.yourdomain.com/foods/barcode/${barcode}`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json'
        }
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error);
      }

      const food = await response.json();
      setResult(food);

      // Navigate to food details or show success message
      console.log('Food created:', food);

    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <video ref={ref} />
      {loading && <p>Scanning barcode...</p>}
      {error && <p style={{color: 'red'}}>Error: {error}</p>}
      {result && <p>Created food: {result.name}</p>}
    </div>
  );
}
```

### 3. Manual Barcode Entry (Fallback)

Not all users may have camera access or may prefer manual entry. Provide a text input as an alternative:

```javascript
function ManualBarcodeEntry() {
  const [barcode, setBarcode] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);

    try {
      const response = await fetch(`https://api.yourdomain.com/foods/barcode/${barcode}`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        }
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error);
      }

      const food = await response.json();
      // Handle success

    } catch (err) {
      alert(`Error: ${err.message}`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <input
        type="text"
        placeholder="Enter barcode number"
        value={barcode}
        onChange={(e) => setBarcode(e.target.value)}
        pattern="[0-9]+"
        required
      />
      <button type="submit" disabled={loading}>
        {loading ? 'Scanning...' : 'Scan Barcode'}
      </button>
    </form>
  );
}
```

### 4. User Experience Best Practices

**Loading States:**
- Show a loading indicator while scanning
- Display "Processing barcode..." message
- Disable the scan button during processing

**Error Handling:**
- **Product not found (404):** Offer option to manually create the food entry
- **Network errors:** Show retry button with exponential backoff
- **Invalid barcode:** Validate barcode format before sending request

**Success States:**
- Show the created food details
- Provide option to immediately add to meal diary
- Display option to edit nutritional values if needed

**Permission Handling:**
- Request camera permissions gracefully
- Explain why camera access is needed
- Provide manual entry fallback if permissions denied

### 5. Complete React Example with Error Handling

```javascript
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';

function BarcodeFeature() {
  const [scanMode, setScanMode] = useState('camera'); // 'camera' or 'manual'
  const [barcode, setBarcode] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const navigate = useNavigate();

  const scanBarcode = async (barcodeValue) => {
    setLoading(true);
    setError(null);

    try {
      const response = await fetch(
        `${process.env.REACT_APP_API_URL}/foods/barcode/${barcodeValue}`,
        {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`,
          }
        }
      );

      if (!response.ok) {
        const errorData = await response.json();

        if (response.status === 404) {
          // Product not found - offer manual creation
          setError({
            type: 'not_found',
            message: 'Product not found in database',
            barcode: barcodeValue
          });
          return;
        }

        throw new Error(errorData.error || 'Failed to scan barcode');
      }

      const food = await response.json();

      // Success! Navigate to the food details or diary entry
      navigate(`/foods/${food.id}`);

    } catch (err) {
      setError({
        type: 'error',
        message: err.message
      });
    } finally {
      setLoading(false);
    }
  };

  const handleManualEntry = (e) => {
    e.preventDefault();
    if (barcode.match(/^[0-9]{8,13}$/)) {
      scanBarcode(barcode);
    } else {
      setError({
        type: 'validation',
        message: 'Please enter a valid barcode (8-13 digits)'
      });
    }
  };

  const handleCreateManually = () => {
    navigate('/foods/create', { state: { barcode: error.barcode } });
  };

  return (
    <div className="barcode-scanner">
      <div className="mode-selector">
        <button
          onClick={() => setScanMode('camera')}
          className={scanMode === 'camera' ? 'active' : ''}
        >
          📷 Scan with Camera
        </button>
        <button
          onClick={() => setScanMode('manual')}
          className={scanMode === 'manual' ? 'active' : ''}
        >
          ⌨️ Enter Manually
        </button>
      </div>

      {scanMode === 'camera' ? (
        <div className="camera-scanner">
          {/* Your camera scanner component here */}
          <p>Point camera at barcode</p>
        </div>
      ) : (
        <form onSubmit={handleManualEntry}>
          <input
            type="text"
            placeholder="Enter barcode number"
            value={barcode}
            onChange={(e) => setBarcode(e.target.value)}
            pattern="[0-9]{8,13}"
            disabled={loading}
          />
          <button type="submit" disabled={loading}>
            {loading ? 'Scanning...' : 'Scan Barcode'}
          </button>
        </form>
      )}

      {loading && (
        <div className="loading">
          <span className="spinner"></span>
          <p>Fetching product information...</p>
        </div>
      )}

      {error && (
        <div className="error">
          <p>{error.message}</p>
          {error.type === 'not_found' && (
            <div>
              <p>Would you like to add this product manually?</p>
              <button onClick={handleCreateManually}>
                Create Food Entry
              </button>
            </div>
          )}
          {error.type === 'error' && (
            <button onClick={() => scanBarcode(barcode)}>
              Retry
            </button>
          )}
        </div>
      )}
    </div>
  );
}

export default BarcodeFeature;
```

### 6. Mobile App Considerations

**React Native Example:**

```javascript
import { Camera } from 'react-native-camera';
import { useState } from 'react';

function BarcodeScannerScreen() {
  const [scanned, setScanned] = useState(false);

  const handleBarCodeScanned = async ({ type, data }) => {
    if (scanned) return;
    setScanned(true);

    try {
      const response = await fetch(
        `https://api.yourdomain.com/foods/barcode/${data}`,
        {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${await getToken()}`,
          }
        }
      );

      const food = await response.json();

      // Navigate to success screen
      navigation.navigate('FoodDetails', { food });

    } catch (error) {
      Alert.alert('Error', error.message);
      setScanned(false); // Allow retry
    }
  };

  return (
    <Camera
      style={{ flex: 1 }}
      onBarCodeRead={handleBarCodeScanned}
      barCodeTypes={['ean13', 'ean8', 'upc_e', 'upc_a']}
    >
      {scanned && <ActivityIndicator size="large" />}
    </Camera>
  );
}
```

### 7. Testing the Feature

**Test with Known Barcodes:**

Here are some valid barcodes from Open Food Facts you can use for testing:

| Barcode | Product |
|---------|---------|
| 3017620422003 | Nutella (Ferrero) |
| 5449000000996 | Coca-Cola |
| 3228857000852 | Danone Yogurt |
| 8076809513012 | Barilla Pasta |

**Test Cases:**
1. ✅ Valid barcode with complete data
2. ✅ Valid barcode with missing description
3. ❌ Invalid barcode (product not found)
4. ❌ Empty barcode
5. ❌ Malformed barcode (letters, special characters)
6. ❌ No internet connection
7. ❌ Invalid authentication token

### 8. Open Food Facts Attribution

Open Food Facts is a collaborative, free, and open database of food products. When using their API:

**Display Attribution:**
Add this text in your app's settings or about section:
```
Product data sourced from Open Food Facts (https://world.openfoodfacts.org)
```

**License:** Open Database License (ODbL)

**Contributing Back:**
Consider allowing users to contribute missing products back to Open Food Facts through your app.

## Workflow Integration

### Typical User Flow

1. **User wants to log food**
   - Opens diary entry screen
   - Clicks "Scan Barcode" button

2. **Scanner opens**
   - Camera view appears
   - User points at product barcode
   - Barcode detected automatically

3. **Processing**
   - Loading indicator shows
   - API call to `/foods/barcode/{code}`
   - Food entry created

4. **Success**
   - Food details displayed
   - User can:
     - Add to diary immediately
     - View full nutritional info
     - Edit if needed
     - Save to favorites

5. **Product Not Found (Fallback)**
   - Error message displayed
   - Option to create manually
   - Pre-fill barcode in manual form

### Integration with Meal Logging

After scanning, immediately allow adding to diary:

```javascript
// After successful scan
const food = await response.json();

// Show quick-add dialog
showQuickAddDialog({
  food: food,
  onAdd: async (mealType, quantityGrams) => {
    await fetch('/diary/entries', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        food_id: food.id,
        date: new Date().toISOString().split('T')[0],
        meal_type: mealType,
        quantity_grams: quantityGrams,
        notes: `Scanned barcode: ${barcode}`
      })
    });
  }
});
```

## Technical Notes

### Important Considerations

1. **All nutritional values are per 100g** - The values returned from the barcode scan are per 100 grams, consistent with the Ultra-Bis data model.

2. **One-time creation** - Each barcode scan creates a new food entry. If the product already exists, consider implementing duplicate checking on the frontend.

3. **Network dependency** - Requires internet connection to access Open Food Facts API. Consider showing appropriate messages when offline.

4. **Rate limiting** - Open Food Facts API is free but should be used responsibly. Implement reasonable caching if scanning the same product multiple times.

5. **Data quality** - Not all products in Open Food Facts have complete nutritional data. Handle missing fields gracefully.

### Security

- **Authentication required** - All barcode scanning requests require a valid JWT token
- **User-scoped** - Created foods are associated with the authenticated user
- **Input validation** - Barcode format is validated on both frontend and backend

### Performance

- **Response time** - Typical response: 1-3 seconds (depends on Open Food Facts API)
- **Caching** - Consider caching frequently scanned products on the client
- **Offline mode** - Show appropriate message when network unavailable

## Support and Troubleshooting

### Common Issues

**"Product not found" for valid barcode:**
- Product may not exist in Open Food Facts database
- Offer manual entry option
- Consider allowing users to contribute to Open Food Facts

**Slow scan times:**
- Check internet connection
- Open Food Facts API may be slow (external dependency)
- Implement timeout (10 seconds recommended)

**Camera not working:**
- Check permissions
- Verify HTTPS (required for camera access in web browsers)
- Provide manual entry fallback

**Duplicate products:**
- Implement client-side checking before scanning
- Search existing foods by name before creating new entry
- Allow merging duplicates

## Future Enhancements

Potential features to consider:

- **Duplicate detection** - Check if food with same barcode already exists
- **Offline scanning** - Cache common products for offline use
- **Bulk scanning** - Scan multiple products for grocery shopping
- **Product history** - Track recently scanned products
- **Barcode favorites** - Quick access to frequently scanned items
- **Crowdsourcing** - Allow users to add missing products to Open Food Facts
- **Multi-language support** - Get product names in user's language
- **Allergen warnings** - Display allergen information from Open Food Facts

## API Reference Summary

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/foods/barcode/{code}` | POST | Required | Scan barcode and create food entry |

**Related Endpoints:**
- `POST /foods` - Manually create food (fallback)
- `GET /foods` - List all foods (check for duplicates)
- `POST /diary/entries` - Add scanned food to diary

## Contact and Support

For technical support or questions about this integration:
- Check the main API documentation in `FRONTEND_SPEC.md`
- Review the project documentation in `CLAUDE.md`
- Open an issue in the project repository

---

**Version:** 1.0
**Last Updated:** 2025-01-15
**API Version:** v1

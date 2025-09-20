#!/bin/bash

# Food Catalog API Test Script for Development Environment
# Make sure the development server is running on localhost:8081 before running this script

BASE_URL="http://localhost:8081"

echo "🍎 Food Catalog API Development Test Script"
echo "============================================="
echo

# Test 1: Get API info
echo "1. Getting API info..."
curl -s "$BASE_URL/" | head -n 10
echo -e "\n"

# Test 2: Create a new food item
echo "2. Creating a new food item (Apple)..."
APPLE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/foods" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Apple",
    "description": "Fresh red apple",
    "category": "fruit",
    "unit": "piece",
    "calories": 95,
    "protein": 0.5,
    "carbs": 25.0,
    "fat": 0.3
  }')

echo "$APPLE_RESPONSE" | jq '.'
APPLE_ID=$(echo "$APPLE_RESPONSE" | jq -r '.id')
echo "Created food with ID: $APPLE_ID"
echo

# Test 3: Create another food item
echo "3. Creating another food item (Chicken Breast)..."
CHICKEN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/foods" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Chicken Breast",
    "description": "Lean protein source",
    "category": "meat",
    "unit": "100g",
    "calories": 231,
    "protein": 43.5,
    "carbs": 0.0,
    "fat": 5.0
  }')

echo "$CHICKEN_RESPONSE" | jq '.'
CHICKEN_ID=$(echo "$CHICKEN_RESPONSE" | jq -r '.id')
echo "Created food with ID: $CHICKEN_ID"
echo

# Test 4: Get all foods
echo "4. Getting all foods..."
curl -s "$BASE_URL/api/foods" | jq '.'
echo

# Test 5: Get specific food by ID
echo "5. Getting apple by ID ($APPLE_ID)..."
curl -s "$BASE_URL/api/foods/$APPLE_ID" | jq '.'
echo

# Test 6: Update a food item
echo "6. Updating apple description and unit..."
UPDATED_APPLE=$(curl -s -X PUT "$BASE_URL/api/foods/$APPLE_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Crispy organic red apple",
    "unit": "piece",
    "calories": 90
  }')

echo "$UPDATED_APPLE" | jq '.'
echo

# Test 7: Get updated food
echo "7. Getting updated apple..."
curl -s "$BASE_URL/api/foods/$APPLE_ID" | jq '.'
echo

# Test 8: Test invalid requests
echo "8. Testing error cases..."

echo "8a. Invalid JSON:"
curl -s -X POST "$BASE_URL/api/foods" \
  -H "Content-Type: application/json" \
  -d '{"invalid": json}'
echo -e "\n"

echo "8b. Missing required fields:"
curl -s -X POST "$BASE_URL/api/foods" \
  -H "Content-Type: application/json" \
  -d '{"description": "Missing name and unit"}'
echo -e "\n"

echo "8c. Invalid unit:"
curl -s -X POST "$BASE_URL/api/foods" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Food",
    "description": "Test",
    "category": "test",
    "unit": "invalid_unit",
    "calories": 100,
    "protein": 10.0,
    "carbs": 10.0,
    "fat": 5.0
  }'
echo -e "\n"

echo "8d. Non-existent food ID:"
curl -s "$BASE_URL/api/foods/999"
echo -e "\n"

echo "8e. Invalid food ID:"
curl -s "$BASE_URL/api/foods/abc"
echo -e "\n"

# Test 9: Delete a food item
echo "9. Deleting chicken (ID: $CHICKEN_ID)..."
curl -s -X DELETE "$BASE_URL/api/foods/$CHICKEN_ID" -w "HTTP Status: %{http_code}\n"
echo

# Test 10: Verify deletion
echo "10. Verifying chicken deletion..."
curl -s "$BASE_URL/api/foods/$CHICKEN_ID"
echo -e "\n"

# Test 11: Get all foods after deletion
echo "11. Getting all foods after deletion..."
curl -s "$BASE_URL/api/foods" | jq '.'
echo

# Test 12: Clean up - delete remaining food
echo "12. Cleaning up - deleting apple (ID: $APPLE_ID)..."
curl -s -X DELETE "$BASE_URL/api/foods/$APPLE_ID" -w "HTTP Status: %{http_code}\n"
echo

echo "✅ Development API testing complete!"
echo
echo "Summary of tested endpoints:"
echo "- GET    /               - API info"
echo "- POST   /api/foods      - Create food"
echo "- GET    /api/foods      - Get all foods"
echo "- GET    /api/foods/{id} - Get food by ID"
echo "- PUT    /api/foods/{id} - Update food"
echo "- DELETE /api/foods/{id} - Delete food"
echo
echo "New features tested:"
echo "- Unit field validation (100g, 100ml, piece)"
echo "- Model changes (no fiber/sugar/sodium)"
echo "- Development environment (port 8081)"
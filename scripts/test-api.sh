#!/bin/bash

# Food Catalog API Test Script
# Make sure the server is running on localhost:8080 before running this script

BASE_URL="http://100.88.240.113:8080"
#BASE_URL="http://localhost:8080"

echo "🍎 Food Catalog API Test Script"
echo "================================="
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
    "calories": 95,
    "protein": 0.5,
    "carbs": 25.0,
    "fat": 0.3,
    "fiber": 4.0,
    "sugar": 19.0,
    "sodium": 1.0
  }')

echo "$APPLE_RESPONSE" | jq '.'
APPLE_ID=$(echo "$APPLE_RESPONSE" | jq -r '.id')
echo "Created food with ID: $APPLE_ID"
echo

# Test 3: Create another food item
echo "3. Creating another food item (Banana)..."
BANANA_RESPONSE=$(curl -s -X POST "$BASE_URL/api/foods" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Banana",
    "description": "Ripe yellow banana",
    "category": "fruit",
    "calories": 105,
    "protein": 1.3,
    "carbs": 27.0,
    "fat": 0.4,
    "fiber": 3.1,
    "sugar": 14.0,
    "sodium": 1.0
  }')

echo "$BANANA_RESPONSE" | jq '.'
BANANA_ID=$(echo "$BANANA_RESPONSE" | jq -r '.id')
echo "Created food with ID: $BANANA_ID"
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
echo "6. Updating apple description..."
UPDATED_APPLE=$(curl -s -X PUT "$BASE_URL/api/foods/$APPLE_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Crispy organic red apple",
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
  -d '{"description": "Missing name and category"}'
echo -e "\n"

echo "8c. Non-existent food ID:"
curl -s "$BASE_URL/api/foods/999"
echo -e "\n"

echo "8d. Invalid food ID:"
curl -s "$BASE_URL/api/foods/abc"
echo -e "\n"

# Test 9: Delete a food item
echo "9. Deleting banana (ID: $BANANA_ID)..."
curl -s -X DELETE "$BASE_URL/api/foods/$BANANA_ID" -w "HTTP Status: %{http_code}\n"
echo

# Test 10: Verify deletion
echo "10. Verifying banana deletion..."
curl -s "$BASE_URL/api/foods/$BANANA_ID"
echo -e "\n"

# Test 11: Get all foods after deletion
echo "11. Getting all foods after deletion..."
curl -s "$BASE_URL/api/foods" | jq '.'
echo

# Test 12: Clean up - delete remaining food
echo "12. Cleaning up - deleting apple (ID: $APPLE_ID)..."
curl -s -X DELETE "$BASE_URL/api/foods/$APPLE_ID" -w "HTTP Status: %{http_code}\n"
echo

echo "✅ API testing complete!"
echo
echo "Summary of tested endpoints:"
echo "- GET    /               - API info"
echo "- POST   /api/foods      - Create food"
echo "- GET    /api/foods      - Get all foods"
echo "- GET    /api/foods/{id} - Get food by ID"
echo "- PUT    /api/foods/{id} - Update food"
echo "- DELETE /api/foods/{id} - Delete food"
# Frontend Updates Required: Migration to Gram-Based System

## Overview

The backend API has been migrated from a serving-based system to a **gram-based quantity system**. All food nutritional values are now standardized to **per 100 grams**, and all quantities throughout the system are expressed in grams.

## Breaking Changes

### 1. Food Model
**No API changes**, but important clarification:
- All nutritional values (calories, protein, carbs, fat, fiber) represent values **per 100 grams**
- When displaying food nutrition to users, make sure to indicate "per 100g"

### 2. Recipe API Changes

#### Create Recipe (POST /recipes)
**Before:**
```json
{
  "name": "Chicken & Rice Bowl",
  "serving_size": 1,
  "ingredients": [
    {
      "food_id": 1,
      "quantity": 2
    }
  ]
}
```

**After:**
```json
{
  "name": "Chicken & Rice Bowl",
  "ingredients": [
    {
      "food_id": 1,
      "quantity_grams": 200
    }
  ]
}
```

**Changes:**
- ❌ Removed: `serving_size` field
- ✅ Changed: `quantity` → `quantity_grams` (use actual grams, e.g., 200 instead of 2)

#### Update Recipe (PUT /recipes/{id})
**Before:**
```json
{
  "name": "Updated Name",
  "serving_size": 1.5
}
```

**After:**
```json
{
  "name": "Updated Name"
}
```

**Changes:**
- ❌ Removed: `serving_size` field (only `name` can be updated)

#### Add/Update Ingredient
**Before:**
```json
{
  "food_id": 1,
  "quantity": 2
}
```

**After:**
```json
{
  "food_id": 1,
  "quantity_grams": 200
}
```

**Changes:**
- ✅ Changed: `quantity` → `quantity_grams`

#### Recipe Responses (GET /recipes, GET /recipes/{id})
**Before:**
```json
{
  "id": 1,
  "name": "Chicken & Rice Bowl",
  "serving_size": 1,
  "calories": 500,
  "protein": 40,
  "ingredients": [
    {
      "food_id": 1,
      "food_name": "Chicken",
      "quantity": 2,
      "calories": 330
    }
  ]
}
```

**After:**
```json
{
  "id": 1,
  "name": "Chicken & Rice Bowl",
  "total_weight": 350,
  "total_calories": 500,
  "total_protein": 40,
  "total_carbs": 35,
  "total_fat": 15,
  "total_fiber": 5,
  "calories_per_100g": 143,
  "protein_per_100g": 11.4,
  "carbs_per_100g": 10,
  "fat_per_100g": 4.3,
  "fiber_per_100g": 1.4,
  "ingredients": [
    {
      "food_id": 1,
      "food_name": "Chicken",
      "quantity_grams": 200,
      "calories": 330
    }
  ]
}
```

**Changes:**
- ❌ Removed: `serving_size`
- ✅ Added: `total_weight` (sum of all ingredient grams)
- ✅ Changed: `calories`, `protein`, etc. → `total_calories`, `total_protein`, etc.
- ✅ Added: `calories_per_100g`, `protein_per_100g`, etc. (normalized nutrition)
- ✅ Changed in ingredients: `quantity` → `quantity_grams`

### 3. Diary Entry API Changes

#### Create Diary Entry (POST /diary/entries)
**Before:**
```json
{
  "food_id": 1,
  "date": "2025-01-15",
  "meal_type": "breakfast",
  "serving_size": 1.5,
  "notes": "Post-workout"
}
```

**After:**
```json
{
  "food_id": 1,
  "date": "2025-01-15",
  "meal_type": "breakfast",
  "quantity_grams": 150,
  "notes": "Post-workout"
}
```

**Changes:**
- ✅ Changed: `serving_size` → `quantity_grams`
- Use actual grams (e.g., 150g instead of 1.5 servings)

#### Update Diary Entry (PUT /diary/entries/{id})
**Before:**
```json
{
  "serving_size": 2,
  "meal_type": "breakfast",
  "notes": "Updated"
}
```

**After:**
```json
{
  "quantity_grams": 200,
  "meal_type": "breakfast",
  "notes": "Updated"
}
```

**Changes:**
- ✅ Changed: `serving_size` → `quantity_grams`

#### Diary Entry Responses (GET /diary/entries)
**Before:**
```json
{
  "id": 1,
  "food_id": 1,
  "serving_size": 1.5,
  "calories": 247.5
}
```

**After:**
```json
{
  "id": 1,
  "food_id": 1,
  "quantity_grams": 150,
  "calories": 247.5
}
```

**Changes:**
- ✅ Changed: `serving_size` → `quantity_grams`

### 4. Recipe in Diary

When logging a recipe to the diary, users now specify **actual grams consumed** (not a serving multiplier).

**Before:**
```json
{
  "recipe_id": 1,
  "serving_size": 1.5,  // "I ate 1.5 servings"
  "meal_type": "lunch"
}
```

**After:**
```json
{
  "recipe_id": 1,
  "quantity_grams": 525,  // "I ate 525g of this recipe"
  "meal_type": "lunch"
}
```

**How to calculate:**
- If the recipe's `total_weight` is 350g and user ate the whole recipe → `quantity_grams: 350`
- If user ate half the recipe → `quantity_grams: 175`
- If user ate 1.5x the recipe → `quantity_grams: 525`

## UI/UX Recommendations

### Input Fields

1. **Food Logging:**
   - Replace "Serving size" with "Quantity (grams)"
   - Input: Number field with "g" suffix
   - Example: `[150] g` or with suffix `[150 g]`
   - Consider adding quick buttons: 50g, 100g, 150g, 200g

2. **Recipe Creation:**
   - Label: "Quantity (grams)" instead of "Servings"
   - Input: Number field for grams per ingredient
   - Example: "Chicken: `[200] g`"

3. **Recipe Display:**
   - Show total recipe weight: "Total: 350g"
   - Show per-100g nutrition for easy comparison
   - Optional: Calculate servings based on typical portion sizes (e.g., "~2 servings of 175g each")

### Display Formatting

**Food Cards:**
```
Chicken Breast
Per 100g:
• 165 calories
• 31g protein
• 0g carbs
• 3.6g fat
```

**Recipe Cards:**
```
Chicken & Rice Bowl
Total recipe: 350g

Total nutrition:
• 500 calories
• 40g protein
• 35g carbs

Per 100g:
• 143 calories
• 11.4g protein
• 10g carbs
```

**Diary Entries:**
```
Breakfast - Chicken Breast
150g → 247.5 calories
```

## Calculation Examples

### Logging Food to Diary
- Food: Chicken (165 cal per 100g)
- User eats: 150g
- Calculation: `165 * (150 / 100) = 247.5 calories`

### Logging Recipe to Diary
- Recipe: 350g total, 500 total calories
- User eats: 175g (half the recipe)
- Calculation: `500 * (175 / 350) = 250 calories`

### Creating a Recipe
- Ingredient 1: 200g chicken (165 cal/100g) = 330 cal
- Ingredient 2: 150g rice (111 cal/100g) = 166.5 cal
- Total recipe: 350g, 496.5 calories
- Per 100g: ~142 calories

## Validation Rules

- `quantity_grams` must be > 0
- Typical ranges:
  - Foods: 10g - 1000g (most common: 50-500g)
  - Recipe ingredients: 5g - 1000g
  - Recipe portions in diary: 50g - 2000g

## Error Messages

Update error messages to reflect gram-based system:
- ❌ "Serving size must be greater than 0"
- ✅ "Quantity in grams must be greater than 0"

## Migration Notes

- **Existing data:** All existing servings have been automatically converted (multiplied by 100)
  - Old: 1 serving → New: 100g
  - Old: 1.5 servings → New: 150g
  - Old: 2 servings → New: 200g
- **No data loss:** All historical diary entries remain accurate
- **Backwards compatibility:** None - this is a breaking change

## TypeScript Interface Updates

```typescript
// Before
interface CreateRecipeRequest {
  name: string;
  serving_size: number;
  ingredients: {
    food_id: number;
    quantity: number;
  }[];
}

// After
interface CreateRecipeRequest {
  name: string;
  ingredients: {
    food_id: number;
    quantity_grams: number;
  }[];
}

// Before
interface RecipeResponse {
  id: number;
  name: string;
  serving_size: number;
  calories: number;
  protein: number;
  // ...
}

// After
interface RecipeResponse {
  id: number;
  name: string;
  total_weight: number;
  total_calories: number;
  total_protein: number;
  total_carbs: number;
  total_fat: number;
  total_fiber: number;
  calories_per_100g: number;
  protein_per_100g: number;
  carbs_per_100g: number;
  fat_per_100g: number;
  fiber_per_100g: number;
  ingredients: IngredientWithDetails[];
}

// Before
interface CreateDiaryEntryRequest {
  food_id?: number;
  recipe_id?: number;
  date: string;
  meal_type: 'breakfast' | 'lunch' | 'dinner' | 'snack';
  serving_size: number;
  notes?: string;
}

// After
interface CreateDiaryEntryRequest {
  food_id?: number;
  recipe_id?: number;
  date: string;
  meal_type: 'breakfast' | 'lunch' | 'dinner' | 'snack';
  quantity_grams: number;
  notes?: string;
}

// Before
interface DiaryEntry {
  id: number;
  food_id?: number;
  recipe_id?: number;
  serving_size: number;
  calories: number;
  // ...
}

// After
interface DiaryEntry {
  id: number;
  food_id?: number;
  recipe_id?: number;
  quantity_grams: number;
  calories: number;
  // ...
}
```

## Testing Checklist

- [ ] Update all API client calls to use `quantity_grams` instead of `serving_size`/`quantity`
- [ ] Update TypeScript interfaces
- [ ] Test recipe creation with gram-based ingredients
- [ ] Test recipe display (verify total_weight, per-100g values)
- [ ] Test food logging with gram quantities
- [ ] Test recipe logging with gram quantities
- [ ] Test diary entry updates
- [ ] Verify all nutrition calculations are correct
- [ ] Update error message displays
- [ ] Test edge cases (very small/large quantities)
- [ ] Update any hardcoded references to "servings"
- [ ] Update user-facing labels and help text

## Questions?

If you have any questions about these changes or need clarification on the calculations, please reach out!

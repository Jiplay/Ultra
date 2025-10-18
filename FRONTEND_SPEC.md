# Frontend Developer Specification
## Ultra-Bis Nutrition Tracking App

**Version:** 1.0
**Last Updated:** January 2025
**Backend API:** http://localhost:8080

---

## Table of Contents
1. [Project Overview](#project-overview)
2. [Getting Started](#getting-started)
3. [Authentication](#authentication)
4. [API Reference](#api-reference)
5. [Data Models](#data-models)
6. [User Flows](#user-flows)
7. [Required Screens](#required-screens)
8. [UI/UX Guidelines](#uiux-guidelines)
9. [Code Examples](#code-examples)
10. [Testing Checklist](#testing-checklist)

---

## Project Overview

### Purpose
Build a comprehensive nutrition tracking application for athletes and fitness enthusiasts to monitor their food intake, track body metrics, and achieve their fitness goals.

### Key Features
- ✅ User authentication (JWT-based)
- ✅ Food database with search
- ✅ Daily meal logging (breakfast, lunch, dinner, snacks)
- ✅ Personalized nutrition goals with recommendations
- ✅ Daily nutrition summaries with goal adherence
- ✅ Body metrics tracking (weight, body fat %, muscle mass)
- ✅ Progress trends (7/30/90 days)
- ✅ Profile management

### Target Platforms
- **Web App** (Desktop & Mobile responsive)
- **Mobile App** (React Native / Flutter)
- **Progressive Web App** (PWA)

### Recommended Tech Stack

#### Web
- **Framework:** React 18+ or Vue 3+
- **State Management:** React Context + Hooks or Zustand/Pinia
- **Routing:** React Router v6 or Vue Router
- **HTTP Client:** Axios or Fetch API
- **UI Library:** Tailwind CSS, Material-UI, or Chakra UI
- **Charts:** Recharts or Chart.js
- **Date Handling:** date-fns or Day.js
- **Forms:** React Hook Form or Formik

#### Mobile
- **Framework:** React Native or Flutter
- **State Management:** Redux Toolkit or Provider
- **Navigation:** React Navigation
- **HTTP Client:** Axios
- **UI Library:** React Native Paper or NativeBase

---

## Getting Started

### Backend Setup
```bash
# Start the backend API
cd backend
docker-compose up -d

# API will be available at:
http://localhost:8080
```

### API Base URL
```javascript
const API_BASE_URL = 'http://localhost:8080';
// Production: 'https://api.yourdomain.com'
```

### Environment Variables
Create a `.env` file:
```env
VITE_API_URL=http://localhost:8080
# or for React
REACT_APP_API_URL=http://localhost:8080
```

---

## Authentication

### Overview
The API uses JWT (JSON Web Tokens) for authentication. After login/registration, you receive a token that must be included in subsequent requests.

### Token Storage
**Recommended:** Store JWT token in:
- `localStorage` (web) - easier but less secure
- `sessionStorage` (web) - more secure, clears on tab close
- `AsyncStorage` (React Native)
- Secure storage (mobile apps)

### Authentication Flow

```
1. User registers/logs in
   ↓
2. Backend returns JWT token + user data
   ↓
3. Store token in localStorage/state
   ↓
4. Include token in Authorization header for protected routes
   ↓
5. On 401 error, redirect to login
```

### Headers for Protected Routes
```javascript
headers: {
  'Authorization': `Bearer ${token}`,
  'Content-Type': 'application/json'
}
```

---

## API Reference

### Base URL
```
http://localhost:8080
```

### Authentication Endpoints

#### POST /auth/register
Register a new user.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "password123",
  "name": "John Doe"
}
```

**Response (201):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe",
    "age": 0,
    "gender": "",
    "height": 0,
    "activity_level": "moderate",
    "goal_type": "maintain",
    "created_at": "2025-01-15T10:00:00Z",
    "updated_at": "2025-01-15T10:00:00Z"
  }
}
```

**Errors:**
- `400` - Invalid request body / Password too short
- `409` - Email already registered

---

#### POST /auth/login
Login with existing account.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response (200):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe",
    // ... other user fields
  }
}
```

**Errors:**
- `401` - Invalid credentials

---

#### GET /auth/me
Get current user profile (Protected).

**Headers:**
```
Authorization: Bearer {token}
```

**Response (200):**
```json
{
  "id": 1,
  "email": "user@example.com",
  "name": "John Doe",
  "age": 28,
  "gender": "male",
  "height": 180,
  "activity_level": "active",
  "goal_type": "lose",
  "created_at": "2025-01-15T10:00:00Z",
  "updated_at": "2025-01-15T10:00:00Z"
}
```

---

#### PUT /users/profile
Update user profile (Protected).

**Request:**
```json
{
  "name": "John Updated",
  "age": 28,
  "gender": "male",
  "height": 180,
  "activity_level": "active",
  "goal_type": "lose"
}
```

**Activity Levels:**
- `sedentary` - Little or no exercise
- `light` - Light exercise 1-3 days/week
- `moderate` - Moderate exercise 3-5 days/week
- `active` - Hard exercise 6-7 days/week
- `very_active` - Very hard exercise, physical job

**Goal Types:**
- `maintain` - Maintain current weight
- `lose` - Lose weight
- `gain` - Gain weight/muscle

**Response (200):** Updated user object

---

### Food Endpoints

#### GET /foods
Get all foods (Public).

**Response (200):**
```json
[
  {
    "id": 1,
    "name": "Chicken Breast",
    "description": "Grilled skinless chicken breast",
    "calories": 165,
    "protein": 31,
    "carbs": 0,
    "fat": 3.6,
    "fiber": 0,
    "created_at": "2025-01-15T10:00:00Z",
    "updated_at": "2025-01-15T10:00:00Z"
  }
]
```

---

#### GET /foods/{id}
Get food by ID (Public).

**Response (200):** Single food object

**Errors:**
- `404` - Food not found

---

#### POST /foods
Create new food (Public).

**Request:**
```json
{
  "name": "Salmon",
  "description": "Atlantic salmon, baked",
  "calories": 206,
  "protein": 22,
  "carbs": 0,
  "fat": 13,
  "fiber": 0
}
```

**Response (201):** Created food object

---

#### PUT /foods/{id}
Update food (Public).

**Request:** Same as POST

**Response (200):** Updated food object

---

#### DELETE /foods/{id}
Delete food (Public).

**Response (204):** No content

---

### Nutrition Goals Endpoints

#### POST /goals
Create nutrition goal (Protected).

**Request:**
```json
{
  "calories": 2200,
  "protein": 165,
  "carbs": 220,
  "fat": 73,
  "fiber": 31,
  "start_date": "2025-01-15T00:00:00Z",
  "end_date": null
}
```

**Response (201):**
```json
{
  "id": 1,
  "user_id": 1,
  "calories": 2200,
  "protein": 165,
  "carbs": 220,
  "fat": 73,
  "fiber": 31,
  "start_date": "2025-01-15T00:00:00Z",
  "end_date": null,
  "is_active": true,
  "created_at": "2025-01-15T10:00:00Z",
  "updated_at": "2025-01-15T10:00:00Z"
}
```

---

#### GET /goals
Get active nutrition goal (Protected).

**Response (200):** Goal object

**Errors:**
- `404` - No active goal found

---

#### GET /goals/all
Get all nutrition goals history (Protected).

**Response (200):** Array of goal objects

---

#### POST /goals/recommended
Calculate recommended nutrition goals (Protected).

**Request:**
```json
{
  "weight": 75,
  "target_weight": 70,
  "weeks_to_goal": 8
}
```

**Response (200):**
```json
{
  "calories": 2000,
  "protein": 150,
  "carbs": 200,
  "fat": 67,
  "fiber": 28,
  "message": "Goal: Lose 5.0 kg in 8 weeks"
}
```

---

#### PUT /goals/{id}
Update nutrition goal (Protected).

**Request:**
```json
{
  "calories": 2300,
  "protein": 170,
  "carbs": 230,
  "fat": 77,
  "fiber": 32
}
```

**Response (200):** Updated goal object

---

### Diary Endpoints (Meal Logging)

#### POST /diary/entries
Log food to diary (Protected).

**Request:**
```json
{
  "food_id": 1,
  "date": "2025-01-15",
  "meal_type": "breakfast",
  "serving_size": 1.5,
  "notes": "Post-workout meal"
}
```

**Meal Types:**
- `breakfast`
- `lunch`
- `dinner`
- `snack`

**Response (201):**
```json
{
  "id": 1,
  "user_id": 1,
  "food_id": 1,
  "recipe_id": null,
  "date": "2025-01-15T00:00:00Z",
  "meal_type": "breakfast",
  "serving_size": 1.5,
  "notes": "Post-workout meal",
  "calories": 247.5,
  "protein": 46.5,
  "carbs": 0,
  "fat": 5.4,
  "fiber": 0,
  "created_at": "2025-01-15T10:00:00Z",
  "updated_at": "2025-01-15T10:00:00Z"
}
```

---

#### GET /diary/entries?date=YYYY-MM-DD
Get diary entries for a date (Protected).

**Query Parameters:**
- `date` (optional) - Format: `2025-01-15`, defaults to today

**Response (200):** Array of diary entry objects

---

#### GET /diary/summary/{date}
Get daily summary with goal adherence (Protected).

**Path Parameters:**
- `date` - Format: `2025-01-15`

**Response (200):**
```json
{
  "date": "2025-01-15",
  "total_calories": 1850.5,
  "total_protein": 145.2,
  "total_carbs": 200.3,
  "total_fat": 65.8,
  "total_fiber": 28.5,
  "goal_calories": 2200,
  "goal_protein": 165,
  "goal_carbs": 220,
  "goal_fat": 73,
  "goal_fiber": 31,
  "adherence": {
    "calories": 84.1,
    "protein": 88.0,
    "carbs": 91.0,
    "fat": 90.1,
    "fiber": 91.9
  },
  "entries": [/* array of diary entries */]
}
```

---

#### PUT /diary/entries/{id}
Update diary entry (Protected).

**Request:**
```json
{
  "serving_size": 2,
  "meal_type": "lunch",
  "notes": "Updated portion"
}
```

**Response (200):** Updated entry object

---

#### DELETE /diary/entries/{id}
Delete diary entry (Protected).

**Response (204):** No content

---

### Body Metrics Endpoints

#### POST /metrics
Log body metrics (Protected).

**Request:**
```json
{
  "date": "2025-01-15",
  "weight": 75.2,
  "body_fat_percent": 15.5,
  "muscle_mass_percent": 45.8,
  "notes": "Morning weigh-in, fasted"
}
```

**Response (201):**
```json
{
  "id": 1,
  "user_id": 1,
  "date": "2025-01-15T00:00:00Z",
  "weight": 75.2,
  "body_fat_percent": 15.5,
  "muscle_mass_percent": 45.8,
  "bmi": 23.2,
  "notes": "Morning weigh-in, fasted",
  "created_at": "2025-01-15T10:00:00Z",
  "updated_at": "2025-01-15T10:00:00Z"
}
```

---

#### GET /metrics
Get all body metrics (Protected).

**Response (200):** Array of metric objects

---

#### GET /metrics/latest
Get latest body measurement (Protected).

**Response (200):** Single metric object

---

#### GET /metrics/trends?period=7d|30d|90d
Get trend analysis (Protected).

**Query Parameters:**
- `period` - Options: `7d`, `30d`, `90d` (default: `30d`)

**Response (200):**
```json
{
  "period": "30d",
  "metrics": [/* array of metric objects */],
  "trend": {
    "weight_change": -2.5,
    "body_fat_change": -1.2,
    "muscle_mass_change": 0.8,
    "average_weight": 76.5,
    "average_body_fat": 16.2,
    "average_muscle_mass": 45.3
  }
}
```

---

#### DELETE /metrics/{id}
Delete body metric (Protected).

**Response (204):** No content

---

## Data Models

### TypeScript Interfaces

```typescript
// User
interface User {
  id: number;
  email: string;
  name: string;
  age: number;
  gender: 'male' | 'female' | '';
  height: number; // cm
  activity_level: 'sedentary' | 'light' | 'moderate' | 'active' | 'very_active';
  goal_type: 'maintain' | 'lose' | 'gain';
  created_at: string;
  updated_at: string;
}

// Food
interface Food {
  id: number;
  name: string;
  description: string;
  calories: number;
  protein: number;
  carbs: number;
  fat: number;
  fiber: number;
  created_at: string;
  updated_at: string;
}

// Nutrition Goal
interface NutritionGoal {
  id: number;
  user_id: number;
  calories: number;
  protein: number;
  carbs: number;
  fat: number;
  fiber: number;
  start_date: string;
  end_date: string | null;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

// Diary Entry
interface DiaryEntry {
  id: number;
  user_id: number;
  food_id: number | null;
  recipe_id: number | null;
  date: string;
  meal_type: 'breakfast' | 'lunch' | 'dinner' | 'snack';
  serving_size: number;
  notes: string;
  calories: number;
  protein: number;
  carbs: number;
  fat: number;
  fiber: number;
  created_at: string;
  updated_at: string;
}

// Body Metric
interface BodyMetric {
  id: number;
  user_id: number;
  date: string;
  weight: number; // kg
  body_fat_percent: number;
  muscle_mass_percent: number;
  bmi: number;
  notes: string;
  created_at: string;
  updated_at: string;
}

// Daily Summary
interface DailySummary {
  date: string;
  total_calories: number;
  total_protein: number;
  total_carbs: number;
  total_fat: number;
  total_fiber: number;
  goal_calories: number;
  goal_protein: number;
  goal_carbs: number;
  goal_fat: number;
  goal_fiber: number;
  adherence: {
    calories: number;
    protein: number;
    carbs: number;
    fat: number;
    fiber: number;
  };
  entries: DiaryEntry[];
}

// Trend Data
interface TrendData {
  period: string;
  metrics: BodyMetric[];
  trend: {
    weight_change: number;
    body_fat_change: number;
    muscle_mass_change: number;
    average_weight: number;
    average_body_fat: number;
    average_muscle_mass: number;
  };
}
```

---

## User Flows

### 1. Onboarding Flow

```
1. Landing Page
   ↓
2. Sign Up Page
   - Email validation
   - Password (min 6 chars)
   - Name
   ↓
3. Profile Setup
   - Age, Gender, Height
   - Activity Level
   - Goal Type (lose/maintain/gain)
   ↓
4. Goal Recommendation
   - Input: Current weight, Target weight, Weeks
   - Calculate: Recommended calories & macros
   ↓
5. Set Goals
   - Review recommendations
   - Adjust if needed
   - Save
   ↓
6. Dashboard (First time tour)
```

### 2. Daily Usage Flow

```
1. Login
   ↓
2. Dashboard
   - Today's summary card
   - Quick log meal button
   - Progress chart
   ↓
3. Log Meal
   - Select meal type
   - Search/select food
   - Adjust serving size
   - Add notes
   - Save
   ↓
4. View Daily Summary
   - Nutrition totals
   - Goal adherence %
   - Meal breakdown
```

### 3. Progress Tracking Flow

```
1. Dashboard
   ↓
2. Progress Tab
   ↓
3. Body Metrics
   - View chart (7/30/90 days)
   - Log new measurement
   ↓
4. Nutrition Trends
   - Average calories/macros
   - Goal adherence over time
```

---

## Required Screens

### 1. Authentication Screens

#### Login Screen
**Elements:**
- Email input
- Password input
- "Remember me" checkbox
- Login button
- "Forgot password?" link
- "Don't have an account? Sign up" link

#### Sign Up Screen
**Elements:**
- Name input
- Email input
- Password input
- Confirm password input
- Sign up button
- "Already have an account? Login" link

---

### 2. Main Application Screens

#### Dashboard / Home
**Elements:**
- Greeting: "Good morning, [Name]"
- Today's date with date picker
- Daily summary card:
  - Calories: [consumed]/[goal]
  - Progress bars for protein, carbs, fat, fiber
  - Adherence percentage
- Quick actions:
  - "Log Breakfast"
  - "Log Lunch"
  - "Log Dinner"
  - "Log Snack"
- Recent meals list
- Weekly chart (calories by day)

#### Food Database Screen
**Elements:**
- Search bar with filters
- Food list/grid with:
  - Food name
  - Calories
  - Protein/Carbs/Fat
  - Add button
- "Create Custom Food" button
- Sort options (name, calories, protein)

#### Log Meal Screen
**Elements:**
- Meal type selector (breakfast/lunch/dinner/snack)
- Date picker
- Food search/select
- Serving size input with +/- buttons
- Nutrition preview (auto-calculated)
- Notes textarea
- Save button

#### Diary Screen (Daily View)
**Elements:**
- Date picker
- Meal sections:
  - Breakfast (expandable)
  - Lunch (expandable)
  - Dinner (expandable)
  - Snacks (expandable)
- Each entry shows:
  - Food name
  - Serving size
  - Calories
  - Edit/Delete buttons
- Daily totals footer
- "Add meal" floating button

#### Goals Screen
**Elements:**
- Active goal card:
  - Calories target
  - Protein/Carbs/Fat/Fiber targets
  - Edit button
- "Get Recommendations" button
- Goal history list

#### Goal Recommendation Screen
**Elements:**
- Current weight input
- Target weight input
- Weeks to goal input
- Calculate button
- Results display:
  - Recommended calories
  - Macro breakdown
  - "Use these goals" button

#### Progress Screen
**Elements:**
- Tab selector: Body / Nutrition
- **Body Tab:**
  - Period selector (7d/30d/90d)
  - Weight chart
  - Body fat % chart
  - Muscle mass % chart
  - Trend summary card
  - "Log Measurement" button
- **Nutrition Tab:**
  - Average adherence chart
  - Calories trend
  - Macros distribution pie chart

#### Log Body Metrics Screen
**Elements:**
- Date picker
- Weight input (kg)
- Body fat % input
- Muscle mass % input
- BMI display (auto-calculated)
- Notes textarea
- Save button

#### Profile Screen
**Elements:**
- User info section:
  - Name, Email
  - Edit button
- Personal info:
  - Age, Gender, Height
  - Activity level selector
  - Goal type selector
  - Save button
- Account section:
  - Change password
  - Logout button
  - Delete account

---

## UI/UX Guidelines

### Design Principles
1. **Mobile-First** - Design for mobile, scale up for desktop
2. **Clear Hierarchy** - Most important info first
3. **Quick Actions** - Minimize taps to log meals
4. **Visual Feedback** - Show progress visually (charts, progress bars)
5. **Consistent** - Use same patterns throughout

### Color Scheme Suggestions

```css
/* Primary */
--primary: #4F46E5; /* Indigo */
--primary-dark: #4338CA;
--primary-light: #818CF8;

/* Success / Goals Met */
--success: #10B981; /* Green */

/* Warning / Close to Goal */
--warning: #F59E0B; /* Amber */

/* Error / Over Goal */
--error: #EF4444; /* Red */

/* Neutral */
--gray-50: #F9FAFB;
--gray-100: #F3F4F6;
--gray-800: #1F2937;
--gray-900: #111827;

/* Macros */
--protein: #EF4444; /* Red */
--carbs: #3B82F6; /* Blue */
--fat: #F59E0B; /* Amber */
--fiber: #10B981; /* Green */
```

### Typography
- **Headings:** Inter, SF Pro, or Roboto (700)
- **Body:** Inter, SF Pro, or Roboto (400, 500)
- **Numbers:** Tabular numbers for calories/macros

### Iconography
Recommended: Heroicons, Font Awesome, or Material Icons

**Key Icons:**
- Dashboard: Home
- Diary: Calendar/Book
- Foods: Apple/Utensils
- Progress: Chart/Trophy
- Profile: User

### Components

#### Progress Bar
```jsx
<ProgressBar
  value={consumed}
  max={goal}
  color={getColor(percentage)}
  showPercentage
/>
```

#### Macro Ring/Donut Chart
```jsx
<MacroChart
  protein={consumed.protein}
  carbs={consumed.carbs}
  fat={consumed.fat}
  proteinGoal={goal.protein}
  carbsGoal={goal.carbs}
  fatGoal={goal.fat}
/>
```

#### Meal Card
```jsx
<MealCard
  mealType="breakfast"
  entries={breakfastEntries}
  totalCalories={sumCalories}
  onAdd={() => handleAddMeal('breakfast')}
/>
```

### Responsive Breakpoints
```css
/* Mobile */
@media (max-width: 640px) { }

/* Tablet */
@media (min-width: 641px) and (max-width: 1024px) { }

/* Desktop */
@media (min-width: 1025px) { }
```

---

## Code Examples

### API Service Setup

```typescript
// api/client.ts
import axios from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor to add token
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('auth_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Response interceptor for errors
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Redirect to login
      localStorage.removeItem('auth_token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export default apiClient;
```

### Authentication Service

```typescript
// services/authService.ts
import apiClient from './api/client';

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface RegisterData {
  email: string;
  password: string;
  name: string;
}

export const authService = {
  async register(data: RegisterData) {
    const response = await apiClient.post('/auth/register', data);
    const { token, user } = response.data;
    localStorage.setItem('auth_token', token);
    return { token, user };
  },

  async login(credentials: LoginCredentials) {
    const response = await apiClient.post('/auth/login', credentials);
    const { token, user } = response.data;
    localStorage.setItem('auth_token', token);
    return { token, user };
  },

  async getCurrentUser() {
    const response = await apiClient.get('/auth/me');
    return response.data;
  },

  logout() {
    localStorage.removeItem('auth_token');
    window.location.href = '/login';
  },

  isAuthenticated() {
    return !!localStorage.getItem('auth_token');
  },
};
```

### Food Service

```typescript
// services/foodService.ts
import apiClient from './api/client';
import { Food } from '../types';

export const foodService = {
  async getAllFoods() {
    const response = await apiClient.get<Food[]>('/foods');
    return response.data;
  },

  async getFoodById(id: number) {
    const response = await apiClient.get<Food>(`/foods/${id}`);
    return response.data;
  },

  async createFood(food: Omit<Food, 'id' | 'created_at' | 'updated_at'>) {
    const response = await apiClient.post<Food>('/foods', food);
    return response.data;
  },

  async updateFood(id: number, food: Partial<Food>) {
    const response = await apiClient.put<Food>(`/foods/${id}`, food);
    return response.data;
  },

  async deleteFood(id: number) {
    await apiClient.delete(`/foods/${id}`);
  },
};
```

### Diary Service

```typescript
// services/diaryService.ts
import apiClient from './api/client';
import { DiaryEntry, DailySummary } from '../types';

export const diaryService = {
  async logEntry(entry: {
    food_id: number;
    date: string;
    meal_type: 'breakfast' | 'lunch' | 'dinner' | 'snack';
    serving_size: number;
    notes?: string;
  }) {
    const response = await apiClient.post<DiaryEntry>('/diary/entries', entry);
    return response.data;
  },

  async getEntries(date?: string) {
    const params = date ? { date } : {};
    const response = await apiClient.get<DiaryEntry[]>('/diary/entries', { params });
    return response.data;
  },

  async getDailySummary(date: string) {
    const response = await apiClient.get<DailySummary>(`/diary/summary/${date}`);
    return response.data;
  },

  async updateEntry(id: number, entry: Partial<DiaryEntry>) {
    const response = await apiClient.put<DiaryEntry>(`/diary/entries/${id}`, entry);
    return response.data;
  },

  async deleteEntry(id: number) {
    await apiClient.delete(`/diary/entries/${id}`);
  },
};
```

### React Hook Example

```typescript
// hooks/useAuth.ts
import { useState, useEffect, createContext, useContext } from 'react';
import { authService } from '../services/authService';
import { User } from '../types';

interface AuthContextType {
  user: User | null;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, name: string) => Promise<void>;
  logout: () => void;
  loading: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Check if user is logged in on mount
    const checkAuth = async () => {
      if (authService.isAuthenticated()) {
        try {
          const userData = await authService.getCurrentUser();
          setUser(userData);
        } catch (error) {
          authService.logout();
        }
      }
      setLoading(false);
    };
    checkAuth();
  }, []);

  const login = async (email: string, password: string) => {
    const { user } = await authService.login({ email, password });
    setUser(user);
  };

  const register = async (email: string, password: string, name: string) => {
    const { user } = await authService.register({ email, password, name });
    setUser(user);
  };

  const logout = () => {
    authService.logout();
    setUser(null);
  };

  return (
    <AuthContext.Provider value={{ user, login, register, logout, loading }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) throw new Error('useAuth must be used within AuthProvider');
  return context;
};
```

### Date Formatting Utilities

```typescript
// utils/date.ts
export const formatDate = (date: Date | string): string => {
  const d = typeof date === 'string' ? new Date(date) : date;
  return d.toISOString().split('T')[0]; // YYYY-MM-DD
};

export const formatDisplayDate = (date: Date | string): string => {
  const d = typeof date === 'string' ? new Date(date) : date;
  return d.toLocaleDateString('en-US', {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
};

export const isToday = (date: Date | string): boolean => {
  const d = typeof date === 'string' ? new Date(date) : date;
  const today = new Date();
  return formatDate(d) === formatDate(today);
};
```

---

## Testing Checklist

### Authentication
- [ ] Register with valid data
- [ ] Register with existing email (should fail)
- [ ] Register with weak password (should fail)
- [ ] Login with valid credentials
- [ ] Login with invalid credentials (should fail)
- [ ] Access protected route without token (should redirect)
- [ ] Token persists after page refresh
- [ ] Logout clears token

### Food Management
- [ ] View all foods
- [ ] Search/filter foods
- [ ] Create custom food
- [ ] Edit food
- [ ] Delete food

### Meal Logging
- [ ] Log meal for today
- [ ] Log meal for past date
- [ ] Change serving size (nutrition recalculates)
- [ ] Edit logged meal
- [ ] Delete logged meal
- [ ] View entries by date

### Daily Summary
- [ ] View today's summary
- [ ] View past date summary
- [ ] Adherence percentages calculate correctly
- [ ] Progress bars display correctly
- [ ] Empty state when no meals logged

### Goals
- [ ] Get personalized recommendations
- [ ] Create custom goal
- [ ] Edit goal
- [ ] View goal history
- [ ] No active goal warning

### Body Metrics
- [ ] Log weight and body composition
- [ ] View metrics history
- [ ] View 7-day trends
- [ ] View 30-day trends
- [ ] View 90-day trends
- [ ] Trend calculations are correct
- [ ] BMI auto-calculates

### Profile
- [ ] Update personal info
- [ ] Change activity level
- [ ] Change goal type
- [ ] View profile

### Edge Cases
- [ ] Handle API errors gracefully
- [ ] Show loading states
- [ ] Handle no internet connection
- [ ] Handle token expiration
- [ ] Empty states (no foods, no entries, etc.)
- [ ] Date picker edge cases
- [ ] Large numbers formatting
- [ ] Negative numbers handling

---

## Performance Considerations

### Optimization Tips
1. **Lazy Loading** - Load routes/components on demand
2. **Caching** - Cache food list, user data
3. **Debouncing** - Search inputs
4. **Pagination** - For large food lists
5. **Image Optimization** - If adding food images
6. **Bundle Size** - Code splitting

### State Management
- Use React Query or SWR for server state
- Local state for UI only
- Avoid unnecessary re-renders

---

## Accessibility

### Requirements
- [ ] Keyboard navigation
- [ ] Screen reader support
- [ ] ARIA labels
- [ ] Color contrast (WCAG AA)
- [ ] Focus indicators
- [ ] Form validation messages
- [ ] Alt text for images
- [ ] Semantic HTML

---

## Deployment Checklist

### Pre-Production
- [ ] Change API base URL to production
- [ ] Add proper error tracking (Sentry)
- [ ] Add analytics (Google Analytics, Mixpanel)
- [ ] Add loading spinners
- [ ] Add success/error toasts
- [ ] Add confirmation dialogs for destructive actions
- [ ] Test on multiple browsers
- [ ] Test on multiple devices
- [ ] Optimize bundle size
- [ ] Add PWA manifest (if web)
- [ ] Add service worker for offline

### Environment Variables
```env
VITE_API_URL=https://api.yourdomain.com
VITE_SENTRY_DSN=...
VITE_GA_ID=...
```

---

## Support & Questions

### Backend API
- **Local:** http://localhost:8080
- **Health Check:** GET /health
- **Logs:** `docker-compose logs -f api`

### Common Issues

**401 Unauthorized:**
- Token missing or expired
- Re-login required

**404 Not Found:**
- Check API endpoint URL
- Check resource ID

**500 Internal Server Error:**
- Check backend logs
- Report to backend team

---

## Next Steps

1. **Set up project** with chosen framework
2. **Implement authentication** (login/register)
3. **Create API service layer**
4. **Build dashboard** (most important screen)
5. **Implement meal logging** (core feature)
6. **Add goals & recommendations**
7. **Build progress tracking**
8. **Polish UI/UX**
9. **Test thoroughly**
10. **Deploy**

---

## Resources

- **API Testing:** Use `request/nutrition.http` file
- **Backend Repo:** Contact backend team
- **Design Assets:** [To be provided]
- **Brand Guidelines:** [To be provided]

---

**Good luck! 🚀**

If you have any questions, contact the backend team or refer to the API documentation above.

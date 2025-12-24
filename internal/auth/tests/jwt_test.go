package tests

import (
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ultra-bis/internal/auth"
)

func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name    string
		userID  uint
		email   string
		wantErr bool
	}{
		{
			name:    "Valid user credentials",
			userID:  1,
			email:   "test@example.com",
			wantErr: false,
		},
		{
			name:    "Valid user with different ID",
			userID:  100,
			email:   "admin@example.com",
			wantErr: false,
		},
		{
			name:    "Zero user ID",
			userID:  0,
			email:   "zero@example.com",
			wantErr: false, // Still valid, but edge case
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := auth.GenerateToken(tt.userID, tt.email)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				// Verify token format (JWT has 3 parts: header.payload.signature)
				assert.Regexp(t, `^[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+$`, token)
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	tests := []struct {
		name          string
		setupToken    func() string
		expectedError bool
		checkClaims   func(*testing.T, *auth.Claims)
	}{
		{
			name: "Valid token",
			setupToken: func() string {
				token, _ := auth.GenerateToken(1, "test@example.com")
				return token
			},
			expectedError: false,
			checkClaims: func(t *testing.T, claims *auth.Claims) {
				assert.Equal(t, uint(1), claims.UserID)
				assert.Equal(t, "test@example.com", claims.Email)
				assert.NotZero(t, claims.IssuedAt)
				assert.NotZero(t, claims.ExpiresAt)
			},
		},
		{
			name: "Empty token",
			setupToken: func() string {
				return ""
			},
			expectedError: true,
		},
		{
			name: "Invalid token format",
			setupToken: func() string {
				return "invalid.token.format.extra"
			},
			expectedError: true,
		},
		{
			name: "Malformed JWT",
			setupToken: func() string {
				return "not-a-jwt-token"
			},
			expectedError: true,
		},
		{
			name: "Token with wrong signature",
			setupToken: func() string {
				// Create a token signed with a different secret
				claims := &auth.Claims{
					UserID: 1,
					Email:  "test@example.com",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now()),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte("wrong-secret-key"))
				return tokenString
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setupToken()
			claims, err := auth.ValidateToken(token)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, claims)
				if tt.checkClaims != nil {
					tt.checkClaims(t, claims)
				}
			}
		})
	}
}

func TestTokenExpiration(t *testing.T) {
	// Generate a token
	token, err := auth.GenerateToken(1, "test@example.com")
	require.NoError(t, err)

	// Validate the token
	claims, err := auth.ValidateToken(token)
	require.NoError(t, err)
	require.NotNil(t, claims)

	// Check expiration is approximately 30 days from now
	expectedExpiry := time.Now().Add(30 * 24 * time.Hour)
	actualExpiry := claims.ExpiresAt.Time

	// Allow 5 second tolerance for test execution time
	timeDiff := actualExpiry.Sub(expectedExpiry)
	assert.Less(t, timeDiff.Abs(), 5*time.Second, "Token expiration should be ~30 days from now")
}

func TestTokenIssuedAt(t *testing.T) {
	// Generate a token
	token, err := auth.GenerateToken(1, "test@example.com")
	require.NoError(t, err)

	// Validate the token
	claims, err := auth.ValidateToken(token)
	require.NoError(t, err)
	require.NotNil(t, claims)

	// Check IssuedAt is approximately now
	issuedAt := claims.IssuedAt.Time
	timeSinceIssued := time.Since(issuedAt)

	// Should be issued within the last 5 seconds
	assert.Less(t, timeSinceIssued, 5*time.Second, "Token should be issued recently")
	assert.GreaterOrEqual(t, timeSinceIssued, time.Duration(0), "IssuedAt should not be in the future")
}

func TestMultipleTokenGeneration(t *testing.T) {
	// Generate multiple tokens for the same user
	token1, err := auth.GenerateToken(1, "test@example.com")
	require.NoError(t, err)

	// Sleep for 1 second to ensure different IssuedAt times
	time.Sleep(1 * time.Second)

	token2, err := auth.GenerateToken(1, "test@example.com")
	require.NoError(t, err)

	// Tokens should be different (due to different IssuedAt times)
	assert.NotEqual(t, token1, token2, "Multiple tokens for the same user should be different")

	// Both tokens should be valid
	claims1, err := auth.ValidateToken(token1)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), claims1.UserID)

	claims2, err := auth.ValidateToken(token2)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), claims2.UserID)
}

func TestTokenWithDifferentUsers(t *testing.T) {
	// Generate tokens for different users
	token1, err := auth.GenerateToken(1, "user1@example.com")
	require.NoError(t, err)

	token2, err := auth.GenerateToken(2, "user2@example.com")
	require.NoError(t, err)

	// Validate and check claims
	claims1, err := auth.ValidateToken(token1)
	require.NoError(t, err)
	assert.Equal(t, uint(1), claims1.UserID)
	assert.Equal(t, "user1@example.com", claims1.Email)

	claims2, err := auth.ValidateToken(token2)
	require.NoError(t, err)
	assert.Equal(t, uint(2), claims2.UserID)
	assert.Equal(t, "user2@example.com", claims2.Email)
}

func TestExtractTokenFromHeader(t *testing.T) {
	tests := []struct {
		name          string
		authorization string
		expectedToken string
	}{
		{
			name:          "Valid Bearer token",
			authorization: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test",
			expectedToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test",
		},
		{
			name:          "Missing Bearer prefix",
			authorization: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test",
			expectedToken: "",
		},
		{
			name:          "Empty header",
			authorization: "",
			expectedToken: "",
		},
		{
			name:          "Only 'Bearer' without token",
			authorization: "Bearer",
			expectedToken: "",
		},
		{
			name:          "Bearer with extra spaces",
			authorization: "Bearer  token-with-spaces",
			expectedToken: "", // Current implementation doesn't trim extra spaces
		},
		{
			name:          "Lowercase bearer",
			authorization: "bearer token123",
			expectedToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a real http.Request with the Authorization header
			req, err := newTestRequest(tt.authorization)
			require.NoError(t, err)

			token := auth.ExtractTokenFromHeader(req)
			assert.Equal(t, tt.expectedToken, token)
		})
	}
}

// newTestRequest creates a test HTTP request with an Authorization header
func newTestRequest(authHeader string) (*http.Request, error) {
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		return nil, err
	}
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	return req, nil
}

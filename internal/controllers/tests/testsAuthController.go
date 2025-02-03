package controllers_test

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
	"context"

	"github.com/Desk888/api/internal/controllers"
	"github.com/Desk888/api/internal/initializers"
	"github.com/Desk888/api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	testRouter *gin.Engine
	testDB     *gorm.DB
	testUser   models.User
	jwtToken   string
	resetToken string
)

// TestMain: Setup and teardown for all tests
func TestMain(m *testing.M) {
	// Initialize test environment
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	// Load test environment variables (adjust as needed)
	os.Setenv("ENV", "test")
	os.Setenv("DB", "host=localhost user=postgres password=postgres dbname=test_auth_db port=5432 sslmode=disable")
	os.Setenv("SECRET", "test-secret-123")

	// Initialize database and Redis
	initializers.InitDB()
	initializers.MigrateTables()
	initializers.InitRedis()

	// Create test user
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	testUser = models.User{
		FirstName:    "Test",
		LastName:     "User",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: string(hash),
	}
	initializers.DB.Create(&testUser)

	// Initialize Gin router
	testRouter = gin.Default()
	authGroup := testRouter.Group("/auth")
	authGroup.POST("/signup", controllers.Signup)
	authGroup.POST("/signin", controllers.Signin)
	authGroup.POST("/signout", controllers.Signout)
	authGroup.GET("/validate", controllers.Validate)
	authGroup.GET("/list_sessions", controllers.ListSessions)
	authGroup.POST("/initiate-reset", controllers.InitiatePasswordReset)
	authGroup.POST("/validate-reset-token", controllers.ValidateResetToken)
	authGroup.POST("/update-password", controllers.UpdatePassword)
}

func teardown() {
	// Clear test database
	initializers.DB.Exec("DROP TABLE users")
	initializers.DB.Exec("DROP TABLE IF EXISTS schema_migrations")

	// Clear Redis database using an appropriate context
	ctx := context.Background()
	initializers.RedisClient.FlushDB(ctx)
}


// Helper: Perform HTTP request and return response
func performRequest(r *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	requestBody, _ := json.Marshal(body)
	req, _ := http.NewRequest(method, path, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// --- Test Cases ---

// Test 1: Successful user signup
func TestSignup_Success(t *testing.T) {
	payload := map[string]interface{}{
		"firstName":   "New",
		"lastName":    "User",
		"username":    "newuser",
		"email":       "new@example.com",
		"password":    "newpassword123",
		"phoneNumber": "+1234567890",
	}

	w := performRequest(testRouter, "POST", "/auth/signup", payload)
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "User created successfully", response["message"])
}

// Test 2: Signup with duplicate email
func TestSignup_DuplicateEmail(t *testing.T) {
	payload := map[string]interface{}{
		"firstName": "Test",
		"lastName":  "User",
		"username":  "testuser2",
		"email":     "test@example.com", // Duplicate email
		"password":  "password123",
	}

	w := performRequest(testRouter, "POST", "/auth/signup", payload)
	assert.Equal(t, http.StatusConflict, w.Code)
}

// Test 3: Successful signin
func TestSignin_Success(t *testing.T) {
	payload := map[string]interface{}{
		"email":    "test@example.com",
		"password": "password123",
	}

	w := performRequest(testRouter, "POST", "/auth/signin", payload)
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEmpty(t, response["token"])
	jwtToken = response["token"].(string) // Save token for later tests
}

// Test 4: Signin with invalid credentials
func TestSignin_InvalidCredentials(t *testing.T) {
	payload := map[string]interface{}{
		"email":    "test@example.com",
		"password": "wrongpassword",
	}

	w := performRequest(testRouter, "POST", "/auth/signin", payload)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// Test 5: Validate JWT token
func TestValidateToken_Success(t *testing.T) {
	req, _ := http.NewRequest("GET", "/auth/validate", nil)
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test 6: Password reset flow (initiate → validate → update)
func TestPasswordReset_Flow(t *testing.T) {
	// Step 1: Initiate reset
	initiatePayload := map[string]interface{}{
		"email": "test@example.com",
	}
	w := performRequest(testRouter, "POST", "/auth/initiate-reset", initiatePayload)
	assert.Equal(t, http.StatusOK, w.Code)

	// Retrieve generated reset token from logs (mock implementation)
	// In real-world, parse email or database
	var user models.User
	initializers.DB.First(&user, "email = ?", "test@example.com")
	resetToken = user.PasswordResetToken
	log.Printf("Reset token: %s", resetToken)

	// Step 2: Validate token
	validatePayload := map[string]interface{}{
		"email": "test@example.com",
		"token": resetToken,
	}
	w = performRequest(testRouter, "POST", "/auth/validate-reset-token", validatePayload)
	assert.Equal(t, http.StatusOK, w.Code)

	// Step 3: Update password
	updatePayload := map[string]interface{}{
		"email":    "test@example.com",
		"token":    resetToken,
		"password": "newpassword123",
	}
	w = performRequest(testRouter, "POST", "/auth/update-password", updatePayload)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify new password works
	signinPayload := map[string]interface{}{
		"email":    "test@example.com",
		"password": "newpassword123",
	}
	w = performRequest(testRouter, "POST", "/auth/signin", signinPayload)
	assert.Equal(t, http.StatusOK, w.Code)
}

// Test 7: Signout and session invalidation
func TestSignout_Success(t *testing.T) {
	req, _ := http.NewRequest("POST", "/auth/signout", nil)
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify token is invalidated
	req, _ = http.NewRequest("GET", "/auth/validate", nil)
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	w = httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// Test 8: Expired reset token
func TestPasswordReset_ExpiredToken(t *testing.T) {
	// Force-expire the token
	var user models.User
	initializers.DB.First(&user, "email = ?", "test@example.com")
	user.PasswordResetExpiry = time.Now().Add(-1 * time.Hour) // Set expiry to past
	initializers.DB.Save(&user)

	payload := map[string]interface{}{
		"email":    "test@example.com",
		"token":    resetToken,
		"password": "newpassword123",
	}

	w := performRequest(testRouter, "POST", "/auth/update-password", payload)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
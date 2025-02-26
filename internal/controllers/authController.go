package controllers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Desk888/api/internal/initializers"
	"github.com/Desk888/api/internal/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"github.com/google/uuid"
)

const (
	maxSessionsPerUser = 5 // Maximum number of active sessions per user
	redisTimeout      = 5 * time.Second // Timeout for Redis operations
	tokenExpiry       = 24 * time.Hour // Expiry time for JWT tokens
)

/* 
	SessionData represents the data stored in a user session.
 	It includes the user's ID, username, IP address, user agent, and the time the session was created. 
*/
type SessionData struct {
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}

// Redis Helper functions
func contextWithTimeout() (context.Context, context.CancelFunc) {
	// Create a context with a timeout
	return context.WithTimeout(context.Background(), redisTimeout)
}

func generateTokenHash(token string) string {
	// Generate a SHA-256 hash of the token
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func getUserSessionKey(userID uint) string {
	// Generate a key for storing user sessions in Redis
	return fmt.Sprintf("user_sessions:%d", userID)
}

func Signup(c *gin.Context) {
	var body struct {
		FirstName    string `json:"firstName" binding:"required"`
		LastName     string `json:"lastName" binding:"required"`
		Username     string `json:"username" binding:"required"`
		Email        string `json:"email" binding:"required,email"`
		Password     string `json:"password" binding:"required,min=8"`
		PhoneNumber  string `json:"phoneNumber"`
		City         string `json:"city"`
		Country      string `json:"country"`
	}

	// Bind request body to struct for payload validation
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user
	user := models.User{
		FirstName:    body.FirstName,
		LastName:     body.LastName,
		Username:     body.Username,
		Email:        body.Email,
		PasswordHash: string(hash),
		PhoneNumber:  body.PhoneNumber,
		City: body.City,
		Country: body.Country,
	}

	// Save user to database
	result := initializers.DB.Create(&user)

	// Check for duplicate key errors (Email and Username)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key") {
			if strings.Contains(result.Error.Error(), "email") {
				c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
				return
			}
			if strings.Contains(result.Error.Error(), "username") {
				c.JSON(http.StatusConflict, gin.H{"error": "Username already taken"})
				return
			}
		}
		log.Printf("Failed to create user: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user": gin.H{
			"id":        user.ID,
			"username":  user.Username,
			"email":     user.Email,
			"firstName": user.FirstName,
			"lastName":  user.LastName,
			"phoneNumber": user.PhoneNumber,
			"city": user.City,
			"country": user.Country,
		},
	})
}

func Signin(c *gin.Context) {
	var body struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	// Bind request body to struct for payload validation
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by email
	var user models.User
	if err := initializers.DB.First(&user, "email = ?", body.Email).Error; err != nil {
		_ = bcrypt.CompareHashAndPassword([]byte("$2a$10$dummyHashForTimingAttack"), []byte(body.Password))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Compare password hashes to finalize authentication
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Create JWT token
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":      user.ID,
		"exp":      now.Add(tokenExpiry).Unix(),
		"iat":      now.Unix(),
		"username": user.Username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))

	// Error handling for token signing
	if err != nil {
		log.Printf("Error signing token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}

	// Create session data
	sessionData := SessionData{
		UserID:    user.ID,
		Username:  user.Username,
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		CreatedAt: now,
	}

	sessionJSON, err := json.Marshal(sessionData)
	if err != nil {
		log.Printf("Error marshaling session data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	// Store session data in Redis
	ctx, cancel := contextWithTimeout()
	defer cancel()

	tokenHash := generateTokenHash(tokenString)
	userSessionKey := getUserSessionKey(user.ID)

	pipe := initializers.RedisClient.TxPipeline()
	pipe.Set(ctx, tokenHash, string(sessionJSON), tokenExpiry)
	pipe.ZAdd(ctx, userSessionKey, redis.Z{
		Score:  float64(now.Unix()),
		Member: tokenHash,
	})
	pipe.ZRemRangeByRank(ctx, userSessionKey, 0, -maxSessionsPerUser-1)

	if _, err := pipe.Exec(ctx); err != nil {
		log.Printf("Redis transaction failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	// Set token as cookie
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		"Authorization",
		tokenString,
		int(tokenExpiry.Seconds()),
		"/",
		"",
		true,
		true,
	)

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user": gin.H{
			"id":        user.ID,
			"username":  user.Username,
			"email":     user.Email,
			"firstName": user.FirstName,
			"lastName":  user.LastName,
		},
	})
}

func Validate(c *gin.Context) {
	// Get token from Authorization header
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
		return
	}

	// Extract token from Authorization header
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	tokenHash := generateTokenHash(tokenString)

	// Get session data from Redis
	ctx, cancel := contextWithTimeout()
	defer cancel()

	sessionJSON, err := initializers.RedisClient.Get(ctx, tokenHash).Result()
	if err != nil {
		if err == redis.Nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired"})
			return
		}
		log.Printf("Redis error in Validate: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate session"})
		return
	}

	// Unmarshal session data
	var session SessionData
	if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
		log.Printf("Error unmarshaling session data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid session data"})
		return
	}

	// Verify JWT token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("SECRET")), nil
	})
	
	// Check for token validity
	if err != nil || !token.Valid {
		pipe := initializers.RedisClient.TxPipeline()
		pipe.Del(ctx, tokenHash)
		pipe.ZRem(ctx, getUserSessionKey(session.UserID), tokenHash)
		pipe.Exec(ctx)

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Get fresh user data from database
	var user models.User
	if err := initializers.DB.First(&user, session.UserID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Update session expiry and return user data
	pipe := initializers.RedisClient.TxPipeline()
	pipe.Expire(ctx, tokenHash, tokenExpiry)
	pipe.ZAdd(ctx, getUserSessionKey(session.UserID), redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: tokenHash,
	})
	if _, err := pipe.Exec(ctx); err != nil {
		log.Printf("Failed to update session: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":        user.ID,
			"username":  user.Username,
			"email":     user.Email,
			"firstName": user.FirstName,
			"lastName":  user.LastName,
		},
	})
}

func Signout(c *gin.Context) {
	// Get token from Authorization header
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
		return
	}
	// Extract token from Authorization header
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	tokenHash := generateTokenHash(tokenString)

	// Get session data from Redis and delete it
	ctx, cancel := contextWithTimeout()
	defer cancel()

	sessionJSON, err := initializers.RedisClient.Get(ctx, tokenHash).Result()
	if err != nil && err != redis.Nil {
		log.Printf("Redis error in Signout: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process logout"})
		return
	}
	// Clean up session data error handling
	if err == nil {
		var session SessionData
		if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
			log.Printf("Error unmarshaling session data: %v", err)
		} else {
			pipe := initializers.RedisClient.TxPipeline()
			pipe.Del(ctx, tokenHash)
			pipe.ZRem(ctx, getUserSessionKey(session.UserID), tokenHash)
			if _, err := pipe.Exec(ctx); err != nil {
				log.Printf("Failed to clean up session: %v", err)
			}
		}
	}
	// Clear token cookie
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		"Authorization",
		"",
		-1,
		"/",
		"",
		true,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully logged out",
	})
}

func ListSessions(c *gin.Context) {
	// Get user ID from JWT claims
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	
	userClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user claims"})
		return
	}
	// Get user sessions from Redis
	userID := uint(userClaims["sub"].(float64))
	ctx, cancel := contextWithTimeout()
	defer cancel()

	userSessionKey := getUserSessionKey(userID)
	tokens, err := initializers.RedisClient.ZRange(ctx, userSessionKey, 0, -1).Result()
	if err != nil {
		log.Printf("Failed to get user sessions: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve sessions"})
		return
	}
	
	// Unmarshal session data and return it
	var sessions []SessionData
	for _, tokenHash := range tokens {
		sessionJSON, err := initializers.RedisClient.Get(ctx, tokenHash).Result()
		if err != nil {
			continue
		}

		var session SessionData
		if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
			continue
		}
		sessions = append(sessions, session)
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
	})
}

// Google OAuth
func SignInWithProvider(c *gin.Context) {
	// Get the provider name from the URL
	provider := c.Param("provider")
	q := c.Request.URL.Query()
	q.Add("provider", provider)
	c.Request.URL.RawQuery = q.Encode()

	gothic.BeginAuthHandler(c.Writer, c.Request)
}

func Callback(c *gin.Context) {
	// Get the provider name from the URL
	provider := c.Param("provider")
	q := c.Request.URL.Query()
	q.Add("provider", provider)
	c.Request.URL.RawQuery = q.Encode()

	// Complete the authentication process
	_, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return 
	}

	c.Redirect(http.StatusTemporaryRedirect, "/auth/success")
}

func Success(c *gin.Context) {
	// Return success message after authentication
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully authenticated",
	})
}

func InitiatePasswordReset(c *gin.Context) {
	// Initiate password reset
	var body struct {
		Email string `json:"email" binding:"required,email"`
	}

	// Bind request body
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by email
	var user models.User
	if err := initializers.DB.First(&user, "email = ?", body.Email).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Generate reset token
	resetToken := generateResetToken()
	user.PasswordResetToken = resetToken
	user.PasswordResetExpiry = time.Now().Add(15 * time.Minute) // Token expires in 15 minutes

	// Save token and expiry to the database
	if err := initializers.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save reset token"})
		return
	}

	// Send reset token to user's email (mock implementation)
	go sendResetEmail(user.Email, resetToken)

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset token sent to your email",
	})
}

func generateResetToken() string {
	// Generate a random token
	return uuid.New().String()
}

func sendResetEmail(email, token string) {
	// TODO: Implement email sending function
}

func ValidateResetToken(c *gin.Context) {
	var body struct {
		Email string `json:"email" binding:"required,email"`
		Token string `json:"token" binding:"required"`
	}

	// Bind request body
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by email
	var user models.User
	if err := initializers.DB.First(&user, "email = ?", body.Email).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Validate reset token
	if user.PasswordResetToken != body.Token || time.Now().After(user.PasswordResetExpiry) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired reset token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Reset token is valid",
	})
}

func UpdatePassword(c *gin.Context) {
	var body struct {
		Email    string `json:"email" binding:"required,email"`
		Token    string `json:"token" binding:"required"`
		Password string `json:"password" binding:"required,min=8"`
	}

	// Bind request body
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by email
	var user models.User
	if err := initializers.DB.First(&user, "email = ?", body.Email).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Validate reset token
	if user.PasswordResetToken != body.Token || time.Now().After(user.PasswordResetExpiry) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired reset token"})
		return
	}

	// Hash new password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Update password and clear reset token
	user.PasswordHash = string(hash)
	user.PasswordResetToken = ""
	user.PasswordResetExpiry = time.Time{}

	// Save updated user
	if err := initializers.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password updated successfully",
	})
}
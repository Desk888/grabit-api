package controllers

import (
	"net/http"
	"time"
	"os"
	"strings"
	"log"
	"fmt"

	"github.com/Desk888/api/internal/initializers"
	"github.com/Desk888/api/internal/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func Signup(c *gin.Context) {
	// User Registration Functionality

    var body struct {
        FirstName    string `json:"firstName" binding:"required"`
        LastName     string `json:"lastName" binding:"required"`
        Username     string `json:"username" binding:"required"`
        Email        string `json:"email" binding:"required,email"`
        Password     string `json:"password" binding:"required,min=8"`
        PhoneNumber  string `json:"phoneNumber"`
    }

    // Bind and validate the request body
    if err := c.ShouldBindJSON(&body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Hash the password
    hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
        return
    }

    // Create the user object
    user := models.User{
        FirstName:    body.FirstName,
        LastName:     body.LastName,
        Username:     body.Username,
        Email:        body.Email,
        PasswordHash: string(hash),
        PhoneNumber:  body.PhoneNumber,
    }

    // Save the user to the database
    result := initializers.DB.Create(&user)
    if result.Error != nil {
        if strings.Contains(result.Error.Error(), "duplicate key") {
            if strings.Contains(result.Error.Error(), "email") {
                c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
				log.Printf("Email already registered")
                return
            }
            if strings.Contains(result.Error.Error(), "username") {
                c.JSON(http.StatusConflict, gin.H{"error": "Username already taken"})
				log.Printf("Username already taken")
                return
            }
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		log.Printf("Failed to create user")
        return
    }

    // Respond with success
    c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

func Signin(c *gin.Context) {
	// User Login Functionality 

    var body struct {
        Email    string `json:"email" binding:"required,email"`
        Password string `json:"password" binding:"required"`
    }

	// Check if the key / value data of payload is correct
    if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if credetentials are correct
    var user models.User
	if err := initializers.DB.First(&user, "email = ?", body.Email).Error; err != nil {
    	_ = bcrypt.CompareHashAndPassword([]byte("$2a$10$dummyHashForTimingAttack"), []byte(body.Password))
    	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
    	return
	}


    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password)); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
        return
    }

    // Create token
    now := time.Now()
    claims := jwt.MapClaims{
        "sub": user.ID,
        "exp": now.Add(time.Hour * 24).Unix(),
        "iat": now.Unix(),
        "username": user.Username,
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		log.Printf("Error signing token: %v", err)
        return
    }

    // Set cookie
    c.SetSameSite(http.SameSiteStrictMode)
    c.SetCookie(
        "Authorization",
        tokenString,
        3600*24,  // 24 hours
        "/",      // Path
        "",       // Domain
        true,     // Secure
        true,     // HttpOnly
    )

	// Return user with successful status response
    c.JSON(http.StatusOK, gin.H{
        "token": tokenString,
        "user": gin.H{
            "id": user.ID,
            "username": user.Username,
            "email": user.Email,
        },
    })
}

func Validate(c *gin.Context) {
	// Validate the user token

    // Extract the JWT token from the Authorization header
    tokenString := c.GetHeader("Authorization")
    if tokenString == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
        return
    }

    // Remove the "Bearer " prefix if it exists
    tokenString = strings.TrimPrefix(tokenString, "Bearer ")

    // Parse the token
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        // Ensure that the token's signing method is valid (HS256 in this case)
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }

        // Return the secret key for verification
        return []byte(os.Getenv("SECRET")), nil
    })

    if err != nil || !token.Valid {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
        return
    }

    // Extract the claims from the token (user ID and other claims)
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok || !token.Valid {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
        return
    }

    // Fetch user from the database using the user ID from the claims
    var user models.User
    if err := initializers.DB.First(&user, claims["sub"]).Error; err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
        return
    }

    // Return filtered user data
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
	// User signout functionality. Clear the auth cookie by setting it to expire
    c.SetSameSite(http.SameSiteStrictMode)
    c.SetCookie(
        "Authorization",  // name
        "",              // value (empty)
        -1,             // maxAge (-1 means delete immediately)
        "/",            // path
        "",             // domain
        true,           // secure
        true,           // httpOnly
    )

    c.JSON(http.StatusOK, gin.H{
        "message": "Successfully logged out",
    })
}
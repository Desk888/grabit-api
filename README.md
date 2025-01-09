
# Grabit - API Documentation
________________________________________________________________________
 Official documentation for the **Grabit** open-source project maintained by the Grabit team. This documentation is constantly being reviewed and updated by the team in correspondence with ongoing development and new features implementation.

### Quick Start Guide
________________________________________________________________________
#### Specifications:

- **Main Language:** Go `v1.23.0`
- **Development Db:** PostgresSQL 13.3+ installed locally (required for testing).
- **Production Db:** PostgreSQL 13.3 (managed separately on **Supabase**)

#### Dependencies:

Install the following packages required for the functioning of this application, using `go get` in your terminal.

```bash
go get -u gorm.io/gorm
go get -u gorm.io/driver/postgres
go get -u github.com/golang-jwt/jwt/v5
go get -u github.com/gin-gonic/gin
go get -u github.com/joho/godotenv
```

#### Docker Compose Commands

1. **Build** the images (skip cache if needed):
   ```bash
   docker-compose build --no-cache

   Start Containers: 
   docker-compose up

   Stop Containers:
   docker-compose down


All requirements can be found in the `requirements.txt` file
#### Directories Structure:

- `cmd`: main app execution with gin routers.
- `internal`: manages project utilities like middlewares, initializers, route controllers etc.
- `internal/controllers`: manage the functionality to be addressed to individual routes.
- `internal/initializers`: manages main project requirements concurrently. 
- `internal/models`: manages the database schema using GORM framework
- `internal/middleware`: runs all middleware utilities of the application.

________________________________________________________________________
### cmd / main.go

This is the main execution file, which is handling the execution of the entire application, here what the main is running:
- Initialisers (Initialise DB connection with Docker and Migrates Tables)
- Middlewares (Software that runs as an intermediary)
- HTTP routes with Gin assigned to controllers.

```go
func init() {
	/*
		* The init function is used to initialize the application helper functions
		* The helper functions are found in the internal/initializers folder
		* These functions are important for the system to function correctly
	*/
	initializers.InitDB()
	initializers.MigrateTables()
}

func main() {
	/*
		* Main program execution and nest for api routes
	*/

	r := gin.Default() // Initiliase Gin Router

	// Authentication routes
	authGroup := r.Group("/auth")
	authGroup.POST("/signup", controllers.Signup)
	authGroup.POST("/signin", controllers.Signin)
	authGroup.POST("/signout", controllers.Signout)
	authGroup.GET("/validate", middleware.RequireAuth, controllers.Validate)

	r.Run()
}
```
________________________________________________________________________
### API Features

#### Database Connection:

This is the db.go connection file that connects to a PostgreSQL instance in localhost or cloud-hosted solution using the connection string specified in the `.env` variables file. 

! *Setup your localhost database for testing* !

```go
package initializers // create unique package initializers

import (
	"gorm.io/driver/postgres" // postgres driver for db
	"gorm.io/gorm" // gorm for db config
	"log"
	"os" // os for env variables
)
  
var DB *gorm.DB // initialise gorm DB

func InitDB() {
	// Get env file db connection string
	dsn := os.Getenv("DB")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	// Throw error if connection failed
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	DB = db
}
```

________________________________________________________________________
#### Models & Model Migration:

Models are defined using the GORM framework of Go, and migrated to the database solution through the `modelsMigration.go` initialiser.

```go
package models // create unique models package

import (
	"time" // time package for handling timestamps
	"gorm.io/gorm" // import GORM framework
) 

// ENums for Conditions
const (
	ConditionUsedFair         = "Used - Fair"
	ConditionUsedGood         = "Used - Good"
	ConditionUsedExcellent    = "Used - Excellent"
	ConditionBrandNewUnboxed  = "Brand New - Unboxed"
	ConditionBrandNewSealed   = "Brand New - Used"
)

// User model
type User struct {
	gorm.Model
	ProfilePictureURL string `gorm:"size:255"` // URL to the profile picture in S3
	FirstName    string
	LastName     string
	Username     string `gorm:"uniqueIndex;not null"`
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	PhoneNumber  string
	Ads          []Ad       `gorm:"foreignKey:UserID"`
	FavouriteAds []Favorite `gorm:"foreignKey:UserID"`
}

// Ad Category model
type Category struct {
	gorm.Model
	Name string `gorm:"unique;not null"`
	Ads  []Ad   `gorm:"foreignKey:CategoryID"` 
}

// Ads model
type Ad struct {
	gorm.Model
	Title        string    `gorm:"not null"`
	Description  string    `gorm:"type:text"`
	CategoryID   uint      `gorm:"not null"` 
	UserID       uint      `gorm:"not null"` 
	Condition    string    `gorm:"not null;check:condition IN ('Used - Fair','Used - Good','Used - Excellent','Brand New - Unboxed','Brand New - Sealed')"`
	City         string
	Postcode     string
	PhoneNumber  string
	EmailAddress string
	CreatedAt    time.Time
	Category     Category `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"` 
	User         User     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	
	// @Danny See how to best implement S3 storage for the ads images
}

// Ad Favorite model
type Favorite struct {
	gorm.Model
	UserID uint `gorm:"not null"` 
	AdID   uint `gorm:"not null"`
	Ad     Ad   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	User   User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CreatedAt time.Time
}
```

Here's the structure for migrating the models in the `modelsMigration.go` file

```go
package initializers

import "github.com/Desk888/api/internal/models" // import models

func MigrateTables() {
	// Migrate all models
	DB.AutoMigrate(&models.User{})
	DB.AutoMigrate(&models.Category{})
	DB.AutoMigrate(&models.Ad{})
	DB.AutoMigrate(&models.Favorite{})
	
}
```

________________________________________________________________________
#### Authentication & Google OAuth2:

The authorization is implemented using JWT and OAuth for Google (In Development)
Further we will implement **Redis** for faster login query and and token invalidation when logging out.

##### Signup:

Register a user account

```go
func Signup(c *gin.Context) {
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
```

##### Signin:

Sign in the user to the their Grabit account by generating a session JWT token lasting 24h.

```go
func Signin(c *gin.Context) {
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
```

##### Signout:

Here is the simple implementation of cookie clearing that indicates signing out a user.
Further down the development of this application we will implement Redis to invalidate the tokens.

```go
func Signout(c *gin.Context) {
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
```
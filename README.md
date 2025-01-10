
# Grabit - API Documentation
________________________________________________________________________
 Official documentation for the **Grabit** open-source project maintained by the Grabit team. This documentation is constantly being reviewed and updated by the team in correspondence with ongoing development and new feature implementations. Feel free to contact us at grabit.application@gmail.com

### Quick Start Guide
________________________________________________________________________
#### Specifications:

- **Main Language:** Go `v1.23.0`
- **Development Db:** PostgresSQL 13.3+ installed on Docker Image.
- **Production Db:** PostgreSQL 13.3 (managed separately on **Supabase** [To be created])
- **Redis:** installed on Docker image with **Redis Commander Gui Tool**

#### Dependencies:

Install the following packages required for the functioning of this application, using `go get` in your terminal.

```bash
go get -u gorm.io/gorm
go get -u gorm.io/driver/postgres
go get -u github.com/dgrijalva/jwt-go
go get -u github.com/gin-gonic/gin
go get -u github.com/joho/godotenv
go get -u github.com/redis/go-redis/v9
go get -u golang.org/x/crypto/bcrypt

```

#### Docker Compose Commands:

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
### Main Execution | cmd / main.go:

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

#### Redis Connection:

```go
ar RedisClient *redis.Client // Initialize the Redis client

func InitRedis()  {
	// Create a new Redis client``
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "redis:6379", // Redis host (Docker)
		Password: "", // No password set
		DB:       0, // Utilise default DB
	})

	// Test the connection
	if err := RedisClient.Ping(context.Background()).Err(); err != nil {
		fmt.Println("Error connecting to Redis:", err) // remove in prod
		log.Println("Error connecting to Redis:", err)
	}
	fmt.Println("Connected to Redis successfully") // remove in prod
	log.Println("Connected to Redis successfully")
}
```

#### Models Migration:

```go
func MigrateTables() {
	// Migrate all models
	DB.AutoMigrate(&models.User{})
	DB.AutoMigrate(&models.Category{})
	DB.AutoMigrate(&models.Ad{})
	DB.AutoMigrate(&models.Favorite{})
	
}
```

________________________________________________________________________

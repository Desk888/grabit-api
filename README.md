
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

All requirements can be found in the `requirements.txt` file
#### Directories Structure:

- `cmd`: main app execution with gin routers.
- `internal`: manages project utilities like middlewares, initializers, route controllers etc.
- `internal/controllers`: manage the functionality to be addressed to individual routes.
- `internal/initializers`: manages main project requirements concurrently. 
- `internal/models`: manages the database schema using GORM framework
- `internal/middleware`: runs all middleware utilities of the application.

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

// User model example
type User struct {
	gorm.Model // identify the GORM framework to be used
	ID uint `gorm:"primaryKey"` // create fields
	FirstName string
	LastName string
	Username string `gorm:"unique"`
	Email string `gorm:"unique"`
	PasswordHash string
	PhoneNumber string
	CreatedAt time.Time
}
```

Here's the structure for migrating the models in the `modelsMigration.go` file

```go
package initializers

import "github.com/Desk888/api/internal/models" // import models

func MigrateTables() {
	// Migrate all models, User model migrated as example
	DB.AutoMigrate(&models.User{})
}
```

⚠️ All the required application models are currently in development

________________________________________________________________________



# Grabit - API Documentation
________________________________________________________________________
 Official documentation for the **Grabit** open-source project maintained by the Grabit team. This documentation is constantly being reviewed and updated by the team in correspondence with ongoing development and new feature implementations. Feel free to contact us at grabit.application@gmail.com

### Quick Start Guide
________________________________________________________________________
#### Specifications:

- **Main Language:** Go `v1.23.0`
- **Development Db:** PostgresSQL 13.3+ installed on Docker Image.
- **Production Db:** PostgreSQL 13.3 (managed separately on **Supabase** [To be created])
- **Redis:** installed on Docker image with **Redis Commander GUI Tool**

#### Dependencies:

Install the following packages required for the functioning of this application, using `go get` in your terminal.

```bash
go get -u gorm.io/gorm
go get -u gorm.io/driver/postgres
go get -u github.com/golang-jwt/jwt/v5
go get -u github.com/gin-gonic/gin
go get -u github.com/joho/godotenv
go get -u github.com/redis/go-redis/v9
go get -u github.com/markbates/goth 
go get -u golang.org/x/oauth2/google
go get -u github.com/markbates/goth/gothic@v1.80.0

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
### Main Documentation In Development ⚙️


package initializers

import (
	"log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

var DB *gorm.DB

func InitDB() {
	var err error
	connStr := "host=" + os.Getenv("DB_HOST") + 
		" port=" + os.Getenv("DB_PORT") + 
		" user=" + os.Getenv("DB_USER") + 
		" password=" + os.Getenv("DB_PASSWORD") + 
		" dbname=" + os.Getenv("DB_NAME") + 
		" sslmode=" + os.Getenv("DB_SSLMODE")
	DB, err = gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}
	
	log.Println("Connected to the postgres database successfully")
}

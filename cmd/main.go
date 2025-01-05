package main 

import (
	"github.com/Desk888/api/internal/initializers"
	"github.com/gin-gonic/gin"
	"github.com/Desk888/api/internal/middleware"
	"github.com/Desk888/api/internal/controllers"
)

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
	r := gin.Default()

	// Authentication routes
	r.POST("/signup", controllers.Signup)
	r.POST("/login", controllers.Signin)
	r.GET("/validate", middleware.RequireAuth, controllers.Validate)
	// Create logout functionality

	r.Run()
}

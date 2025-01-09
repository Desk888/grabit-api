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

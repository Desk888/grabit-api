package main

import (
	"log"
	"github.com/Desk888/api/internal/controllers"
	"github.com/Desk888/api/internal/initializers"
	"github.com/Desk888/api/internal/middleware"
	"github.com/gin-gonic/gin"
)

func init() {
	/*
	* The init function is used to initialize the application helper functions
	* The helper functions are found in the internal/initializers folder
	* These functions are important for the system to function correctly
	 */
	defer log.Println("Initializers successfully executed")

	initializers.InitDB()
	initializers.MigrateTables()
	initializers.InitRedis()
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

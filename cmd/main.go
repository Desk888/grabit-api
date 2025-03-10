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

	initializers.InitDB()         // Initialize the database
	initializers.MigrateTables()  // Migrate the database tables
	initializers.InitRedis()      // Initialize the Redis connection
	initializers.InitGoogleAuth() // Initialize the Google Auth
	initializers.InitS3() 		// Initialize the S3 connection
}

func main() {
	/*
	* Main program execution and nest for api routes
	 */

	r := gin.Default() // Initiliase Gin Router

	// Route Groups
	authGroup := r.Group("/auth")
	profileGroup := r.Group("/profile")

	// //////////////////////////

	// Standard Authentication
	authGroup.POST("/signup", controllers.Signup)
	authGroup.POST("/signin", controllers.Signin)
	authGroup.POST("/signout", controllers.Signout)
	authGroup.GET("/validate", middleware.RequireAuth, controllers.Validate)
	authGroup.GET("/list_sessions", middleware.RequireAuth, controllers.ListSessions)
	authGroup.POST("/initiate-reset", controllers.InitiatePasswordReset)
	authGroup.POST("/validate-reset-token", controllers.ValidateResetToken)
	authGroup.POST("/update-password", controllers.UpdatePassword)

	// Google Authentication
	authGroup.GET("/:provider", controllers.SignInWithProvider)
	authGroup.GET("/:provider/callback", controllers.Callback)
	authGroup.GET("/success", controllers.Success)

	// //////////////////////////

	// Profile routes
	profileGroup.GET("/:userID", middleware.RequireAuth, controllers.ViewProfile)
	profileGroup.PUT("/:userID", controllers.EditProfile)
	profileGroup.DELETE("/:userID", middleware.RequireAuth, controllers.DeleteProfile)

	r.Run()
}

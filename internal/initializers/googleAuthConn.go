package initializers

import (
	"log"
	"os"

	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
)

func InitGoogleAuth() {

	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	clientCallbackURL := os.Getenv("CLIENT_CALLBACK_URL")

	if clientID == "" || clientSecret == "" || clientCallbackURL == "" {
		log.Fatal("Missing GOOGLE_CLIENT_ID or GOOGLE_CLIENT_SECRET or GOOGLE_CLIENT_CALLBACK_URL")
	}

	goth.UseProviders(
		google.New(clientID, clientSecret, clientCallbackURL),
	)
}
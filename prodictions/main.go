package prodictions

import (
	"crypto/rsa"
	"fmt"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/marcato15/api.prodictions.com/db"
)

var (
	// DB The db instance implementing the db interface
	DB = db.SetupAws()
	// Server The server instance
	Server = echo.New()
)

// Boot The start function for the app
func Boot() {
	// Create the service's client with the session.
	//Server.Use(middleware.JWTWithConfig(middleware.JWTConfig{
	//	SigningKey:    publicKey,
	//	SigningMethod: "RS256",
	//}))
	Server.Pre(middleware.RemoveTrailingSlash())
	Server.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://prodictions.dev", "http://manage.prodictions.dev"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	//Server.Use(permissions.Permissions())
	Server.Logger.Fatal(Server.Start(":1323"))
}

func extractClaims(c echo.Context) map[string]interface{} {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	prodictionClaims := map[string]interface{}{}
	namespace := "https://api.prodictions.com/claims/"
	for key, claim := range claims {
		keyParts := strings.Split(key, namespace)
		//Check if there are multiple things, b/c that means theres a match
		if len(keyParts) == 2 {
			claimName := keyParts[1]
			prodictionClaims[claimName] = claim
		}
	}

	return prodictionClaims
}

func loadKey() *rsa.PublicKey {
	publicKeyStr := `-----BEGIN CERTIFICATE-----
	[CERTIFICATE EXPIRED]
	-----END CERTIFICATE-----`

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKeyStr))
	if err != nil {
		fmt.Println(err)
	}
	return publicKey
}

package permissions

import (
	"fmt"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

var customPermissions map[string]map[string]string

func init() {
	customPermissions = map[string]map[string]string{}
}

func RegisterCustomPermission(method string, route string, customPermission string) {
	customPermissions[route] = map[string]string{}
	customPermissions[route][method] = customPermission
}

func Permissions() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			fmt.Println("custom permissions", customPermissions)
			req := c.Request()
			//Get Claims. Potentially from multiple sources, like getClaimsFromJWT(c)
			userClaims := []string{
				"eventgroups:create",
				"eventgroups:list",
				//"admin",
			}

			isValid := validateRequest(req, userClaims)
			if isValid != true {
				return c.String(http.StatusUnauthorized, "Unauthorized Action")
			}

			if err := next(c); err != nil {
				c.Error(err)
			}
			return nil
		}
	}
}

func validateRequest(req *http.Request, userClaims []string) bool {
	requestClaim := determineClaimForRequest(req)
	return validateUserClaimsAganstRequestClaim(userClaims, requestClaim)
}

// validateUserClaimsAganstRequestClaim Determines if the user has the right claims for the request
func validateUserClaimsAganstRequestClaim(userClaims []string, resourceClaim string) bool {

	for _, userClaim := range userClaims {

		//For now assume admin can do all the things
		if userClaim == "admin" {
			return true
		}

		//Convert * in user claim to matching resource for next step
		userClaim = replaceWildcardsInUserClaim(userClaim, resourceClaim)
		//There's a match!
		if userClaim == resourceClaim {
			return true
		}

	}

	return false
}

func replaceWildcardsInUserClaim(userClaim string, resourceClaim string) string {

	resourceClaimParts := strings.Split(resourceClaim, ":")
	resourceResource := resourceClaimParts[0]
	resourceAction := resourceClaimParts[1]

	userClaimParts := strings.Split(userClaim, ":")

	//If only one part, then treat like *:action
	userResource := ""
	userAction := ""
	if len(userClaimParts) == 1 {
		userResource = resourceResource
		userAction = userClaimParts[0]
	} else {
		userResource = userClaimParts[0]
		userAction = userClaimParts[1]
	}

	if userResource == "*" {
		userResource = resourceResource
	}

	if userAction == "*" || userResource == "admin" {
		userAction = resourceAction
	}

	return userResource + ":" + userAction
}

func checkForCustomClaimForRequest(req *http.Request) string {
	path := req.URL.Path
	method := req.Method
	if path, ok := customPermissions[path]; ok {
		if claim, ok := path[method]; ok {
			return claim
		}
	}
	return ""
}

// determineClaim Determines the claim needed for the Request
func determineClaimForRequest(req *http.Request) string {

	//First check for any custom claims
	customClaim := checkForCustomClaimForRequest(req)
	if customClaim != "" {
		return customClaim
	}

	action := ""

	switch method := req.Method; method {
	case "GET":
		action = "list"
	case "POST":
		action = "create"
	case "PUT":
		action = "edit"
	case "PATCH":
		action = "edit"
	case "DELETE":
		action = "delete"
	default:
		//Not sure a better default
		action = "list"
	}

	//Determine Permissions required to complete this request
	path := req.URL.Path
	slugs := strings.Split(path, "/")

	//This is the root url. Return action only
	if len(slugs) < 1 {
		return "home:" + action
	}

	resource := slugs[1]

	return resource + ":" + action
}

func getClaimsFromJWT(c echo.Context) map[string]interface{} {
	user := c.Get("user").(*jwt.Token)
	allClaims := user.Claims.(jwt.MapClaims)
	localClaims := map[string]interface{}{}
	namespace := "https://api.prodictions.com/claims/"
	for key, claim := range allClaims {
		keyParts := strings.Split(key, namespace)
		//Check if there are multiple things, b/c that means theres a match
		if len(keyParts) == 2 {
			claimName := keyParts[1]
			localClaims[claimName] = claim
		}
	}
	return localClaims
}

package prodictions

import (
	"net/http"

	"github.com/labstack/echo"
)

func init() {
	Server.GET("/", homePage)
}

func homePage(c echo.Context) error {
	links := map[string]interface{}{
		"_links": map[string]string{
			"eventgroups": "/eventgroups",
			"events":      "/events",
			"prodictions": "/prodictions",
			"outcomes":    "/outcomes",
			"tickets":     "/tickets",
			"user":        "/user",
			"self":        "/",
		},
	}
	return c.JSON(http.StatusOK, links)
}

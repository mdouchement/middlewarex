package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mdouchement/middlewarex"
)

func main() {
	engine := echo.New()
	engine.Debug = true
	engine.Use(middleware.Recover())
	engine.Use(middleware.Logger())
	// Must be a `Pre' middleware
	engine.Pre(middlewarex.Versioning("vnd.github.v2", "vnd.github.v1", "vnd.github.v2"))

	router := engine.Group("")
	v1 := router.Group("/v1")
	v1.GET("/toto", func(c echo.Context) error {
		return c.HTML(http.StatusOK, "[v1] toto")
	})
	v1.GET("/tata", func(c echo.Context) error {
		return c.HTML(http.StatusOK, "[v1] tata")
	})

	v2 := router.Group("/v2")
	v2.GET("/toto", func(c echo.Context) error {
		return c.HTML(http.StatusOK, "[v2] toto")
	})

	printRoutes(engine)
	if err := engine.Start("localhost:6000"); err != nil {
		return
	}
}

func printRoutes(e *echo.Echo) {
	ignored := map[string]bool{
		".":     true,
		"/*":    true,
		"/v1":   true,
		"/v1/*": true,
		"/v2":   true,
		"/v2/*": true,
	}

	fmt.Println("Routes:")
	for _, route := range e.Routes() {
		if ignored[route.Path] {
			continue
		}
		fmt.Printf("%6s %s\n", route.Method, route.Path)
	}
}

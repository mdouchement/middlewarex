package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/mdouchement/middlewarex"
)

func main() {
	engine := echo.New()
	engine.Use(middleware.Recover())
	engine.Use(middleware.RequestLogger())
	// Must be a `Pre' middleware
	engine.Pre(middlewarex.Versioning("vnd.github.v2", "vnd.github.v1", "vnd.github.v2"))

	router := engine.Group("")
	v1 := router.Group("/v1")
	v1.GET("/toto", func(c *echo.Context) error {
		return c.HTML(http.StatusOK, "[v1] toto")
	})
	v1.GET("/tata", func(c *echo.Context) error {
		return c.HTML(http.StatusOK, "[v1] tata")
	})

	v2 := router.Group("/v2")
	v2.GET("/toto", func(c *echo.Context) error {
		return c.HTML(http.StatusOK, "[v2] toto")
	})

	printRoutes(engine)
	if err := engine.Start("localhost:6000"); err != nil {
		return
	}
}

func printRoutes(e *echo.Echo) {
	fmt.Println("Routes:")
	for _, route := range e.Router().Routes() {
		if route.Name == echo.NotFoundRouteName {
			continue
		}
		fmt.Printf("%6s %s\n", route.Method, route.Path)
	}
}

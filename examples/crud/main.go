package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mdouchement/middlewarex"
)

// ===== Controller
type crud1ctrl struct{}

func (ctrl *crud1ctrl) Create(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func (ctrl *crud1ctrl) List(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func (ctrl *crud1ctrl) Show(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func (ctrl *crud1ctrl) Update(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func (ctrl *crud1ctrl) Delete(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

// =====

func main() {
	engine := echo.New()
	engine.Debug = true
	engine.Use(middleware.Recover())
	engine.Use(middleware.Logger())

	router := engine.Group("")
	middlewarex.CRUD(router, "/tests", &crud1ctrl{})

	printRoutes(engine)
	if err := engine.Start("localhost:6000"); err != nil {
		return
	}
}

func printRoutes(e *echo.Echo) {
	ignored := map[string]bool{
		".":  true,
		"/*": true,
	}

	fmt.Println("Routes:")
	for _, route := range e.Routes() {
		if ignored[route.Path] {
			continue
		}
		fmt.Printf("%6s %s\n", route.Method, route.Path)
	}
}

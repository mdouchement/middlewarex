package middlewarex_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/mdouchement/middlewarex"
)

type crud1ctrl struct{}

func (_ *crud1ctrl) Create(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func (_ *crud1ctrl) List(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func (_ *crud1ctrl) Show(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func (_ *crud1ctrl) Update(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func (_ *crud1ctrl) Delete(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func TestCRUD(t *testing.T) {
	e := echo.New()
	router := e.Group("")
	middlewarex.CRUD(router, "/tests", &crud1ctrl{})

	seen := map[string]bool{}
	for _, route := range e.Routes() {
		if !strings.HasPrefix(route.Path, "/tests") {
			continue
		}

		v := fmt.Sprintf("%s %s", route.Method, route.Path)
		if seen[v] {
			t.Fatalf("Already seen %s", v)
		}

		seen[v] = true
	}

	if len(seen) != 5 {
		t.Fatalf("Expecting 4 handlers to be defined but got %d", len(seen))
	}
	if !seen["POST /tests"] {
		t.Fatal("Expecting Create handler to be defined")
	}
	if !seen["GET /tests"] {
		t.Fatal("Expecting List handler to be defined")
	}
	if !seen["GET /tests/:id"] {
		t.Fatal("Expecting Show handler to be defined")
	}
	if !seen["PATCH /tests/:id"] {
		t.Fatal("Expecting Update handler to be defined")
	}
	if !seen["DELETE /tests/:id"] {
		t.Fatal("Expecting Delete handler to be defined")
	}
}

type crud2ctrl struct{}

func (_ *crud2ctrl) List(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func (_ *crud2ctrl) Show(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func (_ *crud2ctrl) Delete(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func TestCRUDWithoutAllMethods(t *testing.T) {
	e := echo.New()
	router := e.Group("")
	middlewarex.CRUD(router, "/tests", &crud2ctrl{})

	seen := map[string]bool{}
	for _, route := range e.Routes() {
		if !strings.HasPrefix(route.Path, "/tests") {
			continue
		}

		v := fmt.Sprintf("%s %s", route.Method, route.Path)
		if seen[v] {
			t.Fatalf("Already seen %s", v)
		}

		seen[v] = true
	}

	if len(seen) != 3 {
		t.Fatalf("Expecting 4 handlers to be defined but got %d", len(seen))
	}
	if !seen["GET /tests"] {
		t.Fatal("Expecting List handler to be defined")
	}
	if !seen["GET /tests/:id"] {
		t.Fatal("Expecting Show handler to be defined")
	}
	if !seen["DELETE /tests/:id"] {
		t.Fatal("Expecting Delete handler to be defined")
	}
}

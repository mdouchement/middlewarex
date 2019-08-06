package middlewarex_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/mdouchement/middlewarex"
)

func versioningEngine() *echo.Echo {
	e := echo.New()

	v1 := e.Group("/v1")
	v1.GET("/toto", func(c echo.Context) error {
		return c.HTML(http.StatusOK, "[v1] toto")
	})
	v2 := e.Group("/v2")
	v2.GET("/toto", func(c echo.Context) error {
		return c.HTML(http.StatusOK, "[v2] toto")
	})

	return e
}

func TestVersioning(t *testing.T) {
	e := versioningEngine()
	e.Debug = true
	e.Pre(middlewarex.Versioning("vnd.myapp.v2", "vnd.myapp.v1", "vnd.myapp.v2"))

	// With the header
	req := httptest.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder() // aka ResponseWriter
	req.URL.Path = "/toto"
	req.Header.Set(middlewarex.XApplicationVersion, "vnd.myapp.v1")
	e.ServeHTTP(rec, req)
	if rec.Body.String() != "[v1] toto" {
		t.Fatalf("Expected '[v1] toto' but got '%s'", rec.Body.String())
	}
	if rec.HeaderMap.Get(middlewarex.XApplicationVersion) != "vnd.myapp.v1" {
		t.Fatalf("Expected 'vnd.myapp.v1' but got '%s'", rec.Body.String())
	}
	if rec.HeaderMap.Get(middlewarex.XApplicationStableVersion) != "vnd.myapp.v2" {
		t.Fatalf("Expected 'vnd.myapp.v2' but got '%s'", rec.Body.String())
	}

	req = httptest.NewRequest(echo.GET, "/toto", nil)
	rec = httptest.NewRecorder()
	req.URL.Path = "/toto"
	req.Header.Set(middlewarex.XApplicationVersion, "vnd.myapp.v2")
	e.ServeHTTP(rec, req)
	if rec.Body.String() != "[v2] toto" {
		t.Fatalf("Expected '[v2] toto' but got '%s'", rec.Body.String())
	}

	// With versioned path
	req = httptest.NewRequest(echo.GET, "/toto", nil)
	rec = httptest.NewRecorder()
	req.URL.Path = "/v1/toto"
	e.ServeHTTP(rec, req)
	if rec.Body.String() != "[v1] toto" {
		t.Fatalf("Expected '[v1] toto' but got '%s'", rec.Body.String())
	}

	req = httptest.NewRequest(echo.GET, "/toto", nil)
	rec = httptest.NewRecorder()
	req.URL.Path = "/v2/toto"
	e.ServeHTTP(rec, req)
	if rec.Body.String() != "[v2] toto" {
		t.Fatalf("Expected '[v2] toto' but got '%s'", rec.Body.String())
	}

	// Failing mix
	req = httptest.NewRequest(echo.GET, "/toto", nil)
	rec = httptest.NewRecorder()
	req.URL.Path = "/toto"
	req.Header.Set(middlewarex.XApplicationVersion, "vnd.myapp.v42")
	e.ServeHTTP(rec, req)
	if !strings.Contains(rec.Body.String(), "Unsuported X-Application-Version: vnd.myapp.v42") {
		t.Fatalf("Expected message 'Unsuported X-Application-Version: vnd.myapp.v42' but got '%s'", rec.Body.String())
	}

	req = httptest.NewRequest(echo.GET, "/toto", nil)
	rec = httptest.NewRecorder()
	req.URL.Path = "/v1/toto"
	req.Header.Set(middlewarex.XApplicationVersion, "vnd.myapp.v2")
	e.ServeHTTP(rec, req) // The route is rewritten here as `/v2/v1/toto`
	if !strings.Contains(rec.Body.String(), "Not Found") {
		t.Fatalf("Expected message 'Not Found' but got '%s'", rec.Body.String())
	}
}

// ------------------ //
// Benchmarks         //
// ------------------ //

// BenchmarkVersioningRW rewrites using header.
func BenchmarkVersioningRW(b *testing.B) {
	e := versioningEngine()
	e.Pre(middlewarex.Versioning("vnd.myapp.v2", "vnd.myapp.v1", "vnd.myapp.v2"))

	req := httptest.NewRequest(echo.GET, "/toto", nil)
	rec := httptest.NewRecorder() // aka ResponseWriter
	req.Header.Set(middlewarex.XApplicationVersion, "vnd.myapp.v1")
	for i := 0; i < b.N; i++ {
		e.ServeHTTP(rec, req)
	}
}

// BenchmarkVersioningRW versioned routes with the versioning middleware.
func BenchmarkVersioningVRwM(b *testing.B) {
	e := versioningEngine()
	e.Pre(middlewarex.Versioning("vnd.myapp.v2", "vnd.myapp.v1", "vnd.myapp.v2"))

	req := httptest.NewRequest(echo.GET, "/v1/toto", nil)
	rec := httptest.NewRecorder() // aka ResponseWriter
	for i := 0; i < b.N; i++ {
		e.ServeHTTP(rec, req)
	}
}

// BenchmarkVersioningRW versioned routes without the versioning middleware.
func BenchmarkVersioningVR(b *testing.B) {
	e := versioningEngine()

	req := httptest.NewRequest(echo.GET, "/v1/toto", nil)
	rec := httptest.NewRecorder() // aka ResponseWriter
	for i := 0; i < b.N; i++ {
		e.ServeHTTP(rec, req)
	}
}

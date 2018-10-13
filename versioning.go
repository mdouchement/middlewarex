package middlewarex

import (
	"fmt"
	"path"

	"github.com/labstack/echo"
)

const (
	// XApplicationVersion is the header for the asked API version (e.g. vnd.github.v1)
	XApplicationVersion = "X-Application-Version"
	// XApplicationStableVersion is the header for the stable API version (e.g. vnd.github.v3)
	XApplicationStableVersion = "X-Application-Stable-Version"
)

// Versioning rewrites routes to match the last part of the version header.
// e.g. `X-Application-Version: vnd.github.v3' header will prefix the request's path by `/v3'.
// The stable API version will be returned in the response's headers.
func Versioning(stable string, supported ...string) echo.MiddlewareFunc {
	msupported := map[string]bool{}
	prefixes := map[string]string{}
	for _, vnd := range supported {
		msupported[vnd] = true
		prefixes[vnd] = vnd2prefix(vnd)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			req := c.Request()
			vnd := req.Header.Get(XApplicationVersion)

			if vnd == "" {
				return next(c)
			}

			c.Response().Header().Set(XApplicationVersion, vnd)
			c.Response().Header().Set(XApplicationStableVersion, stable)

			if !msupported[vnd] {
				return fmt.Errorf("Unsuported %s: %s", XApplicationVersion, vnd)
			}

			req.URL.Path = prefixes[vnd] + req.URL.Path

			return next(c)
		}
	}
}

func vnd2prefix(vnd string) string {
	return path.Join("/", path.Ext(vnd)[1:])
}

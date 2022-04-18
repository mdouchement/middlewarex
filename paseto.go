package middlewarex

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/o1egl/paseto/v2"
)

type (
	// PASETOConfig defines the config for PASETO middleware.
	PASETOConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper middleware.Skipper

		// BeforeFunc defines a function which is executed just before the middleware.
		BeforeFunc middleware.BeforeFunc

		// SuccessHandler defines a function which is executed for a valid token.
		SuccessHandler PASETOSuccessHandler

		// ErrorHandler defines a function which is executed for an invalid token.
		// It may be used to define a custom PASETO error.
		ErrorHandler PASETOErrorHandler

		// ErrorHandlerWithContext is almost identical to ErrorHandler, but it's passed the current context.
		ErrorHandlerWithContext PASETOErrorHandlerWithContext

		// Signing key to validate token.
		// Required.
		SigningKey []byte

		// Validators is the list of custom validators.
		// Time validation is enforced.
		Validators []paseto.Validator

		// Context key to store user information from the token into context.
		// Optional. Default value "user".
		ContextKey string

		// TokenLookup is a string in the form of "<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		// - "param:<name>"
		// - "cookie:<name>"
		TokenLookup string

		// AuthScheme to be used in the Authorization header.
		// Optional. Default value "Bearer".
		AuthScheme string
	}

	// Token represents a PASETO JSONToken with its footer.
	Token struct {
		paseto.JSONToken
		Footer string
	}

	// PASETOSuccessHandler defines a function which is executed for a valid token.
	PASETOSuccessHandler func(echo.Context)

	// PASETOErrorHandler defines a function which is executed for an invalid token.
	PASETOErrorHandler func(error) error

	// PASETOErrorHandlerWithContext is almost identical to PASETOErrorHandler, but it's passed the current context.
	PASETOErrorHandlerWithContext func(error, echo.Context) error

	pasetoExtractor func(echo.Context) (string, error)
)

// Errors
var (
	ErrPASETOMissing     = echo.NewHTTPError(http.StatusBadRequest, "missing or malformed paseto")
	ErrPASETOUnsupported = echo.NewHTTPError(http.StatusBadRequest, "unsupported paseto version/purpose")
)

var (
	// DefaultPASETOConfig is the default PASETO auth middleware config.
	DefaultPASETOConfig = PASETOConfig{
		Skipper:     middleware.DefaultSkipper,
		ContextKey:  "paseto",
		TokenLookup: "header:" + echo.HeaderAuthorization,
		AuthScheme:  "Bearer",
		Validators:  []paseto.Validator{},
	}
)

// PASETO returns a JSON Platform-Agnostic SEcurity TOkens (PASETO) auth middleware.
//
// For valid token, it sets the user in context and calls next handler.
// For invalid token, it returns "401 - Unauthorized" error.
// For missing token, it returns "400 - Bad Request" error.
func PASETO(key []byte) echo.MiddlewareFunc {
	c := DefaultPASETOConfig
	c.SigningKey = key
	return PASETOWithConfig(c)
}

// PASETOWithConfig returns a PASETO auth middleware with config.
func PASETOWithConfig(config PASETOConfig) echo.MiddlewareFunc {
	if len(config.SigningKey) != 32 {
		panic("SigningKey must be 32 bytes length")
	}
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultPASETOConfig.Skipper
	}
	if config.ContextKey == "" {
		config.ContextKey = DefaultPASETOConfig.ContextKey
	}
	if config.Validators == nil {
		config.Validators = DefaultPASETOConfig.Validators
	}
	if config.TokenLookup == "" {
		config.TokenLookup = DefaultPASETOConfig.TokenLookup
	}
	if config.AuthScheme == "" {
		config.AuthScheme = DefaultPASETOConfig.AuthScheme
	}

	// Initialize
	parts := strings.Split(config.TokenLookup, ":")
	extractor := pasetoFromHeader(parts[1], config.AuthScheme)
	switch parts[0] {
	case "query":
		extractor = pasetoFromQuery(parts[1])
	case "param":
		extractor = pasetoFromParam(parts[1])
	case "cookie":
		extractor = pasetoFromCookie(parts[1])
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			if config.BeforeFunc != nil {
				config.BeforeFunc(c)
			}

			auth, err := extractor(c)
			if err != nil {
				if config.ErrorHandler != nil {
					return config.ErrorHandler(err)
				}

				if config.ErrorHandlerWithContext != nil {
					return config.ErrorHandlerWithContext(err, c)
				}
				return err
			}

			// TODO: support v2.public
			if !strings.HasPrefix(auth, "v2.local.") {
				if config.ErrorHandler != nil {
					return config.ErrorHandler(ErrPASETOUnsupported)
				}

				if config.ErrorHandlerWithContext != nil {
					return config.ErrorHandlerWithContext(ErrPASETOUnsupported, c)
				}
				return ErrPASETOUnsupported
			}

			var token Token
			err = paseto.Decrypt(auth, config.SigningKey, &token.JSONToken, &token.Footer)
			if err == nil {
				// Store user information from token into context.
				c.Set(config.ContextKey, token)

				err = token.Validate(append(config.Validators, paseto.ValidAt(time.Now()))...)
				if err == nil {
					if config.SuccessHandler != nil {
						config.SuccessHandler(c)
					}
					return next(c)
				}
			}

			if config.ErrorHandler != nil {
				return config.ErrorHandler(err)
			}
			if config.ErrorHandlerWithContext != nil {
				return config.ErrorHandlerWithContext(err, c)
			}
			return &echo.HTTPError{
				Code:     http.StatusUnauthorized,
				Message:  "invalid or expired paseto",
				Internal: err,
			}
		}
	}
}

// pasetoFromHeader returns a `pasetoExtractor` that extracts token from the request header.
func pasetoFromHeader(header string, authScheme string) pasetoExtractor {
	return func(c echo.Context) (string, error) {
		auth := c.Request().Header.Get(header)
		l := len(authScheme)
		if len(auth) > l+1 && auth[:l] == authScheme {
			return auth[l+1:], nil
		}
		return "", ErrPASETOMissing
	}
}

// pasetoFromQuery returns a `pasetoExtractor` that extracts token from the query string.
func pasetoFromQuery(param string) pasetoExtractor {
	return func(c echo.Context) (string, error) {
		token := c.QueryParam(param)
		if token == "" {
			return "", ErrPASETOMissing
		}
		return token, nil
	}
}

// pasetoFromParam returns a `pasetoExtractor` that extracts token from the url param string.
func pasetoFromParam(param string) pasetoExtractor {
	return func(c echo.Context) (string, error) {
		token := c.Param(param)
		if token == "" {
			return "", ErrPASETOMissing
		}
		return token, nil
	}
}

// pasetoFromCookie returns a `pasetoExtractor` that extracts token from the named cookie.
func pasetoFromCookie(name string) pasetoExtractor {
	return func(c echo.Context) (string, error) {
		cookie, err := c.Cookie(name)
		if err != nil {
			return "", ErrPASETOMissing
		}
		return cookie.Value, nil
	}
}

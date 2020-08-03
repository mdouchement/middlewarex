package middlewarex_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mdouchement/middlewarex"
	"github.com/o1egl/paseto/v2"
	"github.com/stretchr/testify/assert"
)

func TestPASETORace(t *testing.T) {
	e := echo.New()
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	key := []byte("400c48a557be10254d235cf8c506e6fe")
	h := middlewarex.PASETO(key)(handler)

	makeReq := func(token string) echo.Context {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := httptest.NewRecorder()
		req.Header.Set(echo.HeaderAuthorization, middlewarex.DefaultPASETOConfig.AuthScheme+" "+token)
		c := e.NewContext(req, res)
		assert.NoError(t, h(c))
		return c
	}

	initial := "v2.local.Q0O8UKihblHPFEjLH0r1dJKntyLDpPItRvbpC49xR_lbdc8Hfx7K4kA6TfFffTD5BAaMXiqnp1yShA"
	race := "v2.local.Y4065LEB1KT-_GlN1hnSRyH4XqMOiYD9HqftLVoZiWz520Uy1zKMya58gZBWw_SsDJeCxCy-zj0FtFZqd_OqKg"

	c := makeReq(initial)
	token := c.Get(middlewarex.DefaultPASETOConfig.ContextKey).(middlewarex.Token)
	assert.Equal(t, token.Subject, "John Doe")

	makeReq(race)
	// Initial context should still be "John Doe", not "Race Condition"
	token = c.Get(middlewarex.DefaultPASETOConfig.ContextKey).(middlewarex.Token)
	assert.Equal(t, token.Subject, "John Doe")
}

func TestPASETO(t *testing.T) {
	e := echo.New()
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	token := "v2.local.Q0O8UKihblHPFEjLH0r1dJKntyLDpPItRvbpC49xR_lbdc8Hfx7K4kA6TfFffTD5BAaMXiqnp1yShA"
	validkey := []byte("400c48a557be10254d235cf8c506e6fe")
	invalidkey := []byte("invalid-57be10254d235cf8c506e6fe")
	validAuth := middlewarex.DefaultPASETOConfig.AuthScheme + " " + token

	generate := func(tk *paseto.JSONToken) string {
		if tk.Subject == "" {
			tk.Subject = "John Doe"
		}
		s, err := paseto.Encrypt(validkey, tk, []byte{})
		assert.NoError(t, err)
		return s
	}

	tests := []struct {
		expPanic   bool
		expErrCode int // 0 for Success
		config     middlewarex.PASETOConfig
		reqURL     string // "/" if empty
		hdrAuth    string
		hdrCookie  string // test.Request doesn't provide SetCookie(); use name=val
		info       string
	}{
		{
			expPanic: true,
			info:     "No signing key provided",
		},
		{
			expPanic: true,
			config:   middlewarex.PASETOConfig{SigningKey: []byte("too small")},
			info:     "Too small signing key provided",
		},
		{
			expPanic: true,
			config:   middlewarex.PASETOConfig{SigningKey: []byte("too laaaaaaaaaaaaaaaaaaaaaaaaaaaarge")},
			info:     "Too small signing key provided",
		},
		{
			config:     middlewarex.PASETOConfig{SigningKey: invalidkey},
			hdrAuth:    validAuth,
			expErrCode: http.StatusUnauthorized,
			info:       "Invalid key",
		},
		{
			config:  middlewarex.PASETOConfig{SigningKey: validkey},
			hdrAuth: validAuth,
			info:    "Valid PASETO",
		},
		{
			config: middlewarex.PASETOConfig{
				SigningKey: validkey,
				AuthScheme: "Token",
			},
			hdrAuth: "Token" + " " + token,
			info:    "Valid PASETO with custom AuthScheme",
		},
		{
			config:     middlewarex.PASETOConfig{SigningKey: validkey},
			hdrAuth:    "v2.local.invalid-auth",
			expErrCode: http.StatusBadRequest,
			info:       "Invalid Authorization header",
		},
		{
			config:     middlewarex.PASETOConfig{SigningKey: validkey},
			hdrAuth:    "unsupported format",
			expErrCode: http.StatusBadRequest,
			info:       "Unsupported Authorization header format",
		},
		{
			config:     middlewarex.PASETOConfig{SigningKey: validkey},
			expErrCode: http.StatusBadRequest,
			info:       "Empty header auth field",
		},
		//
		// Query
		//
		{
			config: middlewarex.PASETOConfig{
				SigningKey:  validkey,
				TokenLookup: "query:paseto",
			},
			reqURL: "/?a=b&paseto=" + token,
			info:   "Valid query method",
		},
		{
			config: middlewarex.PASETOConfig{
				SigningKey:  validkey,
				TokenLookup: "query:paseto",
			},
			reqURL:     "/?a=b&pasetoxyz=" + token,
			expErrCode: http.StatusBadRequest,
			info:       "Invalid query param name",
		},
		{
			config: middlewarex.PASETOConfig{
				SigningKey:  validkey,
				TokenLookup: "query:paseto",
			},
			reqURL:     "/?a=b&paseto=v2.local.invalid-token",
			expErrCode: http.StatusUnauthorized,
			info:       "Invalid query param value",
		},
		{
			config: middlewarex.PASETOConfig{
				SigningKey:  validkey,
				TokenLookup: "query:paseto",
			},
			reqURL:     "/?a=b&paseto=invalid-token",
			expErrCode: http.StatusBadRequest,
			info:       "Unsupported Authorization header format",
		},
		{
			config: middlewarex.PASETOConfig{
				SigningKey:  validkey,
				TokenLookup: "query:paseto",
			},
			reqURL:     "/?a=b",
			expErrCode: http.StatusBadRequest,
			info:       "Empty query",
		},
		//
		// Param
		//
		{
			config: middlewarex.PASETOConfig{
				SigningKey:  validkey,
				TokenLookup: "param:paseto",
			},
			reqURL: "/" + token,
			info:   "Valid param method",
		},
		//
		// Cookie
		//
		{
			config: middlewarex.PASETOConfig{
				SigningKey:  validkey,
				TokenLookup: "cookie:paseto",
			},
			hdrCookie: "paseto=" + token,
			info:      "Valid cookie method",
		},
		{
			config: middlewarex.PASETOConfig{
				SigningKey:  validkey,
				TokenLookup: "cookie:paseto",
			},
			hdrCookie:  "paseto=v2.local.invalid-token",
			expErrCode: http.StatusUnauthorized,
			info:       "Invalid token with cookie method",
		},
		{
			config: middlewarex.PASETOConfig{
				SigningKey:  validkey,
				TokenLookup: "cookie:paseto",
			},
			hdrCookie:  "paseto=invalid-token",
			expErrCode: http.StatusBadRequest,
			info:       "Unsupported Authorization header format",
		},
		{
			config: middlewarex.PASETOConfig{
				SigningKey:  validkey,
				TokenLookup: "cookie:paseto",
			},
			expErrCode: http.StatusBadRequest,
			info:       "Empty cookie",
		},
		//
		// Timestamps validation
		//
		{
			config: middlewarex.PASETOConfig{
				SigningKey: validkey,
			},
			hdrAuth: "Bearer" + " " + generate(&paseto.JSONToken{
				Expiration: time.Now().Add(time.Hour),
			}),
			info: "Valid Expiration",
		},
		{
			config: middlewarex.PASETOConfig{
				SigningKey: validkey,
			},
			hdrAuth: "Bearer" + " " + generate(&paseto.JSONToken{
				Expiration: time.Now(),
			}),
			expErrCode: http.StatusUnauthorized,
			info:       "Invalid Expiration",
		},
		{
			config: middlewarex.PASETOConfig{
				SigningKey: validkey,
			},
			hdrAuth: "Bearer" + " " + generate(&paseto.JSONToken{
				IssuedAt: time.Now().Add(-time.Hour),
			}),
			info: "Valid IssuedAt",
		},
		{
			config: middlewarex.PASETOConfig{
				SigningKey: validkey,
			},
			hdrAuth: "Bearer" + " " + generate(&paseto.JSONToken{
				IssuedAt: time.Now().Add(time.Hour),
			}),
			expErrCode: http.StatusUnauthorized,
			info:       "Invalid IssuedAt",
		},
		{
			config: middlewarex.PASETOConfig{
				SigningKey: validkey,
			},
			hdrAuth: "Bearer" + " " + generate(&paseto.JSONToken{
				NotBefore: time.Now().Add(-time.Hour),
			}),
			info: "Valid NotBefore",
		},
		{
			config: middlewarex.PASETOConfig{
				SigningKey: validkey,
			},
			hdrAuth: "Bearer" + " " + generate(&paseto.JSONToken{
				NotBefore: time.Now().Add(time.Hour),
			}),
			expErrCode: http.StatusUnauthorized,
			info:       "Invalid NotBefore",
		},
	}

	for _, test := range tests {
		if test.reqURL == "" {
			test.reqURL = "/"
		}

		req := httptest.NewRequest(http.MethodGet, test.reqURL, nil)
		res := httptest.NewRecorder()
		req.Header.Set(echo.HeaderAuthorization, test.hdrAuth)
		req.Header.Set(echo.HeaderCookie, test.hdrCookie)
		c := e.NewContext(req, res)

		if test.reqURL == "/"+token {
			c.SetParamNames("paseto")
			c.SetParamValues(token)
		}

		if test.expPanic {
			assert.Panics(t, func() {
				middlewarex.PASETOWithConfig(test.config)
			}, test.info)
			continue
		}

		if test.expErrCode != 0 {
			h := middlewarex.PASETOWithConfig(test.config)(handler)
			he := h(c).(*echo.HTTPError)
			assert.Equal(t, test.expErrCode, he.Code, test.info)
			continue
		}

		h := middlewarex.PASETOWithConfig(test.config)(handler)
		if assert.NoError(t, h(c), test.info) {
			tk := c.Get(middlewarex.DefaultPASETOConfig.ContextKey).(middlewarex.Token)
			assert.Equal(t, "John Doe", tk.Subject, test.info)
		}
	}
}

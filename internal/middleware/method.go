package middleware

import (
	"bytes"
	"io"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
)

// MethodOverride middleware handles _method form field for DELETE operations
func MethodOverride() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Method == "POST" {
				contentType := c.Request().Header.Get("Content-Type")
				if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
					// Read body and restore it
					body, err := io.ReadAll(c.Request().Body)
					if err == nil {
						// Restore the body for later use
						c.Request().Body = io.NopCloser(bytes.NewReader(body))
						// Parse form values from body
						values, err := url.ParseQuery(string(body))
						if err == nil {
							if method := values.Get("_method"); method == "DELETE" {
								c.Request().Method = "DELETE"
							}
						}
					}
				}
			}
			return next(c)
		}
	}
}

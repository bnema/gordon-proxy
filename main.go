package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	e.Any("/github-proxy/:user", func(c echo.Context) error {
		// Handle OAuth callback
		if c.Request().Method == http.MethodGet {
			code := c.QueryParam("code")
			encodedState := c.QueryParam("state")

			// Decode the state parameter to get the original redirectDomain
			decodedState, err := base64.StdEncoding.DecodeString(encodedState)
			if err != nil {
				return c.String(http.StatusBadRequest, "Invalid state parameter")
			}

			state := string(decodedState)
			parts := strings.SplitN(state, ":", 2)
			if len(parts) != 2 || parts[0] != "redirectDomain" {
				return c.String(http.StatusBadRequest, "Invalid state format")
			}

			redirectDomain := parts[1]

			redirectURL := fmt.Sprintf("https://%s/login/oauth/callback?code=%s&state=%s",
				redirectDomain,
				url.QueryEscape(code),
				url.QueryEscape(encodedState))

			return c.Redirect(http.StatusFound, redirectURL)
		}

		// Handle Webhook
		if c.Request().Method == http.MethodPost {
			// Your webhook handling logic here
			// For demonstration, just sending a 200 OK
			return c.String(http.StatusOK, "Webhook Received")
		}

		return echo.NewHTTPError(http.StatusMethodNotAllowed, "Method not allowed")
	})

	e.Start(":3131")
}

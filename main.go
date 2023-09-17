package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	// Health check route
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "Healthy")
	})
	// Custom error handler
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		fmt.Println("Error:", err)
		e.DefaultHTTPErrorHandler(err, c)
	}

	e.Any("/github-proxy/:user", func(c echo.Context) error {
		if c.Request().Method == http.MethodGet {
			code := c.QueryParam("code")
			encodedState := c.QueryParam("state")

			payload := url.Values{}
			payload.Set("client_id", os.Getenv("GITHUB_APP_ID"))
			payload.Set("client_secret", os.Getenv("GITHUB_APP_TOKEN"))
			payload.Set("code", code)

			resp, err := http.PostForm("https://github.com/login/oauth/access_token", payload)
			if err != nil {
				return fmt.Errorf("Failed to get access token: %v", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("Failed to read access token: %v", err)
			}

			parsedQuery, err := url.ParseQuery(string(body))
			if err != nil {
				return fmt.Errorf("Failed to parse access token: %v", err)
			}

			accessToken := parsedQuery.Get("access_token")

			decodedState, err := base64.StdEncoding.DecodeString(encodedState)
			if err != nil {
				return fmt.Errorf("Invalid state parameter: %v", err)
			}

			state := string(decodedState)
			parts := strings.SplitN(state, ":", 2)
			if len(parts) != 2 || parts[0] != "redirectDomain" {
				return fmt.Errorf("Invalid state format")
			}

			redirectDomain := parts[1]

			redirectURL := fmt.Sprintf("https://%s/login/oauth/callback?code=%s&state=%s",
				redirectDomain,
				url.QueryEscape(accessToken),
				url.QueryEscape(encodedState))

			return c.Redirect(http.StatusFound, redirectURL)
		}

		if c.Request().Method == http.MethodPost {
			return c.String(http.StatusOK, "Webhook Received")
		}

		return echo.NewHTTPError(http.StatusMethodNotAllowed, "Method not allowed")
	})

	e.Start(":3131")
}

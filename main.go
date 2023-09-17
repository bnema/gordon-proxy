package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	_ "github.com/joho/godotenv/autoload"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	// Health check route
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "Healthy")
	})

	e.Any("/github-proxy/:user", func(c echo.Context) error {
		// Handle OAuth callback
		if c.Request().Method == http.MethodGet {
			code := c.QueryParam("code")
			encodedState := c.QueryParam("state")

			// Exchange code for access token
			payload := url.Values{}
			payload.Set("client_id", os.Getenv("GITHUB_APP_ID"))
			payload.Set("client_secret", os.Getenv("GITHUB_APP_TOKEN"))
			payload.Set("code", code)

			resp, err := http.PostForm("https://github.com/login/oauth/access_token", payload)
			if err != nil {
				return c.String(http.StatusInternalServerError, "Failed to get access token")
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return c.String(http.StatusInternalServerError, "Failed to read access token")
			}

			parsedQuery, err := url.ParseQuery(string(body))
			if err != nil {
				return c.String(http.StatusInternalServerError, "Failed to parse access token")
			}

			accessToken := parsedQuery.Get("access_token")

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
				url.QueryEscape(accessToken),
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

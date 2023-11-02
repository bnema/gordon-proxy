package main

import (
	"net/http"
	"os"

	"github.com/bnema/gordon-proxy/handler"
	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
)

var newClient handler.GitHubClient

func init() {
	// Define a list of required environment variables
	requiredEnvVars := []string{
		"GITHUB_APP_ID",
		"GITHUB_APP_TOKEN",
		"GITHUB_WEBHOOK_SECRET",
	}

	// Check if all required environment variables are set
	checkEnvVars(requiredEnvVars)

	// Initialize the GitHub client with environment variables
	newClient = handler.GitHubClient{
		ID:            os.Getenv("GITHUB_APP_ID"),
		Secret:        os.Getenv("GITHUB_APP_TOKEN"),
		WebhookSecret: os.Getenv("GITHUB_WEBHOOK_SECRET"),
	}
}

func main() {
	e := echo.New()

	// Health check route
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "Healthy")
	})

	// Bind the GitHub proxy endpoints
	bindGithubProxyEndpoints(e, &newClient)

	// Start the Echo server
	e.Start(":3131")
}

func checkEnvVars(vars []string) {
	for _, v := range vars {
		if os.Getenv(v) == "" {
			panic(v + " environment variable is not set")
		}
	}
}

func bindGithubProxyEndpoints(e *echo.Echo, client *handler.GitHubClient) {
	proxyGroup := e.Group("/github-proxy")
	proxyGroup.GET("/authorize", func(c echo.Context) error {
		return handler.GetGithubOAuth(c, client)
	})

	proxyGroup.GET("/callback", func(c echo.Context) error {
		return handler.GetOAuthCallback(c, client)
	})

	proxyGroup.POST("/webhooks", func(c echo.Context) error {
		return handler.PostGithubWebhook(c, client)
	})

	proxyGroup.GET("/ping", func(c echo.Context) error {
		return handler.GetInfos(c, client)
	})
}

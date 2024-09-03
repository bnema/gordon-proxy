package main

import (
	"net/http"
	"os"

	"github.com/bnema/gordon-proxy/handler"
	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	newClient       handler.GitHubClient
	requiredEnvVars = []string{
		"GITHUB_APP_ID",
		"GITHUB_APP_TOKEN",
		"GITHUB_WEBHOOK_SECRET",
	}
)

func init() {
	// Check if all required environment variables are set
	checkEnvVars(requiredEnvVars)

	// Initialize the GitHub client with environment variables
	newClient = handler.GitHubClient{
		ID:            os.Getenv("GITHUB_APP_ID"),
		Secret:        os.Getenv("GITHUB_APP_TOKEN"),
		WebhookSecret: os.Getenv("GITHUB_WEBHOOK_SECRET"),
	}

	// Touch the metadata file MetadataFilePath
	_, err := os.Stat(handler.MetadataFilePath)
	if os.IsNotExist(err) {
		_, err := os.Create(handler.MetadataFilePath)
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	e := echo.New()
	// e.Use(middleware.Logger()) disabled because it's too verbose and insecure for the users
	e.Use(middleware.Recover())
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)))

	// Health check route
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "Healthy")
	})

	// Apply CORS middleware only to the /version endpoint
	e.GET("/version", func(c echo.Context) error {
		return handler.GetLatestTags(c)
	}, middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"}, // Allow all origins
		AllowMethods: []string{echo.GET, echo.OPTIONS},
	}))

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
	proxyGroup := e.Group("/github")
	proxyGroup.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"https://github.com"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowHeaders: []string{"Content-Type", "X-Requested-With", "X-Hub-Signature-256", "X-Hub-Signature", "User-Agent"},
	}))

	proxyGroup.GET("/authorize", func(c echo.Context) error {
		handler.GetGithubOAuth(c, client)
		return nil
	})

	proxyGroup.GET("/callback", func(c echo.Context) error {
		handler.GetOAuthCallback(c, client)
		return nil
	})

	proxyGroup.POST("/webhook/newrelease", func(c echo.Context) error {
		handler.PostGithubWebhook(c, client)
		return nil
	})

	proxyGroup.POST("/device/code", func(c echo.Context) error {
		return handler.PostDeviceCode(c, client)
	})

	proxyGroup.POST("/device/token", func(c echo.Context) error {
		return handler.PostDeviceToken(c, client)
	})

}

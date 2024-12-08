package main

import (
	"net/http"
	"os"
	"time"

	"github.com/bnema/gordon-proxy/handler"
	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	newClient       handler.GitHubClient
	requiredEnvVars = []string{
		"GITHUB_APP_ID",
		"GITHUB_APP_TOKEN",
		"PORT",
	}
)

func init() {
	// Configure zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})

	// Check if all required environment variables are set
	checkEnvVars(requiredEnvVars)

	// Initialize the GitHub client with environment variables
	newClient = handler.GitHubClient{
		ID:            os.Getenv("GITHUB_APP_ID"),
		Secret:        os.Getenv("GITHUB_APP_TOKEN"),
		WebhookSecret: os.Getenv("GITHUB_WEBHOOK_SECRET"),
	}

	log.Info().Msg("Initialization completed")
}

func main() {
	// Create and start the version service
	versionService := handler.NewVersionService()
	versionService.Start()
	defer versionService.Stop()

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)))

	// Add zerolog middleware
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			start := time.Now()

			err := next(c)

			log.Info().
				Str("method", req.Method).
				Str("path", req.URL.Path).
				Int("status", res.Status).
				Dur("latency", time.Since(start)).
				Msg("Request handled")

			return err
		}
	})

	// Health check route
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "Healthy")
	})

	// Apply CORS middleware only to the /version endpoint
	e.GET("/version", func(c echo.Context) error {
		return handler.GetVersion(c)
	}, middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"}, // Allow all origins
		AllowMethods: []string{echo.GET, echo.OPTIONS},
	}))

	// Bind the GitHub proxy endpoints
	bindGithubProxyEndpoints(e, &newClient)

	// Get the port from the environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start the Echo server
	log.Info().Msgf("Starting server on :%s", port)
	if err := e.Start(":" + port); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}

func checkEnvVars(vars []string) {
	for _, v := range vars {
		if os.Getenv(v) == "" {
			log.Fatal().Str("env_var", v).Msg("Environment variable is not set")
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

	proxyGroup.POST("/device/code", func(c echo.Context) error {
		return handler.PostDeviceCode(c, client)
	})

	proxyGroup.POST("/device/token", func(c echo.Context) error {
		return handler.PostDeviceToken(c, client)
	})
}

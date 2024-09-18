package handler

import (
	"time"

	"github.com/labstack/echo/v4"
)

var (
	uptime  time.Time
	version string
)

func init() {
	uptime = time.Now()
}

type ServerInfo struct {
	Uptime  string `json:"uptime"`
	Version string `json:"version"`
}

// GetInfos returns a JSON response with the server infos
func GetInfos(c echo.Context) error {
	// Calculate uptime
	uptimeDuration := time.Since(uptime)

	// Create response struct
	info := ServerInfo{
		Uptime:  uptimeDuration.String(), // or format as needed
		Version: version,
	}

	// Return JSON response
	return c.JSON(200, info)
}

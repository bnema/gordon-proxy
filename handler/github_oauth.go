package handler

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

var (
	httpClient = &http.Client{
		Timeout: time.Second * 10,
	}
)

type GitHubClient struct {
	ID            string
	Secret        string
	WebhookSecret string
}

func GetGithubOAuth(c echo.Context, client *GitHubClient) error {
	encodedState := c.QueryParam("state")
	githubAuthURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=&state=%s",
		client.ID, url.QueryEscape(encodedState),
	)
	return c.Redirect(http.StatusFound, githubAuthURL)
}

func GetOAuthCallback(c echo.Context, client *GitHubClient) error {
	code := c.QueryParam("code")
	encodedState := c.QueryParam("state")

	payload := url.Values{
		"client_id":     {client.ID},
		"client_secret": {client.Secret},
		"code":          {code},
	}

	// Requesting access token from GitHub
	resp, err := httpClient.PostForm("https://github.com/login/oauth/access_token", payload)
	if err != nil {
		return fmt.Errorf("error while requesting access token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error while reading response body: %w", err)
	}

	parsedQuery, err := url.ParseQuery(string(body))
	if err != nil {
		return fmt.Errorf("error while parsing response body: %w", err)
	}

	accessToken := parsedQuery.Get("access_token")

	// Decode state parameter to retrieve original redirect domain
	decodedState, err := base64.StdEncoding.DecodeString(encodedState)
	if err != nil {
		return fmt.Errorf("error while decoding state: %w", err)
	}

	state := string(decodedState)
	parts := strings.SplitN(state, ":", 2)
	if len(parts) != 2 || parts[0] != "redirectDomain" {
		return fmt.Errorf("invalid state format")
	}

	// Redirecting to the original redirect domain with the access token and state
	redirectDomain := parts[1]
	redirectURL := fmt.Sprintf("%s?access_token=%s&state=%s",
		redirectDomain,
		url.QueryEscape(accessToken),
		url.QueryEscape(encodedState),
	)

	return c.Redirect(http.StatusFound, redirectURL)
}

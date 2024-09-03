package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
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

type OAuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

func GetGithubOAuth(c echo.Context, client *GitHubClient) error {
	encodedState := c.QueryParam("state")
	githubAuthURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&state=%s",
		client.ID, "https://gordon-proxy.bamen.dev/github/callback", url.QueryEscape(encodedState),
	)
	log.Info().Str("url", githubAuthURL).Msg("Redirecting to GitHub OAuth")
	return c.Redirect(http.StatusFound, githubAuthURL)
}

func GetOAuthCallback(c echo.Context, client *GitHubClient) error {
	code := c.QueryParam("code")
	encodedState := c.QueryParam("state")

	accessToken, err := exchangeCodeForToken(client, code)
	if err != nil {
		log.Error().Err(err).Msg("Error exchanging code for token")
		return fmt.Errorf("error exchanging code for token: %w", err)
	}

	redirectURL, err := buildRedirectURL(encodedState, accessToken)
	if err != nil {
		log.Error().Err(err).Msg("Error building redirect URL")
		return fmt.Errorf("error building redirect URL: %w", err)
	}

	log.Info().Str("redirect_url", redirectURL).Msg("Redirecting after OAuth callback")
	return c.Redirect(http.StatusFound, redirectURL)
}

func PostDeviceCode(c echo.Context, client *GitHubClient) error {
	payload := url.Values{
		"client_id": {client.ID},
		"scope":     {"user"},
	}

	resp, err := makePostRequest("https://github.com/login/device/code", payload)
	if err != nil {
		return fmt.Errorf("error requesting device code: %w", err)
	}

	// Ensure all required fields are present in the response
	requiredFields := []string{"device_code", "user_code", "verification_uri", "expires_in", "interval"}
	for _, field := range requiredFields {
		if _, ok := resp.(map[string]interface{})[field]; !ok {
			return fmt.Errorf("missing required field in GitHub response: %s", field)
		}
	}

	return c.JSON(http.StatusOK, resp)
}

func PostDeviceToken(c echo.Context, client *GitHubClient) error {
	deviceCode := c.FormValue("device_code")

	payload := url.Values{
		"client_id":     {client.ID},
		"device_code":   {deviceCode},
		"grant_type":    {"urn:ietf:params:oauth:grant-type:device_code"},
		"client_secret": {client.Secret},
	}

	resp, err := makePostRequest("https://github.com/login/oauth/access_token", payload)
	if err != nil {
		return fmt.Errorf("error requesting device token: %w", err)
	}

	// If the response is an OAuthResponse, convert it to a map
	if oauthResp, ok := resp.(OAuthResponse); ok {
		resp = map[string]string{
			"access_token": oauthResp.AccessToken,
			"token_type":   oauthResp.TokenType,
			"scope":        oauthResp.Scope,
		}
	}

	return c.JSON(http.StatusOK, resp)
}

func exchangeCodeForToken(client *GitHubClient, code string) (string, error) {
	payload := url.Values{
		"client_id":     {client.ID},
		"client_secret": {client.Secret},
		"code":          {code},
	}

	resp, err := makePostRequest("https://github.com/login/oauth/access_token", payload)
	if err != nil {
		return "", err
	}

	oauthResp := resp.(OAuthResponse)
	return oauthResp.AccessToken, nil
}

func buildRedirectURL(encodedState string, accessToken string) (string, error) {
	decodedState, err := base64.StdEncoding.DecodeString(encodedState)
	if err != nil {
		return "", fmt.Errorf("error decoding state: %w", err)
	}

	state := string(decodedState)
	parts := strings.SplitN(state, ":", 2)
	if len(parts) != 2 || parts[0] != "redirectDomain" {
		return "", fmt.Errorf("invalid state format")
	}

	redirectDomain := parts[1]
	return fmt.Sprintf("%s?access_token=%s&state=%s",
		redirectDomain,
		url.QueryEscape(accessToken),
		url.QueryEscape(encodedState),
	), nil
}

func makePostRequest(urlStr string, payload url.Values) (interface{}, error) {
	log.Debug().Str("url", urlStr).Interface("payload", payload).Msg("Making POST request")

	resp, err := httpClient.PostForm(urlStr, payload)
	if err != nil {
		log.Error().Err(err).Str("url", urlStr).Msg("Error making POST request")
		return nil, fmt.Errorf("error making POST request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error reading response body")
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var result interface{}
	if strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		if err := json.Unmarshal(body, &result); err != nil {
			log.Error().Err(err).Msg("Error parsing JSON response")
			return nil, fmt.Errorf("error parsing JSON response: %w", err)
		}
	} else {
		parsedQuery, err := url.ParseQuery(string(body))
		if err != nil {
			log.Error().Err(err).Msg("Error parsing query string response")
			return nil, fmt.Errorf("error parsing query string response: %w", err)
		}
		result = OAuthResponse{
			AccessToken: parsedQuery.Get("access_token"),
			TokenType:   parsedQuery.Get("token_type"),
			Scope:       parsedQuery.Get("scope"),
		}
	}

	log.Debug().Interface("result", result).Msg("POST request completed")
	return result, nil
}

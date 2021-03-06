package splunk

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
)

const (
	authContextSuffix = "/services/authentication/current-context"
)

// Client is a splunk API client
type Client struct {
	// Computed base64 auth header
	authHeader string

	config *Config
}

// Config for a SNOW connection
//
// If HTTPClient is nil, the default will be used
type Config struct {
	// Used if you want to use a custom HTTP client
	HTTPClient *http.Client

	// Base URL of your splunk instance.
	// Do not include a `/`` at the end.
	// ex: https://localhost:8089
	BaseURL string
}

// NewClient Creates and new splunk api client using the provided user/pass and config
func NewClient(ctx context.Context, username, password string, config *Config) (*Client, error) {
	configCopy := *config
	c := &Client{
		config: &configCopy,
	}
	if c.config.HTTPClient == nil {
		c.config.HTTPClient = http.DefaultClient
	}

	// Convert username:password to auth header
	c.authHeader = base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s:%s", username, password)),
	)

	// Perform simple request to make sure login is valid
	resp, err := c.BuildResponse(ctx, "GET", authContextSuffix, nil)
	if err != nil {
		return nil, fmt.Errorf("error making login request: %s", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bad login, response code: %d", resp.StatusCode)
	}

	return c, nil
}

// BuildResponse is a helper function to make a request with parameters.
//
// The suffix is whatever goes after the baseURL.  The suffix can optionally have a `/` at the beginning
func (c *Client) BuildResponse(ctx context.Context, method, suffix string, params map[string]string) (*http.Response, error) {
	if len(suffix) > 0 && suffix[0] != '/' {
		suffix = "/" + suffix
	}

	// Build URL
	URL := fmt.Sprintf("%s%s", c.config.BaseURL, suffix)

	// Build body
	body := &bytes.Buffer{}
	urlValues := url.Values{}
	urlValues.Add("output_mode", "json")
	for key, value := range params {
		urlValues.Add(key, value)
	}

	// Inject the parmaters in the request
	switch method {
	case "POST":
		// Put in body
		body.Write([]byte(urlValues.Encode()))
	default:
		// Put in URL
		URL = fmt.Sprintf("%s?%s", URL, urlValues.Encode())
	}

	// Build request
	req, err := http.NewRequestWithContext(ctx, method, URL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	return c.MakeRequest(req)
}

// MakeRequest adds authentication to the request and performs it
func (c *Client) MakeRequest(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", c.authHeader))

	return c.config.HTTPClient.Do(req)
}

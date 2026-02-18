package jira

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	apiV3Path  = "/rest/api/3"
	apiV2Path  = "/rest/api/2"
	agilePath  = "/rest/agile/1.0"
	maxRetries = 3
)

// Client is the Jira REST API client (supports Cloud and Server/DC).
type Client struct {
	baseURL      string
	authHeader   string
	httpClient   *http.Client
	instanceType InstanceType
}

// NewClient creates a new Jira API client (supports Cloud and Server/DC).
func NewClient(cfg Config) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("jira: base URL is required")
	}
	if cfg.Token == "" {
		return nil, fmt.Errorf("jira: token is required")
	}

	// Trim trailing slash from base URL.
	baseURL := strings.TrimRight(cfg.BaseURL, "/")

	// Determine auth type: if email is provided → Basic, otherwise → Bearer.
	authType := cfg.AuthType
	if authType == "" {
		if cfg.Email != "" {
			authType = AuthBasic
		} else {
			authType = AuthBearer
		}
	}

	var authHeader string
	switch authType {
	case AuthBearer:
		authHeader = "Bearer " + cfg.Token
	default: // AuthBasic
		if cfg.Email == "" {
			return nil, fmt.Errorf("jira: email is required for basic auth")
		}
		creds := cfg.Email + ":" + cfg.Token
		authHeader = "Basic " + base64.StdEncoding.EncodeToString([]byte(creds))
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}
	if cfg.InsecureSkipVerify {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return &Client{
		baseURL:      baseURL,
		authHeader:   authHeader,
		httpClient:   httpClient,
		instanceType: cfg.InstanceType,
	}, nil
}

// IsCloud returns true if this is a Jira Cloud instance.
func (c *Client) IsCloud() bool {
	return c.instanceType == InstanceCloud
}

// DetectInstanceType probes the Jira instance to determine if it's Cloud or Server/DC.
// Sets the instance type on the client and returns it.
func (c *Client) DetectInstanceType() (InstanceType, error) {
	// Try /rest/api/2/serverInfo — Server/DC returns deployment info, Cloud also supports it
	data, err := c.Get("/rest/api/2/serverInfo", nil)
	if err != nil {
		// If serverInfo fails, assume Cloud
		c.instanceType = InstanceCloud
		return InstanceCloud, nil
	}

	var info struct {
		DeploymentType string `json:"deploymentType"`
	}
	if err := json.Unmarshal(data, &info); err != nil {
		c.instanceType = InstanceCloud
		return InstanceCloud, nil
	}

	if info.DeploymentType == "Cloud" {
		c.instanceType = InstanceCloud
	} else {
		c.instanceType = InstanceServer
	}
	return c.instanceType, nil
}

// SetHTTPClient overrides the default HTTP client (useful for testing).
func (c *Client) SetHTTPClient(hc *http.Client) {
	c.httpClient = hc
}

// --- Internal HTTP helpers ---

// request builds and executes an HTTP request to the Jira API.
// path should start with / (e.g. "/rest/api/3/issue/PROJ-1").
func (c *Client) request(method, path string, query url.Values, body interface{}) ([]byte, error) {
	fullURL := c.baseURL + path
	if query != nil {
		fullURL += "?" + query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("jira: failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequest(method, fullURL, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("jira: failed to create request: %w", err)
		}

		req.Header.Set("Authorization", c.authHeader)
		req.Header.Set("Accept", "application/json")
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("jira: request failed: %w", err)
			if isNetworkError(err) {
				return nil, fmt.Errorf("%w\n\nHint: could not reach %s — check your network connection or corporate VPN", lastErr, c.baseURL)
			}
			if attempt < maxRetries {
				time.Sleep(backoff(attempt))
				// Reset body reader for retry.
				if body != nil {
					data, _ := json.Marshal(body)
					bodyReader = bytes.NewReader(data)
				}
				continue
			}
			return nil, lastErr
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("jira: failed to read response body: %w", err)
		}

		// Success.
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return respBody, nil
		}

		// Rate limited — retry after backoff.
		if resp.StatusCode == http.StatusTooManyRequests {
			lastErr = parseAPIError(resp.StatusCode, respBody)
			if attempt < maxRetries {
				time.Sleep(backoff(attempt))
				if body != nil {
					data, _ := json.Marshal(body)
					bodyReader = bytes.NewReader(data)
				}
				continue
			}
			return nil, lastErr
		}

		// Server errors — retry with backoff.
		if resp.StatusCode >= 500 {
			lastErr = parseAPIError(resp.StatusCode, respBody)
			if attempt < maxRetries {
				time.Sleep(backoff(attempt))
				if body != nil {
					data, _ := json.Marshal(body)
					bodyReader = bytes.NewReader(data)
				}
				continue
			}
			return nil, lastErr
		}

		// Client errors (4xx) — don't retry, return immediately.
		return nil, parseAPIError(resp.StatusCode, respBody)
	}

	return nil, lastErr
}

// isNetworkError checks whether the error is a network-level failure
// (DNS resolution, connection refused, timeout) where retrying won't help
// and the user likely needs to check VPN or network connectivity.
func isNetworkError(err error) bool {
	if err == nil {
		return false
	}
	var dnsErr *net.DNSError
	var opErr *net.OpError
	if errors.As(err, &dnsErr) {
		return true
	}
	if errors.As(err, &opErr) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "no such host") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "network is unreachable") ||
		strings.Contains(msg, "i/o timeout")
}

// backoff returns an exponential backoff duration for the given attempt.
func backoff(attempt int) time.Duration {
	d := time.Duration(1<<uint(attempt)) * time.Second
	if d > 60*time.Second {
		d = 60 * time.Second
	}
	return d
}

// parseAPIError attempts to parse a Jira API error response.
func parseAPIError(statusCode int, body []byte) *APIError {
	apiErr := &APIError{StatusCode: statusCode}
	if err := json.Unmarshal(body, apiErr); err != nil {
		// Could not parse — use status code + raw body.
		apiErr.ErrorMessages = []string{fmt.Sprintf("HTTP %d: %s", statusCode, string(body))}
	}
	if len(apiErr.ErrorMessages) == 0 && len(apiErr.Errors) == 0 {
		apiErr.ErrorMessages = []string{fmt.Sprintf("HTTP %d", statusCode)}
	}
	return apiErr
}

// --- Convenience methods ---

// Get performs a GET request.
func (c *Client) Get(path string, query url.Values) ([]byte, error) {
	return c.request(http.MethodGet, path, query, nil)
}

// Post performs a POST request with a JSON body.
func (c *Client) Post(path string, body interface{}) ([]byte, error) {
	return c.request(http.MethodPost, path, nil, body)
}

// Put performs a PUT request with a JSON body.
func (c *Client) Put(path string, body interface{}) ([]byte, error) {
	return c.request(http.MethodPut, path, nil, body)
}

// Delete performs a DELETE request.
func (c *Client) Delete(path string) ([]byte, error) {
	return c.request(http.MethodDelete, path, nil, nil)
}

// --- Path builders ---

// apiPath builds a v3 REST API path (Cloud default).
func apiPath(segments ...string) string {
	return apiV3Path + "/" + strings.Join(segments, "/")
}

// apiPathFor builds a REST API path using the appropriate version for the instance type.
func (c *Client) apiPathFor(segments ...string) string {
	base := apiV3Path
	if c.instanceType == InstanceServer {
		base = apiV2Path
	}
	return base + "/" + strings.Join(segments, "/")
}

// InstanceType returns the detected instance type.
func (c *Client) GetInstanceType() InstanceType {
	return c.instanceType
}

// agilePath builds an Agile REST API path.
func agileAPIPath(segments ...string) string {
	return agilePath + "/" + strings.Join(segments, "/")
}

// BaseURL returns the base URL of the Jira instance.
func (c *Client) BaseURL() string {
	return c.baseURL
}

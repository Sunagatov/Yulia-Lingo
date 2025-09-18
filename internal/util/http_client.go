package util

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var allowedHosts = map[string]bool{
	"api.mymemory.translated.net": true,
	"translate.googleapis.com":    true,
	"api.deepl.com":              true,
}

// HTTPClient provides a secure HTTP client with validation
type HTTPClient struct {
	client *http.Client
}

// NewHTTPClient creates a new secure HTTP client
func NewHTTPClient(timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  false,
				MaxIdleConnsPerHost: 2,
			},
		},
	}
}

// Do executes an HTTP request with security validation
func (c *HTTPClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	if !isValidURL(req.URL.String()) {
		return nil, fmt.Errorf("invalid or unsafe URL: %s", req.URL.String())
	}

	req = req.WithContext(ctx)
	return c.client.Do(req)
}

// CreateHTTPRequest creates a secure HTTP request with validation
func CreateHTTPRequest(ctx context.Context, method, requestURL string, body io.Reader, headers map[string]string) (*http.Request, error) {
	if !isValidURL(requestURL) {
		return nil, fmt.Errorf("invalid or unsafe URL: %s", requestURL)
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set default headers
	req.Header.Set("User-Agent", "Yulia-Lingo/1.0")
	req.Header.Set("Accept", "application/json")

	// Set custom headers
	for key, value := range headers {
		if isValidHeaderName(key) && isValidHeaderValue(value) {
			req.Header.Set(key, value)
		}
	}

	if method == http.MethodPost || method == http.MethodPut {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func isValidURL(rawURL string) bool {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	// Only allow HTTPS
	if parsedURL.Scheme != "https" {
		return false
	}

	// Validate host against allowlist
	host := strings.ToLower(parsedURL.Host)
	return allowedHosts[host]
}

func isValidHeaderName(name string) bool {
	// Basic validation for header names
	if name == "" || len(name) > 100 {
		return false
	}
	for _, r := range name {
		if r < 33 || r > 126 {
			return false
		}
	}
	return true
}

func isValidHeaderValue(value string) bool {
	// Basic validation for header values
	if len(value) > 1000 {
		return false
	}
	for _, r := range value {
		if r < 32 && r != 9 { // Allow tab character
			return false
		}
	}
	return true
}
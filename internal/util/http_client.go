package util

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

var allowedHosts = map[string]bool{
	"api.mymemory.translated.net": true,
	"translate.googleapis.com":    true,
	"api.deepl.com":              true,
}

func CreateHTTPRequest(method, requestURL string, headers map[string]string) (*http.Request, error) {
	if !isValidURL(requestURL) {
		return nil, fmt.Errorf("invalid or unsafe URL: %s", requestURL)
	}

	req, err := http.NewRequest(method, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
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

	if parsedURL.Scheme != "https" {
		return false
	}

	host := strings.ToLower(parsedURL.Host)
	return allowedHosts[host]
}
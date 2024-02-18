package util_services

import (
	"fmt"
	"net/http"
	"strings"
)

func CreatePostHTTPRequest(url, apiKey, apiHost string, payload *strings.Reader) (*http.Request, error) {
	return CreateHTTPRequest("POST", url, apiKey, apiHost, payload)
}

func CreateHTTPRequest(method, url, apiKey, apiHost string, payload *strings.Reader) (*http.Request, error) {
	var newHTTPRequest *http.Request
	var err error

	switch strings.ToUpper(method) {
	case "GET":
		newHTTPRequest, err = http.NewRequest("GET", url, nil)
	case "POST":
		newHTTPRequest, err = http.NewRequest("POST", url, payload)
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", method)
	}

	if err != nil {
		return nil, fmt.Errorf("creation of the new http request was failed: %s", err)
	}

	newHTTPRequest.Header.Add("X-RapidAPI-Key", apiKey)
	newHTTPRequest.Header.Add("X-RapidAPI-Host", apiHost)
	newHTTPRequest.Header.Add("content-type", "application/x-www-form-urlencoded")

	return newHTTPRequest, nil
}

package translate

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"Yulia-Lingo/internal/config"
	"Yulia-Lingo/internal/util"
)

type APIClient interface {
	Translate(word string) (Translation, error)
}

type client struct {
	cfg        *config.Config
	httpClient *http.Client
}

func NewAPIClient(cfg *config.Config) APIClient {
	return &client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *client) Translate(word string) (Translation, error) {
	if word == "" {
		return Translation{}, fmt.Errorf("word cannot be empty")
	}

	requestURL := c.buildURL(word)
	
	req, err := util.CreateHTTPRequest(http.MethodGet, requestURL, c.getHeaders())
	if err != nil {
		return Translation{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Translation{}, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Translation{}, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Translation{}, fmt.Errorf("failed to read response body: %w", err)
	}

	var translation Translation
	if err := json.Unmarshal(body, &translation); err != nil {
		return Translation{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return translation, nil
}

func (c *client) buildURL(word string) string {
	baseURL := c.cfg.Translate.APIURL
	if baseURL == "" {
		baseURL = "https://api.mymemory.translated.net/get"
	}

	params := url.Values{}
	params.Add("q", word)
	params.Add("langpair", "en|ru")

	return baseURL + "?" + params.Encode()
}

func (c *client) getHeaders() map[string]string {
	headers := make(map[string]string)
	
	if c.cfg.Translate.APIKey != "" {
		headers["X-RapidAPI-Key"] = c.cfg.Translate.APIKey
	}
	
	if c.cfg.Translate.APIHost != "" {
		headers["X-RapidAPI-Host"] = c.cfg.Translate.APIHost
	}

	return headers
}
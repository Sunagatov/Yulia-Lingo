package translate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"Yulia-Lingo/internal/config"
	"Yulia-Lingo/internal/logger"
	"Yulia-Lingo/internal/util"
)

const (
	maxWordLength     = 50
	maxResponseSize   = 1024 * 1024 // 1MB
	defaultAPIURL     = "https://api.mymemory.translated.net/get"
	userAgent         = "Yulia-Lingo/1.0"
	maxRetries        = 3
	retryDelay        = time.Second
)

var (
	allowedHosts = map[string]bool{
		"api.mymemory.translated.net": true,
		"translate.googleapis.com":    true,
	}
	wordRegex = regexp.MustCompile(`^[a-zA-Z\s'-]+$`)
)

type APIClient interface {
	Translate(ctx context.Context, word string) (Translation, error)
}

type client struct {
	cfg        *config.Config
	httpClient *http.Client
	log        logger.Logger
}

func NewAPIClient(cfg *config.Config, log logger.Logger) APIClient {
	return &client{
		cfg: cfg,
		log: log,
		httpClient: &http.Client{
			Timeout: cfg.Translate.Timeout,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  false,
				MaxIdleConnsPerHost: 2,
			},
		},
	}
}

func (c *client) Translate(ctx context.Context, word string) (Translation, error) {
	if err := c.validateWord(word); err != nil {
		return Translation{}, fmt.Errorf("invalid word: %w", err)
	}

	word = util.SanitizeString(word)

	for attempt := 0; attempt < maxRetries; attempt++ {
		translation, err := c.doTranslate(ctx, word)
		if err == nil {
			return translation, nil
		}

		c.log.Warn(ctx, "Translation attempt failed",
			logger.Field{Key: "attempt", Value: attempt + 1},
			logger.Field{Key: "word", Value: word},
			logger.Field{Key: "error", Value: err.Error()},
		)

		if attempt < maxRetries-1 {
			select {
			case <-ctx.Done():
				return Translation{}, ctx.Err()
			case <-time.After(retryDelay * time.Duration(attempt+1)):
			}
		}
	}

	return Translation{}, fmt.Errorf("translation failed after %d attempts", maxRetries)
}

func (c *client) doTranslate(ctx context.Context, word string) (Translation, error) {
	requestURL, err := c.buildURL(word)
	if err != nil {
		return Translation{}, fmt.Errorf("failed to build URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return Translation{}, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Translation{}, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Translation{}, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := c.readLimitedBody(resp.Body)
	if err != nil {
		return Translation{}, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse MyMemory API response
	var apiResponse struct {
		ResponseData struct {
			TranslatedText string `json:"translatedText"`
		} `json:"responseData"`
		Matches []struct {
			Translation string `json:"translation"`
		} `json:"matches"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		// Fallback: create a simple translation
		return Translation{
			Dictionary: []DictionaryEntry{
				{
					PartOfSpeech: "неизвестно",
					Terms:        []string{"Перевод недоступен"},
				},
			},
		}, nil
	}

	// Build translation from API response
	var terms []string
	if apiResponse.ResponseData.TranslatedText != "" {
		terms = append(terms, apiResponse.ResponseData.TranslatedText)
	}
	for _, match := range apiResponse.Matches {
		if match.Translation != "" && len(terms) < 3 {
			terms = append(terms, match.Translation)
		}
	}

	if len(terms) == 0 {
		terms = []string{"Перевод не найден"}
	}

	translation := Translation{
		Dictionary: []DictionaryEntry{
			{
				PartOfSpeech: "слово",
				Terms:        terms,
			},
		},
	}

	return translation, nil
}

func (c *client) validateWord(word string) error {
	if word == "" {
		return fmt.Errorf("word cannot be empty")
	}
	if len(word) > maxWordLength {
		return fmt.Errorf("word too long (max %d characters)", maxWordLength)
	}
	if !wordRegex.MatchString(word) {
		return fmt.Errorf("word contains invalid characters")
	}
	return nil
}

func (c *client) buildURL(word string) (string, error) {
	baseURL := c.cfg.Translate.APIURL
	if baseURL == "" {
		baseURL = defaultAPIURL
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid API URL: %w", err)
	}

	// Security: Validate allowed hosts
	if !allowedHosts[parsedURL.Host] {
		return "", fmt.Errorf("host not allowed: %s", parsedURL.Host)
	}

	params := url.Values{}
	params.Add("q", word)
	params.Add("langpair", "en|ru")
	parsedURL.RawQuery = params.Encode()

	return parsedURL.String(), nil
}

func (c *client) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip, deflate")

	if c.cfg.Translate.APIKey != "" {
		req.Header.Set("X-RapidAPI-Key", c.cfg.Translate.APIKey)
	}

	if c.cfg.Translate.APIHost != "" {
		req.Header.Set("X-RapidAPI-Host", c.cfg.Translate.APIHost)
	}
}

func (c *client) readLimitedBody(body io.Reader) ([]byte, error) {
	limitedReader := io.LimitReader(body, maxResponseSize)
	return io.ReadAll(limitedReader)
}
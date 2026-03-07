package translate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"

	"Yulia-Lingo/internal/config"
	"Yulia-Lingo/internal/logger"
)

const (
	maxResponseSize = 1024 * 1024
	defaultAPIURL   = "https://api.mymemory.translated.net/get"
	maxRetries      = 3
)

var allowedHosts = map[string]bool{
	"api.mymemory.translated.net": true,
	"translate.googleapis.com":    true,
}

type APIClient interface {
	Translate(ctx context.Context, word string, targetLang string) (Translation, error)
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
				MaxIdleConnsPerHost: 2,
			},
		},
	}
}

func (c *client) Translate(ctx context.Context, word string, targetLang string) (Translation, error) {
	if !isValidWord(word) {
		return Translation{}, fmt.Errorf("invalid word: %s", word)
	}
	word = strings.TrimSpace(word)

	var lastErr error
	for attempt := range maxRetries {
		t, err := c.doTranslate(ctx, word, targetLang)
		if err == nil {
			return t, nil
		}
		lastErr = err
		c.log.Warn(ctx, "translate.attempt_failed",
			logger.Field{Key: "attempt", Value: attempt + 1},
			logger.Field{Key: "error", Value: err.Error()},
		)
		if attempt < maxRetries-1 {
			timer := time.NewTimer(time.Duration(attempt+1) * time.Second)
			select {
			case <-ctx.Done():
				timer.Stop()
				return Translation{}, ctx.Err()
			case <-timer.C:
			}
		}
	}
	return Translation{}, fmt.Errorf("translation failed after %d attempts: %w", maxRetries, lastErr)
}

func (c *client) doTranslate(ctx context.Context, word, targetLang string) (Translation, error) {
	reqURL, err := c.buildURL(word, targetLang)
	if err != nil {
		return Translation{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return Translation{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "Yulia-Lingo/1.0")
	req.Header.Set("Accept", "application/json")
	if c.cfg.Translate.APIKey != "" {
		req.Header.Set("X-RapidAPI-Key", c.cfg.Translate.APIKey)
	}
	if c.cfg.Translate.APIHost != "" {
		req.Header.Set("X-RapidAPI-Host", c.cfg.Translate.APIHost)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Translation{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Translation{}, fmt.Errorf("API status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return Translation{}, fmt.Errorf("failed to read body: %w", err)
	}

	var apiResp struct {
		ResponseData struct {
			TranslatedText string `json:"translatedText"`
		} `json:"responseData"`
		Matches []struct {
			Translation string `json:"translation"`
		} `json:"matches"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return Translation{}, fmt.Errorf("failed to parse response: %w", err)
	}

	var terms []string
	if t := apiResp.ResponseData.TranslatedText; t != "" {
		terms = append(terms, t)
	}
	for _, m := range apiResp.Matches {
		if m.Translation != "" && len(terms) < 3 {
			terms = append(terms, m.Translation)
		}
	}
	if len(terms) == 0 {
		return Translation{}, fmt.Errorf("no translation found for %q", word)
	}

	return Translation{Dictionary: []DictionaryEntry{{PartOfSpeech: "word", Terms: terms}}}, nil
}

func isValidWord(word string) bool {
	if len(word) == 0 || len(word) > 50 {
		return false
	}
	for _, r := range word {
		if !unicode.IsLetter(r) && r != '-' && r != '\'' {
			return false
		}
	}
	return true
}

func (c *client) buildURL(word, targetLang string) (string, error) {
	base := c.cfg.Translate.APIURL
	if base == "" {
		base = defaultAPIURL
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("invalid API URL: %w", err)
	}
	if !allowedHosts[u.Host] {
		return "", fmt.Errorf("host not allowed: %s", u.Host)
	}
	if targetLang == "" {
		targetLang = "ru"
	}
	u.RawQuery = url.Values{"q": {word}, "langpair": {"en|" + targetLang}}.Encode()
	return u.String(), nil
}

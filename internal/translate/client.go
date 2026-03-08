package translate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
	"unicode"

	"Yulia-Lingo/internal/config"
	"Yulia-Lingo/internal/i18n"
	"Yulia-Lingo/internal/logger"
)

const (
	maxResponseSize = 1024 * 1024
	maxRetries      = 3
	maxWordLength   = 50
)

type APIClient interface {
	Translate(ctx context.Context, word, sourceLang, targetLang string) (Translation, error)
}

type client struct {
	baseURL    string
	httpClient *http.Client
	log        logger.Logger
}

func NewAPIClient(cfg *config.Config, log logger.Logger) APIClient {
	return &client{
		baseURL: cfg.Translate.APIURL,
		log:     log,
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

func (c *client) Translate(ctx context.Context, word, sourceLang, targetLang string) (Translation, error) {
	if !isValidWord(word) {
		return Translation{}, fmt.Errorf("invalid word: %s", word)
	}
	if sourceLang == "" {
		sourceLang = string(i18n.LangEN)
	}
	if targetLang == "" {
		targetLang = string(i18n.LangRU)
	}

	var lastErr error
	for attempt := range maxRetries {
		t, err := c.doTranslate(ctx, word, sourceLang, targetLang)
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

func (c *client) doTranslate(ctx context.Context, word, sourceLang, targetLang string) (Translation, error) {
	// Lingva API: GET /api/v1/{source}/{target}/{query}
	reqURL := fmt.Sprintf("%s/api/v1/%s/%s/%s",
		c.baseURL,
		sourceLang,
		targetLang,
		url.PathEscape(word),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return Translation{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Yulia-Lingo/1.0")
	req.Header.Set("Accept", "application/json")

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
		return Translation{}, fmt.Errorf("read body: %w", err)
	}

	// Lingva response shape:
	// {"translation":"...","info":{"definitions":[{"type":"noun","list":["...","..."]}]}}
	var apiResp struct {
		Translation string `json:"translation"`
		Info        struct {
			Definitions []struct {
				List []string `json:"list"`
			} `json:"definitions"`
		} `json:"info"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return Translation{}, fmt.Errorf("parse response: %w", err)
	}

	seen := map[string]bool{}
	var terms []string
	add := func(t string) {
		if t != "" && !seen[t] {
			seen[t] = true
			terms = append(terms, t)
		}
	}

	add(apiResp.Translation)
	for _, def := range apiResp.Info.Definitions {
		for _, term := range def.List {
			if len(terms) >= maxTranslations {
				break
			}
			add(term)
		}
	}

	if len(terms) == 0 {
		return Translation{}, fmt.Errorf("no translation found for %q", word)
	}
	return Translation{Terms: terms}, nil
}

func isValidWord(word string) bool {
	if len(word) == 0 || len(word) > maxWordLength {
		return false
	}
	for _, r := range word {
		if !unicode.IsLetter(r) && r != '-' && r != '\'' {
			return false
		}
	}
	return true
}

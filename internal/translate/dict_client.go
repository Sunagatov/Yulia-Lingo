package translate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"Yulia-Lingo/internal/config"
	"Yulia-Lingo/internal/logger"
)

type DictClient interface {
	PartOfSpeech(ctx context.Context, word string) string
}

type dictClient struct {
	baseURL    string
	httpClient *http.Client
	log        logger.Logger
}

func NewDictClient(cfg *config.Config, httpClient *http.Client, log logger.Logger) DictClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: cfg.Translate.Timeout}
	}
	return &dictClient{baseURL: cfg.Translate.DictAPIURL, httpClient: httpClient, log: log}
}

func (c *dictClient) PartOfSpeech(ctx context.Context, word string) string {
	reqURL := fmt.Sprintf("%s/api/v2/entries/en/%s", c.baseURL, url.PathEscape(word))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("User-Agent", "Yulia-Lingo/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ""
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return ""
	}

	var entries []struct {
		Meanings []struct {
			PartOfSpeech string `json:"partOfSpeech"`
		} `json:"meanings"`
	}
	if err := json.Unmarshal(body, &entries); err != nil || len(entries) == 0 || len(entries[0].Meanings) == 0 {
		return ""
	}
	return entries[0].Meanings[0].PartOfSpeech
}

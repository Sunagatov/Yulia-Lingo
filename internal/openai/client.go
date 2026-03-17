package openai

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	apiKey string
	apiURL string
	model  string
}

func NewClient(apiKey, apiURL, model string) *Client {
	if apiURL == "" {
		apiURL = "https://api.openai.com/v1/chat/completions"
	}
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &Client{
		apiKey: apiKey,
		apiURL: apiURL,
		model:  model,
	}
}

var defaultCategories = []string{
	"Travel & Places", "Food & Drinks", "Work & Business", "Emotions & Feelings",
	"Home & Daily Life", "Hobbies & Interests", "Health & Body", "People & Relationships",
	"Nature & Environment", "Education & Learning", "Money & Shopping", "Technology",
	"Entertainment", "Transportation", "Communication", "Other",
}

type request struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type response struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (c *Client) DetectCategory(word, translation, partOfSpeech string) (string, error) {
	prompt := fmt.Sprintf(`Word: "%s"
Translation: "%s"
Part of Speech: "%s"

Categorize this word into ONE of these categories:
%s

Return ONLY the category name, nothing else.`,
		word,
		translation,
		partOfSpeech,
		strings.Join(defaultCategories, "\n"))

	reqBody := request{
		Model: c.model,
		Messages: []message{
			{Role: "user", Content: prompt},
		},
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", c.apiURL, strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var openAIResp response
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return "", err
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("no response")
	}

	category := strings.TrimSpace(openAIResp.Choices[0].Message.Content)
	return c.validateCategory(category), nil
}

func (c *Client) validateCategory(category string) string {
	for _, valid := range defaultCategories {
		if strings.EqualFold(category, valid) {
			return valid
		}
	}
	return "Other"
}

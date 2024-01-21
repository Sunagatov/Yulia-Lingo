package api

import (
	"fmt"
	"net/http"
	"os"
)

func RequestWordsAPI(word string) (string, error) {
	wordsApiUrl := fmt.Sprintf("https://wordsapiv1.p.rapidapi.com/words/%s", word)

	newHttpRequest, err := createWordsApiHttpRequest(wordsApiUrl)
	if err != nil {
		return "", err
	}

	httpResponse, err := http.DefaultClient.Do(newHttpRequest)
	if err != nil {
		return "", err
	}

	wordResponse, err := convertToWordResponseDto(httpResponse)
	if err != nil {
		return "", err
	}

	return formatWordResponse(word, wordResponse)
}

func createWordsApiHttpRequest(wordsApiUrl string) (*http.Request, error) {
	newHttpRequest, err := http.NewRequest("GET", wordsApiUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("creation of the new http request for WordsAPI was failed: %s", err)
	}

	wordsApiKey := os.Getenv("YOUR_RAPID_API_KEY")
	if wordsApiKey == "" {
		return nil, fmt.Errorf("the key for WordsAPI is not set in environment variables")
	}

	newHttpRequest.Header.Add("X-RapidAPI-Key", wordsApiKey)
	newHttpRequest.Header.Add("X-RapidAPI-Host", "wordsapiv1.p.rapidapi.com")
	return newHttpRequest, nil
}

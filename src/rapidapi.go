package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type WordResponseDto struct {
	Word    string `json:"word"`
	Results []struct {
		Definition   string   `json:"definition"`
		PartOfSpeech string   `json:"partOfSpeech"`
		Synonyms     []string `json:"synonyms,omitempty"`
	} `json:"results"`
}

func requestWordsAPI(word string) (string, error) {
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

func convertToWordResponseDto(httpResponse *http.Response) (WordResponseDto, error) {
	var wordResponseDto WordResponseDto
	httpResponseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return wordResponseDto, err
	}
	if err = json.Unmarshal(httpResponseBody, &wordResponseDto); err != nil {
		return wordResponseDto, err
	}
	return wordResponseDto, nil
}

func formatWordResponse(word string, wordResponse WordResponseDto) (string, error) {
	var formattedWordResponse strings.Builder

	if wordResponse.Word == "" {
		formattedWordResponse.WriteString(fmt.Sprintf("I don't know this word: '%s'\n", word))
		return formattedWordResponse.String(), nil
	}

	formattedWordResponse.WriteString(fmt.Sprintf("*Word:* %s\n", wordResponse.Word))

	for i, result := range wordResponse.Results {
		if i > 0 {
			const delimiter = "\n-------------------------------\n"
			formattedWordResponse.WriteString(delimiter)
		}

		formattedWordResponse.WriteString(fmt.Sprintf("*Definition %d:* %s\n", i+1, result.Definition))
		formattedWordResponse.WriteString(fmt.Sprintf("*Part of Speech:* %s\n", result.PartOfSpeech))
	}

	return formattedWordResponse.String(), nil
}

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type WordResponseDto struct {
	Word    string `json:"rapidapi"`
	Results []struct {
		Definition   string   `json:"definition"`
		PartOfSpeech string   `json:"partOfSpeech"`
		Synonyms     []string `json:"synonyms,omitempty"`
	} `json:"results"`
}

func RequestWordsAPI(word string) (string, error) {
	wordsApiUrl := fmt.Sprintf("https://wordsapiv1.p.rapidapi.com/words/%s", word)
	wordsApiKey := os.Getenv("YOUR_WORD_API_KEY")

	newHttpRequest, err := CreateHTTPRequest("GET", wordsApiUrl, wordsApiKey, wordsApiUrl, nil)

	if err != nil {
		return "", err
	}

	httpResponse, err := http.DefaultClient.Do(newHttpRequest)
	if err != nil {
		return "", err
	}

	wordResponse, err := ConvertToWordResponseDto(httpResponse)
	if err != nil {
		return "", err
	}

	return FormatWordResponse(word, wordResponse)
}

func ConvertToWordResponseDto(httpResponse *http.Response) (WordResponseDto, error) {
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

func FormatWordResponse(word string, wordResponse WordResponseDto) (string, error) {
	var formattedWordResponse strings.Builder

	if wordResponse.Word == "" {
		formattedWordResponse.WriteString(fmt.Sprintf("<b>I don't know this rapidapi:</b> '<i>%s</i>'\n", word))
		return formattedWordResponse.String(), nil
	}

	formattedWordResponse.WriteString(fmt.Sprintf("<b>Word:</b> %s\n", wordResponse.Word))

	for i, result := range wordResponse.Results {
		if i > 0 {
			const delimiter = "\n-------------------------------\n"
			formattedWordResponse.WriteString(delimiter)
		}

		formattedWordResponse.WriteString(fmt.Sprintf("\n<b>Definition %d:</b> %s\n", i+1, result.Definition))
		formattedWordResponse.WriteString(fmt.Sprintf("<b>Part of Speech:</b> %s\n", result.PartOfSpeech))
	}

	return formattedWordResponse.String(), nil
}

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Translation struct {
	TranslatedText string            `json:"trans"`
	Dictionary     []DictionaryEntry `json:"dict"`
}

type DictionaryEntry struct {
	PartOfSpeech     string             `json:"pos"`
	Terms            []string           `json:"terms"`
	Translations     []TranslationEntry `json:"entry"`
	BaseForm         string             `json:"base_form"`
	PartOfSpeechEnum int                `json:"pos_enum"`
}

type TranslationEntry struct {
	Word                string   `json:"word"`
	ReverseTranslations []string `json:"reverse_translation"`
	Score               float64  `json:"score"`
}

func RequestTranslateAPI(wordToTranslate string) (string, error) {
	maxTranslations := 5
	translation, err := TranslateWord(wordToTranslate)
	if err != nil {
		fmt.Printf("Error translating word: %v\n", err)
		return "", err
	}
	formattedTranslation, err := FormatTranslation(maxTranslations, translation, wordToTranslate)
	if err != nil {
		fmt.Printf("Error formatting translation: %v\n", err)
		return "", err
	}
	return formattedTranslation, nil
}

func TranslateWord(word string) (Translation, error) {
	payload := strings.NewReader("from=en&to=ru&text=" + word)
	url := os.Getenv("YOUR_TRANSLATE_API_URL")
	apiKey := os.Getenv("YOUR_TRANSLATE_API_KEY")
	apiHost := os.Getenv("YOUR_TRANSLATE_API_HOST")

	newHTTPRequest, err := CreatePostHTTPRequest(url, apiKey, apiHost, payload)
	if err != nil {
		return Translation{}, err
	}

	httpResponse, err := http.DefaultClient.Do(newHTTPRequest)
	if err != nil {
		return Translation{}, err
	}

	translation, err := ConvertToTranslateResponse(httpResponse)
	if err != nil {
		return Translation{}, err
	}

	return translation, nil
}

func ConvertToTranslateResponse(httpResponse *http.Response) (Translation, error) {
	var translation Translation
	httpResponseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return translation, err
	}

	if err = json.Unmarshal(httpResponseBody, &translation); err != nil {
		return translation, err
	}
	return translation, nil
}

func FormatTranslation(maxTranslations int, translation Translation, wordToTranslate string) (string, error) {
	var formattedTranslation strings.Builder

	formattedTranslation.WriteString(fmt.Sprintf("*Полный перевод слова:* '%s'\n\n", wordToTranslate))
	formattedTranslation.WriteString(strings.Repeat("-", 5) + "\n\n")

	for _, entry := range translation.Dictionary {
		formattedTranslation.WriteString(fmt.Sprintf("*Часть речи:* '%s'\n\n", entry.PartOfSpeech))

		if len(entry.Terms) > 0 {
			if maxTranslations > len(entry.Terms) {
				maxTranslations = len(entry.Terms)
			}
			translations := make([]string, maxTranslations)
			copy(translations, entry.Terms[:maxTranslations])

			formattedTranslation.WriteString(fmt.Sprintf("*Перевод слова:*\n%s\n", strings.Join(translations, ", ")))
		}

		formattedTranslation.WriteString(strings.Repeat("-", 30) + "\n\n")
	}

	return formattedTranslation.String(), nil
}

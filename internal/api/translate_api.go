package api

import (
	"encoding/json"
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

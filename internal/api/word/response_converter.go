package word

import (
	"encoding/json"
	"io"
	"net/http"
)

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

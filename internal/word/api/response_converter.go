package api

import (
	"Yulia-Lingo/internal/word/words_dto"
	"encoding/json"
	"io"
	"net/http"
)

func convertToWordResponseDto(httpResponse *http.Response) (words_dto.WordResponseDto, error) {
	var wordResponseDto words_dto.WordResponseDto
	httpResponseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return wordResponseDto, err
	}
	if err = json.Unmarshal(httpResponseBody, &wordResponseDto); err != nil {
		return wordResponseDto, err
	}
	return wordResponseDto, nil
}

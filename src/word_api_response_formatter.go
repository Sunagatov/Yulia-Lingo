package main

import (
	"fmt"
	"strings"
)

func formatWordResponse(word string, wordResponse WordResponseDto) (string, error) {
	var formattedWordResponse strings.Builder

	if wordResponse.Word == "" {
		formattedWordResponse.WriteString(fmt.Sprintf("I don't know this word: '%s'\n", word))
		return formattedWordResponse.String(), nil
	}

	formattedWordResponse.WriteString(fmt.Sprintf("Word: %s\n", wordResponse.Word))

	for i, result := range wordResponse.Results {
		if i > 0 {
			const delimiter = "\n-------------------------------\n"
			formattedWordResponse.WriteString(delimiter)
		}

		formattedWordResponse.WriteString(fmt.Sprintf("\nDefinition %d: %s\n", i+1, result.Definition))
		formattedWordResponse.WriteString(fmt.Sprintf("Part of Speech: %s\n", result.PartOfSpeech))
	}

	return formattedWordResponse.String(), nil
}

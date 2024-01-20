package main

import (
	"fmt"
	"strings"
)

func escapeMarkdown(text string) string {
	specialCharacters := []string{"-", ".", "!", "(", ")"}
	for _, char := range specialCharacters {
		text = strings.ReplaceAll(text, char, "\\"+char)
	}
	return text
}

func formatWordResponse(word string, wordResponse WordResponseDto) (string, error) {
	var formattedWordResponse strings.Builder

	if wordResponse.Word == "" {
		formattedWordResponse.WriteString(fmt.Sprintf("I don't know this word: '%s'\n", word))
		return formattedWordResponse.String(), nil
	}

	escapedWord := escapeMarkdown(wordResponse.Word)
	formattedWordResponse.WriteString(fmt.Sprintf("*Word:* %s\n", escapedWord))

	for i, result := range wordResponse.Results {
		if i > 0 {
			const delimiter = "\n-------------------------------\n"
			formattedWordResponse.WriteString(delimiter)
		}

		escapedDefinition := escapeMarkdown(result.Definition)
		escapedPartOfSpeech := escapeMarkdown(result.PartOfSpeech)
		formattedWordResponse.WriteString(fmt.Sprintf("*Definition %d:* %s\n", i+1, escapedDefinition))
		formattedWordResponse.WriteString(fmt.Sprintf("*Part of Speech:* %s\n", escapedPartOfSpeech))
	}

	return formattedWordResponse.String(), nil
}

package my_word_list

import "fmt"

func (s *Service) formatWordList(partOfSpeech string, words []Entity, currentPage, totalPages int) string {
	messageText := fmt.Sprintf("*📚 %s* (стр. %d/%d)\n\n", s.getPartOfSpeechName(partOfSpeech), currentPage, totalPages)

	for i, word := range words {
		messageText += fmt.Sprintf("%d. *%s* - %s\n", i+1, word.Word, word.Translation)
	}

	return messageText
}

func (s *Service) getPartOfSpeechName(partOfSpeech string) string {
	names := map[string]string{
		"noun":        "Существительные",
		"verb":        "Глаголы",
		"adjective":   "Прилагательные",
		"adverb":      "Наречия",
		"preposition": "Предлоги",
		"pronoun":     "Местоимения",
	}
	if name, ok := names[partOfSpeech]; ok {
		return name
	}
	return partOfSpeech
}

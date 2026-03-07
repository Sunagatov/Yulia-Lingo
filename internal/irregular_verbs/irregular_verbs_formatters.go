package irregular_verbs

import (
	"fmt"
	"strings"
)

func (s *Service) formatVerbsList(letter string, verbs []Entity, page, totalCount int) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("*Неправильные глаголы на букву '%s'*\n\n", strings.ToUpper(letter)))

	if len(verbs) == 0 {
		builder.WriteString("Глаголы не найдены.")
		return builder.String()
	}

	for i, verb := range verbs {
		builder.WriteString(fmt.Sprintf("%d. *%s* - %s - %s\n",
			page*verbsPerPage+i+1,
			verb.Verb,
			verb.Past,
			verb.PastParticiple,
		))
		if verb.Original != "" {
			builder.WriteString(fmt.Sprintf("   _(%s)_\n", verb.Original))
		}
		builder.WriteString("\n")
	}

	totalPages := (totalCount + verbsPerPage - 1) / verbsPerPage
	builder.WriteString(fmt.Sprintf("\n📄 Страница %d из %d | Всего глаголов: %d",
		page+1, totalPages, totalCount))

	return builder.String()
}

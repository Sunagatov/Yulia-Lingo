package repository

import (
	"Yulia-Lingo/internal/db"
	"Yulia-Lingo/internal/verb/model"
	"fmt"
	"github.com/tealeg/xlsx"
	"log"
	"os"
	"strings"
)

func InsertIrregularVerbs() error {
	postgres, err := db.GetPostgresClient()
	if err != nil {
		return fmt.Errorf("can't get postgres client, err: %v", err)
	}

	query, err := prepareRequestToDB()
	if err != nil {
		return fmt.Errorf("can't get request for insert, err: %v", err)
	}

	_, err = postgres.Exec(query)
	return err
}

func prepareRequestToDB() (string, error) {
	verbs, err := readXlsxFile()

	if err != nil {
		return "", fmt.Errorf("can't get clise of verbs, err: %v", err)
	}
	var sb strings.Builder
	query := "INSERT INTO irregular_verbs (original, verb, first_letter, past, past_participle) VALUES "
	for _, verb := range verbs {
		args := fmt.Sprintf("('%s', '%s', '%s', '%s', '%s')",
			verb.Original,
			verb.Verb,
			string(verb.Verb[0]),
			verb.Past,
			verb.PastP)

		sb.WriteString(query)
		sb.WriteString(args)
		sb.WriteString(";\n")
	}

	return sb.String(), nil
}

func readXlsxFile() ([]model.IrregularVerb, error) {
	var irregularVerbs []model.IrregularVerb

	irregularVerbsFilePath := os.Getenv("IRREGULAR_VERBS_FILE_PATH")
	if irregularVerbsFilePath == "" {
		log.Fatalf("no IRREGULAR_VERBS_FILE_PATH provided in environment variables")
	}

	excelFile, err := xlsx.OpenFile(irregularVerbsFilePath)
	if err != nil {
		return nil, fmt.Errorf("can't open xlsx file, err: %v", err)
	}

	sheet := excelFile.Sheets[0]

	for i, row := range sheet.Rows {
		if i == 0 {
			continue
		}

		if len(row.Cells) == 0 {
			break
		}

		irregularVerb := model.IrregularVerb{
			Verb:     row.Cells[1].String(),
			Past:     row.Cells[2].String(),
			PastP:    row.Cells[3].String(),
			Original: row.Cells[4].String(),
		}

		irregularVerbs = append(irregularVerbs, irregularVerb)
	}

	return irregularVerbs, nil
}

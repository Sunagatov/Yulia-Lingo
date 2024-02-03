package db

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func GetIrregularVerbs(w http.ResponseWriter, r *http.Request) {
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}

	pageSize := 10
	offset := (getPageNumber(page) - 1) * pageSize

	selectQuery := fmt.Sprintf("SELECT * FROM irregular_verbs LIMIT %d OFFSET %d", pageSize, offset)
	rows, err := dbConnection.Query(selectQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	verbs := make([]IrregularVerb, 0)
	for rows.Next() {
		var verb IrregularVerb
		err := rows.Scan(&verb.ID, &verb.Verb, &verb.Past, &verb.PastP)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		verbs = append(verbs, verb)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(verbs)
}

func getPageNumber(page string) int {
	pageNumber := 1
	fmt.Sscanf(page, "%d", &pageNumber)
	if pageNumber <= 0 {
		pageNumber = 1
	}
	return pageNumber
}

var irregularVerbs = []string{
	"Идти - [Go - Went - Gone]",
	"Петь - [Sing - Sang - Sung]",
	"Есть - [Eat - Ate - Eaten]",
	"Спать - [Sleep - Slept - Slept]",
	"Говорить - [Speak - Spoke - Spoken]",
	"Брать - [Take - Took - Taken]",
	"Бежать - [Run - Ran - Run]",
	"Читать - [Read - Read - Read]",
	"Писать - [Write - Wrote - Written]",
	"Плавать - [Swim - Swam - Swum]",
	"Лететь - [Fly - Flew - Flown]",
	"Водить - [Drive - Drove - Driven]",
	"Ломать - [Break - Broke - Broken]",
	"Строить - [Build - Built - Built]",
	"Выбирать - [Choose - Chose - Chosen]",
	"Забывать - [Forget - Forgot - Forgotten]",
	"Встречать - [Meet - Met - Met]",
	"Думать - [Think - Thought - Thought]",
	"Учить - [Teach - Taught - Taught]",
	"Видеть - [See - Saw - Seen]",
	"Пить - [Drink - Drank - Drunk]",
	"Иметь - [Have - Had - Had]",
	"Делать - [Do - Did - Done]",
	"Говорить - [Say - Said - Said]",
	"Покупать - [Buy - Bought - Bought]",
	"Ломать - [Break - Broke - Broken]",
	"Начинать - [Begin - Began - Begun]",
	"Выбирать - [Choose - Chose - Chosen]",
	"Падать - [Fall - Fell - Fallen]",
	"Знать - [Know - Knew - Known]",
	"Говорить - [Speak - Spoke - Spoken]",
	"Спать - [Sleep - Slept - Slept]",
	"Находить - [Find - Found - Found]",
	"Терять - [Lose - Lost - Lost]",
	"Выигрывать - [Win - Won - Won]",
	"Рисовать - [Draw - Drew - Drawn]",
	"Держать - [Hold - Held - Held]",
	"Делать - [Make - Made - Made]",
	"Платить - [Pay - Paid - Paid]",
}

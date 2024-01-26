package handler

import (
	"Yulia-Lingo/internal/word/model"
	wordRep "Yulia-Lingo/internal/word/repository"
	"github.com/spf13/viper"
)

func (handle *MessageHandler) CallbackQuery() {
	callBackQuery := handle.Update.CallbackQuery

	if callBackQuery.Data == viper.GetString("buttons.save-word.value") {
		word := model.Word{
			// I dnk where I get that original word from
			Word: callBackQuery.Data,
			// Here will be translated
			Translate: callBackQuery.Message.Text,
		}
		wordRep.Save(&word)
	}
}

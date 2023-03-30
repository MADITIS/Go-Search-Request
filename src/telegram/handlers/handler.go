package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	config "github.com/maditis/search-go/src/config"
	"github.com/maditis/search-go/src/internal"
)

func Help(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	SendMessage(bot, update, config.Msg.Help)
}

func SendMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, message string) tgbotapi.Message {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, message) //" @"+update.Message.From.UserName
	msg.ParseMode = "HTML"
	msg.ReplyToMessageID = update.Message.MessageID
	msg.DisableWebPagePreview = true
	returnedMsg, err := bot.Send(msg)
	if err != nil {
		internal.Error.Printf("Could not send Message %s", err.Error())
		return returnedMsg
	} else {
		return returnedMsg
	}

}

func NotAuthorized(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	SendMessage(bot, update, "<strong>Not Authorized To Perform This Action.</strong>")
}

func EditMessage(chatID int64, messageID int, text string,bot *tgbotapi.BotAPI,) {
	msg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	msg.ParseMode = "HTML" 
	msg.DisableWebPagePreview = true
	_, err := bot.Send(msg)
	if err != nil {
		internal.Error.Printf("Could Not Edit Message %s", err.Error())
	}
}

func SendInfo(msg string, id int64,bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	SendMessage(bot, update, fmt.Sprintf(msg, id))
}
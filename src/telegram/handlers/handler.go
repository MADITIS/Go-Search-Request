package telegram

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	config "github.com/maditis/search-go/src/config"
	"github.com/maditis/search-go/src/database"
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

func editMessageText(bot *tgbotapi.BotAPI, update tgbotapi.Update, message string) string{
	newText := strings.Replace(message, "#request ", "", -1)
	chatId := strconv.FormatInt(update.Message.Chat.ID, 10)[4:]  // HERE is the new issue 29/1/2023
	newMessage := fmt.Sprintf("<a href='https://t.me/c/%s/%d'>Request</a> By: @%s\n\n",chatId,update.Message.MessageID, update.Message.From.UserName)
	newMessage += fmt.Sprintf("%s\n\nUID: %d | MID: %d | #requests", newText, update.Message.From.ID, update.Message.MessageID)
	return newMessage
}

func NotAuthorized(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	SendMessage(bot, update, "<strong>Not Authorized To Perform This Action.</strong>")
}

func EditMessage(chatID int64, messageID int, text string, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	msg.ParseMode = "HTML"
	msg.DisableWebPagePreview = true
	_, err := bot.Send(msg)
	if err != nil {
		internal.Error.Printf("Could Not Edit Message %s", err.Error())
	}
}

func SendInfo(msg string, id int64, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	SendMessage(bot, update, fmt.Sprintf(msg, id))
}

func KeyboardButton(bot *tgbotapi.BotAPI, update tgbotapi.Update) {

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Pickup", "pickup_data"),
			tgbotapi.NewInlineKeyboardButtonData("Reject", "reject_data"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Status", "show_status"),
		),
	)
	messageText := editMessageText(bot, update, update.Message.Text)
	msg := tgbotapi.NewMessage(config.EnvFields.RequestDest, messageText)
	msg.ParseMode = "HTML"
	// msg.ReplyToMessageID = update.Message.MessageID
	msg.ReplyMarkup = keyboard
	// msg.DisableWebPagePreview = true
	repliedMsg, err := bot.Send(msg)
	if err == nil {
		localTime := formatTime()
		chatID := strconv.FormatInt(repliedMsg.Chat.ID, 10)[4:]
		requestLink := fmt.Sprintf("<a href='https://t.me/c/%s/%d'>%s</a>", chatID, repliedMsg.MessageID, localTime)
		database.InsertOne(update.Message.From.UserName, update.Message.From.ID, repliedMsg.MessageID, localTime, requestLink)
		sendRespondRequestMSG(bot, update,repliedMsg)
	}
}

func sendRespondRequestMSG(bot *tgbotapi.BotAPI, update tgbotapi.Update,repliedMsg tgbotapi.Message) {
	messageLink := ""
	if repliedMsg.Chat.IsSuperGroup(){
		chatId := strconv.FormatInt(repliedMsg.Chat.ID, 10)[4:]
		messageLink = fmt.Sprintf("<a href='https://t.me/c/%s/%d'>Request</a>", chatId, repliedMsg.MessageID)
	} else if repliedMsg.Chat.IsPrivate() {
		messageLink = fmt.Sprintf("<a href='https://t.me/c/%s/%d'>Request</a>", repliedMsg.From.UserName,repliedMsg.MessageID)
	} 

	responseMSg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("<strong>Request submitted, You can find your request here %s</strong>", messageLink))
	responseMSg.ReplyToMessageID = update.Message.MessageID
	responseMSg.ParseMode = "HTML"
	bot.Send(responseMSg)
}


func createAlert(bot *tgbotapi.BotAPI, update tgbotapi.Update, alertMSG string) {
	params := tgbotapi.Params{}
	params.AddNonZero64("callback_query_id", internal.ConvertToInt(update.CallbackQuery.ID))
	params.AddNonEmpty("text", alertMSG)
	params.AddBool("show_alert", true)
	bot.MakeRequest("answerCallbackQuery", params)

}

func cancelPickupButtons(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	newKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Pickup", "pickup_data"),
			tgbotapi.NewInlineKeyboardButtonData("Reject", "callback_data"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Status", "show_status"),
		),
	)
	newText := reformatMessageText(update)
	newText = strings.Replace(newText, "#pending", "#requests", 1)
	msg := tgbotapi.NewEditMessageTextAndMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, newText,newKeyboard)
	msg.ParseMode = "HTML"
	bot.Send(msg)
}
func initialButtons(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	newKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Pickup", "pickup_data"),
			tgbotapi.NewInlineKeyboardButtonData("Reject", "callback_data"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Status", "show_status"),
		),
	)
	// newText := reformatMessageText(update)
	// newText = strings.Replace(newText, "#pending", "#requests", 1)
	msg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, newKeyboard)
	bot.Send(msg)
}

func createInlineButtons(bot *tgbotapi.BotAPI, update tgbotapi.Update, button1Name string, button1Data string, button2Name string, button2Data string) {
	newKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(button1Name, button1Data),
			tgbotapi.NewInlineKeyboardButtonData(button2Name, button2Data),
		),
	)
	// newText := reformatMessageText(update)
	// newText = strings.Replace(newText, "#pending", "#requests", 1)
	msg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, newKeyboard)
	// msg.ParseMode = "HTML"
	bot.Send(msg)
}

func ResultReply(bot *tgbotapi.BotAPI, update tgbotapi.Update, url string, chatid int64, messageid int) {
	newKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Click", url),
			// tgbotapi.NewInlineKeyboardButtonData(button2Name, button2Data),
		),
	)
	msg :=   tgbotapi.NewEditMessageText(chatid, messageid, "Results")
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = &newKeyboard
	bot.Send(msg)
}

func createRejectbutton(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	newKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("Rejected By %s", update.CallbackQuery.From.UserName), "declined"),
		),
	)
	newText := reformatMessageText(update)
	// newText = strings.Replace(newText, "#pending", "#fulfilled", 1)
	re := regexp.MustCompile(`#.+`)
	newText = re.ReplaceAllString(newText, "#rejected")
	msg := tgbotapi.NewEditMessageTextAndMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID,newText, newKeyboard)
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

func CallBackHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update, CallbackQuery *tgbotapi.CallbackQuery) {
	switch update.CallbackQuery.Data {
	case "pickup_data":
		createInlineButtons(bot, update, "Confirm", "confirm_pickup", "Back", "go_back_data")
	case "reject_data":
		fmt.Println("rejecting")
		if internal.CheckIfOwnerSudo(update.CallbackQuery.From.ID) {
			// fmt.Println("rejecting 2")
			if database.DeleteRequest(update.CallbackQuery.Message.MessageID) {
				createRejectbutton(bot, update)
			} else {
				createAlert(bot, update, "Can't Reject the Request. Contact the Admin.")
			}
		} else if database.DeleteByUser(update.CallbackQuery.From.ID, update.CallbackQuery.Message.MessageID) {
			createRejectbutton(bot, update)
		} else {
			createAlert(bot, update, "Not Authorized!")
		}
	case "mark_complete":
		markComplete(bot, update)

	case "show_status", "complete_data":
		if update.CallbackQuery.From.ID != config.EnvFields.OwnerID {
			break
		}
		statusMSG := ""
		var requesterID int64
		var ogmessageID int
		getRequestMessageID(update.CallbackQuery.Message.Text, &requesterID, &ogmessageID)
		result, err := database.FindOne(update.CallbackQuery.Message.MessageID, requesterID)
		if err == nil {
			messageLink := getMessageLink(update)
		    getStatusMSG(result, messageLink, &statusMSG)
			print(statusMSG)
			sendCallBackResponse(bot, update, &statusMSG)
		} else {
			createAlert(bot, update, "Something Has Gone wrong while showing status!")
		}
	case "confirm_pickup":
		if sendPickupToDB(update) {
			postPickup(bot, update)
		} else {
			createAlert(bot, update, "Could not update the request!")
		}
	case "cancel_pick":
		if database.CanCancel(update.CallbackQuery.From.ID, update.CallbackQuery.Message.MessageID) {
			createAlert(bot, update, fmt.Sprintf("This task has been cancelled by the %d who picked it up.", update.CallbackQuery.From.ID))
			cancelPickupButtons(bot, update)
		} else {
			createAlert(bot, update, "Not Authorized to cancel the pickup.")
		}
	case "go_back_data":
		initialButtons(bot, update)
	default:
		createAlert(bot, update, "Not Authorized")
	}
}

func createCompleteButtons(bot *tgbotapi.BotAPI, update tgbotapi.Update, complete_data string) {
	newKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("Completed By %s", update.CallbackQuery.From.UserName), complete_data),
		),
	)
	newText := reformatMessageText(update)
	newText = strings.Replace(newText, "#pending", "#fulfilled", 1)
	msg := tgbotapi.NewEditMessageTextAndMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID,newText, newKeyboard)
	msg.ParseMode = "HTML"
	bot.Send(msg)
	// msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, fmt.Sprintf("%s\n\n#fulfilled", update.CallbackQuery.Message.Text))
	// msg.ReplyMarkup = &newKeyboard
	// msg.ParseMode = "HTML"
	// bot.Send(msg)
	sendResposeFinal(bot, update)

}

func sendResposeFinal(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	var messageID int
	var requesterID int64
	getRequestMessageID(update.CallbackQuery.Message.Text, &requesterID, &messageID)
	chatid := getMessageLink(update)
	msg := fmt.Sprintf("<strong>You request is fulfilled by @%s, find it <a href='%s'>Here</a></strong>",update.CallbackQuery.From.UserName, chatid)
	response := tgbotapi.NewMessage(config.EnvFields.RequestSource, msg) //" @"+update.Message.From.UserName
	response.ParseMode = "HTML"
	response.ReplyToMessageID = messageID
	bot.Send(response)
}

func markComplete(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	var userValue, userKey, localTime string
	checkUsername(&userValue, &userKey, &localTime, update)
	_, err := database.AddComplete(update.CallbackQuery.From.ID, update.CallbackQuery.Message.MessageID, userKey, userValue, localTime)
	if err == nil{
		// status := getStatusMSG(result, update.CallbackQuery.Message.MessageID, strconv.FormatInt(update.CallbackQuery.Message.Chat.ID, 10)[4:])
		createCompleteButtons(bot, update, "complete_data")
	} else {
		createAlert(bot, update, "Only Who has picked up this, can fulfil it")
	}
}

func DeleteMsg(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	go time.AfterFunc(10*time.Second, func() {
		v := tgbotapi.Params{
			"chat_id":    fmt.Sprintf("%d", update.CallbackQuery.Message.Chat.ID),
			"message_id": fmt.Sprintf("%d", update.CallbackQuery.Message.MessageID+1),
		}
		bot.MakeRequest("deleteMessage", v)
	})
}

func sendCallBackResponse(bot *tgbotapi.BotAPI, update tgbotapi.Update, text *string) {
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, *text)
	msg.ParseMode = "HTML"
	msg.DisableWebPagePreview = true
	msgtoDelete, err := bot.Send(msg)
	if err == nil {
		go time.AfterFunc(20*time.Second, func() {
			v := tgbotapi.Params{
				"chat_id":    fmt.Sprintf("%d", update.CallbackQuery.Message.Chat.ID),
				"message_id": fmt.Sprintf("%d", msgtoDelete.MessageID),
			}
			bot.MakeRequest("deleteMessage", v)
		})
	}
	*text = ""
}

func formatTime() string {
	currentTime := time.Now()
	localTime := currentTime.Format("2006-01-02 15:04:05")
	return localTime
}

func getStatusMSG(result database.RequestData, messageLink string, status *string) {
	*status = fmt.Sprintf("<strong>Requested by: </strong> <a href='https://t.me/%s'>%s</a>  (<code>%d</code>) \n<strong>Request link: </strong><a href='%s'>Request Link</a>\n<strong>Request date:</strong> <code>%s</code>", result.RequestedBy.UserName, result.RequestedBy.UserName,result.RequestedBy.UserID,messageLink, result.RequestedBy.DateTime)

	if result.PickedBy != (database.RequestInfo{}) {
		if result.PickedBy.UserName == "" && result.PickedBy.FirstName != "" {
			*status += fmt.Sprintf("\n<strong>Picked up by: </strong><code>%s</code> (<code>%d</code>)\n<strong>Picked up date: </strong><code> %s</code>", result.PickedBy.FirstName, result.PickedBy.UserID, result.PickedBy.DateTime)
		} else {
			*status += fmt.Sprintf("\n<strong>Picked up by: </strong><a href='https://t.me/%s'>%s</a> (<code>%d</code>)\n<strong>Picked up date: </strong><code> %s</code>", result.PickedBy.UserName, result.PickedBy.UserName,result.PickedBy.UserID, result.PickedBy.DateTime)
		}
	}
	if result.CompletedBy != (database.RequestInfo{}) {
		if result.PickedBy.UserName == "" && result.PickedBy.FirstName != "" {
			*status += fmt.Sprintf("\n<strong>Completed up by: </strong><code>%s</code> (<code>%d</code>)\n<strong>Completed up date: </strong><code> %s</code>", result.CompletedBy.FirstName, result.CompletedBy.UserID, result.CompletedBy.DateTime)
		} else {
			*status += fmt.Sprintf("\n<strong>Completed up by: </strong><a href='https://t.me/%s'>%s</a> (<code>%d</code>)\n<strong>Completed up date: </strong><code> %s</code>", result.CompletedBy.UserName,result.CompletedBy.UserName, result.CompletedBy.UserID, result.CompletedBy.DateTime)
		}
	}
}

func checkUsername(userValue, userKey, localTime *string, update tgbotapi.Update ) {
	*userValue = update.CallbackQuery.From.FirstName
	*userKey = "FirstName"
	if update.CallbackQuery.From.UserName != "" {
		*userKey = "UserName"
		*userValue = update.CallbackQuery.From.UserName
	}
	*localTime = formatTime()
}

func sendPickupToDB(update tgbotapi.Update) bool {
	var userValue, userKey, localTime string
	checkUsername(&userValue,&userKey, &localTime, update)
	var (
		requesterID int64
	 	messageID int
	)
	getRequestMessageID(update.CallbackQuery.Message.Text, &requesterID, &messageID)
	result := database.UpdateOne(userKey, update.CallbackQuery.From.ID, update.CallbackQuery.Message.MessageID, localTime, userValue,
		requesterID)

	return result
}

func getMessageLink(update tgbotapi.Update) string {
	var messageLink string
	if update.CallbackQuery.Message.Chat.IsSuperGroup(){
		chatId := strconv.FormatInt(update.CallbackQuery.Message.Chat.ID, 10)[4:]
		messageLink = fmt.Sprintf("https://t.me/c/%s/%d", chatId, update.CallbackQuery.Message.MessageID)
	} else if update.CallbackQuery.Message.Chat.IsPrivate() {
		messageLink = fmt.Sprintf("https://t.me/c/%s/%d", update.CallbackQuery.Message.Chat.UserName,update.CallbackQuery.Message.MessageID)
	} 
	return messageLink
}

// buttons
func postPickup(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	newKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Fulfil", "mark_complete"),
			tgbotapi.NewInlineKeyboardButtonData("Reject", "reject_data"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Cancel", "cancel_pick"),
			tgbotapi.NewInlineKeyboardButtonData("Status", "show_status"),
		),
	)
	// messageText := update.CallbackQuery.Message.Text
	// for _, entity := range update.CallbackQuery.Message.Entities {
	// 	if entity.Type = "url"
	// }
	// url :=
	// newText := strings.Replace(strings.Replace(messageText,"#requests", "#pending", 1), "Requests", )
		
	newText := reformatMessageText(update)
	newText = strings.Replace(newText, "#requests", "#pending", 1)

	msg := tgbotapi.NewEditMessageTextAndMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, newText, newKeyboard)
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

func reformatMessageText(update tgbotapi.Update) string {
	messageText := update.CallbackQuery.Message.Text
	// var length int
	for _, entity :=  range update.CallbackQuery.Message.Entities {
		switch entity.Type {
		case "text_link":
			if entity.Offset == 0 {
				substring := messageText[entity.Offset: entity.Offset+entity.Length]
				url := entity.URL
				link := fmt.Sprintf("<a href='%s'>Request</a>", url)
				// length = len(link) - len(substring)
				messageText = strings.Replace(messageText, substring, link, 1)
			}
			// fmt.Println("entitiest/n/n",update.CallbackQuery.Message.Entities)
			// return messageText
		// case "bold":
		// 	substring := messageText[entity.Offset+ length - 1 :entity.Offset+length +entity.Length ]
		// 	msg := fmt.Sprintf("<strong>%s</strong>", substring)
		// 	length += len(msg) - len(substring)
		// 	messageText = strings.Replace(messageText, substring, msg, 1)
		// case "code":
		// 	substring := messageText[entity.Offset+ length - 1:entity.Offset+length +entity.Length]
		// 	msg := fmt.Sprintf("<code>%s</code>", substring)
		// 	length += len(msg) - len(substring)
		// 	messageText = strings.Replace(messageText, substring, msg, 1)
			
		}
	}
	return messageText
	
}

func getRequestMessageID(messageText string, requesterID *int64, messageID *int) {
	re := regexp.MustCompile(`(?i)(?:UID:\s*)([0-9]+)(?:\s*\|\sMID:\s)([0-9]+)(?:\s*\|)`) 
    submatches := re.FindStringSubmatch(messageText)
	println("Submatched Message: ", submatches[1], submatches[2])
    userID, err := strconv.ParseInt(submatches[1], 10, 64)
    if err != nil {
        fmt.Println("Error converting UserID to int64: ", err)
    } else {
        fmt.Println("UserID: ", userID)
		*requesterID = userID
    }
	mID, err := strconv.Atoi(submatches[2])
	if err != nil {
		fmt.Println("Error Converting og message iD")
		} else {
		fmt.Println("UserID: ", mID)
		*messageID = mID 
	}

}
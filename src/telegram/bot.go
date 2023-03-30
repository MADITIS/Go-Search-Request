package telegram

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/maditis/search-go/src/config"

	"github.com/maditis/search-go/src/drive"
	handler "github.com/maditis/search-go/src/telegram/handlers"

	// config "github.com/maditis/search-go/src/config"
	internal "github.com/maditis/search-go/src/internal"
)

func Start() {
	bot, err := tgbotapi.NewBotAPI("5915960906:AAGoG9QSsznsjA9AYjARFBjzmOCgF7fcL4I")
	internal.ErrorExit(err, "Provide Valid Bot Token")

	bot.Debug = true
	internal.Info.Printf("Authorized on account %v-Bot Started:", bot.Self.UserName)
	handleRequests(bot)
}


func handleRequests(bot *tgbotapi.BotAPI) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	
	updates := bot.GetUpdatesChan(u)
	
	for update := range updates {
		if update.Message == nil || !update.Message.IsCommand(){
			internal.Info.Println("Not a command:")
            continue
        }
		userID := update.Message.From.ID
		startPing := time.Now()
		if (internal.Authorized(update.Message.Chat.ID) || internal.Authorized(userID) || internal.CheckIfOwnerSudo(userID)) {
			handleUpdates(update, bot, startPing)
		} else if internal.CheckFriends(userID) {
			// handler.SendInfo(config.Msg.SudoInfo, userID, bot, update)
			handleUpdates(update, bot, startPing)
		} else {
			handler.SendInfo(config.Msg.NoAccess, userID, bot, update)
		}
		if update.InlineQuery != nil {
			fmt.Print(update.InlineQuery)
		}
	}
	
}

func handleUpdates(update tgbotapi.Update, bot *tgbotapi.BotAPI, startPing time.Time) {
	chatId := update.Message.Chat.ID
	userId := update.Message.From.ID
	chatTitle := update.Message.Chat.Title
	currentUserName := update.Message.From.UserName
	switch update.Message.Command() {
	case "start":
		handler.SendMessage(bot, update, "<strong>Not implemented</strong>")

	case "search":
		go searchHandler(update, bot)

	case "help":
		handler.Help(bot, update)

	case "id":
		if update.Message.ReplyToMessage != nil {
			handler.SendMessage(bot, update, fmt.Sprintf("<strong>UserID</strong>: <code>%d</code>\n<strong>ChatID</strong>: <code>%d</code>", update.Message.ReplyToMessage.From.ID, update.Message.ReplyToMessage.Chat.ID))
		}

	case "sudo":
		fmt.Println(config.EnvFields.Friends, "meesaagrs", update.Message.Text)
		if internal.CheckIfOwnerSudo(userId) {
			authorizeHandler(update, bot, currentUserName, chatTitle, chatId, true)
		} else if (internal.CheckFriends(userId)){
			fmt.Println("Friend", userId)
			if update.Message.ReplyToMessage !=  nil {
				if update.Message.ReplyToMessage.From.ID == userId {
					addSudo(userId, update, bot)
				}
			}else if userIDToAuthorize := strings.Split(update.Message.Text, "/sudo "); len(userIDToAuthorize) == 2{
				userid , err := strconv.ParseInt(strings.TrimSpace(userIDToAuthorize[1]), 10, 64)
				if (err == nil && internal.CheckFriends(userid)) {
					addSudo(userid, update, bot)
				} else {
					handler.SendMessage(bot, update, "<strong> Wrong ID: Not eligible to become a sudo</strong>")
				}
			
			}else {
				handler.SendMessage(bot, update, "<strong>First Make Yourself a Sudo</strong><code>/sudo</code>")
			}
		}else {
			 handler.SendInfo(config.Msg.SudoInfo, userId, bot, update)
		}
	
	case "sudorm":
		if internal.CheckIfOwnerSudo(userId) {
			unauthorizeHandler(update, bot, currentUserName, chatTitle, chatId, true)
		} else {
			handler.SendInfo(config.Msg.NoAccess, userId, bot, update)
		}

	case "unauthorize":
		if internal.CheckIfOwnerSudo(userId) {
			go unauthorizeHandler(update, bot, currentUserName, chatTitle, chatId, false)
		} else if internal.CheckFriends(userId){
			go handler.SendInfo(config.Msg.SudoInfo,userId, bot, update)
		} else {
			go handler.SendInfo(config.Msg.UserNoAccess,userId, bot, update)
		}

	// authorize user/chat using the id or by reply to the message
	case "authorize":
		if internal.CheckIfOwnerSudo(userId) {
			go authorizeHandler(update, bot, currentUserName, chatTitle, chatId, false)
		} else if internal.CheckFriends(userId){
			go handler.SendInfo(config.Msg.SudoInfo,userId, bot, update)
		} else {
			go handler.SendInfo(config.Msg.UserNoAccess,userId, bot, update)
		}

	case "ping":
		timeSince := time.Since(config.Start)
		pingSince := time.Since(startPing)
		msg := "<strong>Ping Time</strong>: <code>%d.%dms</code>\n<strong>Service Is Alive Since</strong>: <code>%s</code>"
		go handler.SendMessage(bot, update, fmt.Sprintf(msg,pingSince.Milliseconds(), pingSince.Microseconds() ,timeSince.Round(time.Second).String()))

	case "stats":
		if internal.CheckIfOwnerSudo(userId) {
			go stats(update, bot, currentUserName, chatTitle, chatId, userId)
		}else if internal.CheckFriends(userId){
			go handler.SendInfo(config.Msg.SudoInfo,userId, bot, update)
		}else {
			go handler.SendInfo(config.Msg.UserNoAccess,userId, bot, update)
		}
	}
}

func stats(update tgbotapi.Update, bot *tgbotapi.BotAPI, username string, chatTitle string,chatID int64, userid int64) {
	var stats = fmt.Sprintf("<strong>Statistics:</strong>\n\n<strong>User ID: </strong><code>%d</code><strong>  (%s)</strong>\n                <u>Permission: </u>", userid, username)

	if internal.CheckIfOwnerSudo(userid) {
		stats += "<code>Sudo</code>\n\n"
	}else {
		stats += "<code>User</code>\n"
		if internal.Authorized(userid) {
			stats += "<u>                Authorized: </u><code>True</code>\n\n"
		} else {
			stats += "<u>                Authorized: </u><code>False</code>\n\n"
		}
	}


	if update.Message.ReplyToMessage != nil {
		repliedID := update.Message.ReplyToMessage.From.ID
		repliedUserName := update.Message.ReplyToMessage.From.UserName
		stats += fmt.Sprintf("<strong>Replied To: </strong><code>%d</code><strong>  (%s)</strong>\n                <u>Permission: </u>", repliedID, repliedUserName)
		if internal.CheckIfOwnerSudo(repliedID){
			stats += "<code>Sudo</code>\n\n"
		} else {
			stats += "<code>User</code>\n"
			if internal.Authorized(repliedID) {
				stats += "<u>                 Authorized :</u><code>True</code>\n\n"
			} else {
				stats += "<u>                 Authorized :</u><code>False</code>\n\n"
			}
		}	
	} 
		
	chatType := update.Message.Chat.Type
	stats += fmt.Sprintf("<strong>Chat: </strong><code>%s</code>\n                <u>Type: </u><code>%s</code>\n                <u>ID: </u><code>%d</code>\n", chatTitle, chatType, chatID)

	if internal.Authorized(chatID) {
		stats += "                <u>Authorized: </u><code>True</code>"
	} else {
		stats += "                <u>Authorized: </u><code>False</code>"
	}
	handler.SendMessage(bot, update, stats)
}


func prepareSudo() {
	
}

func sudoRemove(id int64, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if internal.CheckIfOwnerSudo(id) && !(config.EnvFields.OwnerID == id){
		delete(config.EnvFields.SudoOwner, id)
		handler.SendMessage(bot, update, fmt.Sprintf("%d have been removed from the sudo list on Phoenix. And longer have access to all the features. Contact the admin if you have any questions.", id))
	} else {
		handler.SendMessage(bot, update, fmt.Sprintf("%d<strong> Not a Sudo User Or It's the ID of Owner.</strong>", id))
	}
}

func addSudo(id int64, update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	if internal.CheckIfOwnerSudo(id) {
		handler.SendInfo(config.Msg.SudoAccess, id, bot, update)
	} else {
		if internal.Authorized(id) {
			delete(config.EnvFields.AuthorizedUsers, id)
			config.EnvFields.SudoOwner[id] = len(config.EnvFields.SudoOwner) + 1
		} else {
			config.EnvFields.SudoOwner[id] = len(config.EnvFields.SudoOwner) + 1
		}
		handler.SendInfo(config.Msg.GotSudoAccess, id, bot, update)
	}
}

func authorizeHandler(update tgbotapi.Update, bot *tgbotapi.BotAPI, currentUserName string, chatTitle string, chatId int64, isSudo bool) {
	if update.Message.ReplyToMessage != nil && !update.Message.ReplyToMessage.From.IsBot{

		repliedUserID := update.Message.ReplyToMessage.From.ID
		repliedUserName := update.Message.ReplyToMessage.From.UserName
		if isSudo {
			addSudo(repliedUserID, update, bot)
			return
		}
		switch internal.Authorized(repliedUserID) {
		case false:
			config.EnvFields.AuthorizedUsers[repliedUserID] = len(config.EnvFields.AuthorizedUsers) + 1
			handler.SendMessage(bot, update, fmt.Sprintf("<strong><code>@%s</code> Is Authorized By @%s!</strong>", repliedUserName, currentUserName))
		case true:
			handler.SendMessage(bot, update, fmt.Sprintf("<code>@%s</code> <strong>Is Already Authorized!</strong>", repliedUserName))
		}


	} else if userIDToAuthorize := strings.Split(update.Message.Text, "/sudo "); len(userIDToAuthorize) == 2{
		userid , err := strconv.ParseInt(strings.TrimSpace(userIDToAuthorize[1]), 10, 64)
		if (err == nil && isSudo) {
			addSudo(userid, update, bot)
			return
		}
		if err == nil && !internal.Authorized(userid){
			config.EnvFields.AuthorizedUsers[userid] = len(config.EnvFields.AuthorizedUsers) + 1
			handler.SendMessage(bot, update, fmt.Sprintf("<strong><code>%d</code> Is Authorized! By @%s</strong>", userid, currentUserName))
		} else {
			handler.SendMessage(bot, update, fmt.Sprintf("<code>%d</code> <strong>Invalid User ID! Or Chat Is Already Authorized.</strong>", userid))
		}		

	} else if internal.Authorized(chatId){
		if isSudo {
			handler.SendMessage(bot, update, config.Msg.Sudo)
			return
		}
		handler.SendMessage(bot, update, fmt.Sprintf("<code>%s</code> <strong>Is Already Authorized!</strong>", chatTitle))
	} else {
		if isSudo {
			handler.SendMessage(bot, update, config.Msg.Sudo)
			return
		}
		config.EnvFields.AuthorizedUsers[chatId] = len(config.EnvFields.AuthorizedUsers) + 1
		handler.SendMessage(bot, update, fmt.Sprintf("<strong>@%s Has Authorized This Chat (<code>%s</code>)</strong>", currentUserName, chatTitle))
	}
}

func unauthorizeHandler(update tgbotapi.Update, bot *tgbotapi.BotAPI, currentUserName string, chatTitle string, chatId int64, isSudo bool) {
	if update.Message.ReplyToMessage != nil && !update.Message.ReplyToMessage.From.IsBot{
		repliedUserID := update.Message.ReplyToMessage.From.ID
		repliedUserName := update.Message.ReplyToMessage.From.UserName
		if isSudo {
			sudoRemove(repliedUserID, bot, update)
			return
		}
		switch internal.Authorized(repliedUserID) {
		case true:
			delete(config.EnvFields.AuthorizedUsers, repliedUserID)
			handler.SendMessage(bot, update, fmt.Sprintf("<strong><code>@%s</code> Is Unauthorized By @%s!</strong>", repliedUserName, currentUserName))
		case false:
			handler.SendMessage(bot, update, fmt.Sprintf("<code>@%s</code><strong> Is Not Authorized!</strong>", repliedUserName))
		}

	} else if userIDToAuthorize := strings.Split(update.Message.Text, "/sudorm "); len(userIDToAuthorize) == 2{
		userid , err := strconv.ParseInt(strings.TrimSpace(userIDToAuthorize[1]), 10, 64)
		if isSudo && (err == nil){
			sudoRemove(userid, bot, update)
			return
		}
		if internal.CheckIfOwnerSudo(userid) && (err == nil) {
			handler.SendMessage(bot, update, fmt.Sprintf("<code>%d</code><strong>This is the ID of a Sudo User! Can't Unauthorize it.</strong>", userid))
			return
		}

		if err == nil && internal.Authorized(userid) {
			delete(config.EnvFields.AuthorizedUsers, userid)
			handler.SendMessage(bot, update, fmt.Sprintf("<strong><code>%d</code> Is Unauthorized! By @%s</strong>", userid, currentUserName))
		} else {
			handler.SendMessage(bot, update, fmt.Sprintf("<code>%d</code> <strong>Invalid User ID! Or Chat Is Not Authorized</strong>", userid))
		}
		

	} else if internal.Authorized(chatId){
		if isSudo {
			handler.SendMessage(bot, update, config.Msg.Sudorm)
			return
		}
		delete(config.EnvFields.AuthorizedUsers, chatId)
		handler.SendMessage(bot, update, fmt.Sprintf("<strong><code>%d</code> Is Unauthorized! By @%s</strong>", chatId, currentUserName))
	} else {
		if isSudo {
			handler.SendMessage(bot, update, config.Msg.Sudorm)
			return
		}
		handler.SendMessage(bot, update, fmt.Sprintf("<code>%s</code><strong> Is Not Authorized!</strong>", chatTitle))
	} 
}


func searchHandler(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	var queryText string = "" 
	botName := fmt.Sprintf("/search@%s", bot.Self.UserName)
	messageText := update.Message.Text
	if strings.HasPrefix(messageText, botName)  && strings.HasSuffix(messageText, botName) {
		handler.SendMessage(bot, update, config.Msg.Search)
		return
	}

	if strings.HasPrefix(messageText, "/search") && strings.HasSuffix(messageText, "/search") {
		handler.SendMessage(bot, update, config.Msg.Search)
		return 
	}

	var queryLen []string
	if strings.HasPrefix(messageText, botName+" ") {
		queryLen = strings.SplitN(messageText, botName, 2)
	}else if strings.HasPrefix(messageText, "/search ") {
		queryLen = strings.SplitN(messageText, "/search", 2)
	} else {
		handler.SendMessage(bot, update, fmt.Sprintf("<strong>Invalid Search Query</strong><code> %s</code>\n\n%s", queryText, config.Msg.Search))
		return
	}
	if len(queryLen) == 2 {
		queryText = strings.TrimSpace(queryLen[1])
		internal.Info.Printf("Searching %s", queryText)
		
		go func() {
			returnedMSG := handler.SendMessage(bot, update, fmt.Sprintf("<strong>Searching....%s</strong>", queryText))
			startTime := time.Now()
			result, totalResults, err := drive.Search(queryText)
			if err == nil {
				timeTaken := time.Since(startTime)
				handler.EditMessage(returnedMSG.Chat.ID, returnedMSG.MessageID, fmt.Sprintf(config.Msg.ResultText,queryText, totalResults, result, timeTaken.Milliseconds(), returnedMSG.ReplyToMessage.From.ID, returnedMSG.ReplyToMessage.From.UserName), bot)
			} else {
				handler.EditMessage(returnedMSG.Chat.ID, returnedMSG.MessageID, fmt.Sprintf("No Results Found for %s", queryText), bot)
			}
		}()
	}

}
// func handleInlineQuery(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
// 	inlineQuery := update.InlineQuery.Query
	
// 	if strings.HasPrefix(inlineQuery, "@search ") {
// 		// Query:= strings.TrimSpace(strings.TrimPrefix(inlineQuery, "@search "))
// 		results := tmdb.Tmdb()
// 		var tgResults []interface{}
// 		for _, value := range results.Results {
// 			tgResults = append(tgResults, value)
// 		}
// 		inlineConf := tgbotapi.InlineConfig{
// 			InlineQueryID: update.InlineQuery.ID,
// 			IsPersonal:    true,
// 			Results:        tgResults,
// 			CacheTime:     0,
// 		}
// 		tgbotapi.inline

// 	}
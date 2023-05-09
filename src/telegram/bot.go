package telegram

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/maditis/search-go/src/config"
	"github.com/maditis/search-go/src/database"

	"github.com/maditis/search-go/src/drive"
	handler "github.com/maditis/search-go/src/telegram/handlers"

	// config "github.com/maditis/search-go/src/config"
	internal "github.com/maditis/search-go/src/internal"
)



func Start() {
	bot, err := tgbotapi.NewBotAPI(config.EnvFields.BotToken)
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
		if update.CallbackQuery != nil {
			if userId := update.CallbackQuery.From.ID; update.CallbackQuery.Message.Chat.ID == config.EnvFields.RequestDest || 
							internal.CheckIfOwnerSudo(userId) || internal.Authorized(update.CallbackQuery.Message.Chat.ID) {
				go handler.CallBackHandler(bot, update, update.CallbackQuery)
			}
		}


		if update.Message == nil {
			internal.Info.Println("Not a Message:")
			continue
		}

		userID := update.Message.From.ID
		startPing := time.Now()
		chatId := update.Message.Chat.ID


		if strings.HasPrefix(update.Message.Text, "#request ") {
			if chatId == config.EnvFields.RequestSource || userID == config.EnvFields.OwnerID{
				if update.Message.From.UserName != "" {
					go createRequest(update, bot)
				} else {
					handler.SendMessage(bot, update, "<code>It looks like you haven't set up a username. Please create a unique username to make requests.</code>")
				}
			} else {
				handler.SendMessage(bot, update, "<strong>Please go to <code>Alchemist • 02 • Vhagar • Dragonpit</code> to make a request.</strong>")
			}
		} else if strings.HasPrefix(update.Message.Text, "#request") && strings.HasSuffix(update.Message.Text, "#request") {
			handler.SendMessage(bot, update, config.Msg.RequestText)
		}

		if internal.CheckIfOwnerSudo(userID) || internal.Authorized(userID) || internal.Authorized(update.Message.Chat.ID) || internal.CheckFriends(userID) {
			handleUpdates(update, bot, startPing)
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
		if userId == config.EnvFields.OwnerID || (chatId == config.EnvFields.SearchChat && internal.Authorized(chatId)){
			go searchHandler(update, bot)
		}

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
		} else if internal.CheckFriends(userId) {
			fmt.Println("Friend", userId)
			prepareSudo(update, bot, userId)
		} else {
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
		} else if internal.CheckFriends(userId) {
			go handler.SendInfo(config.Msg.SudoInfo, userId, bot, update)
		} else {
			go handler.SendInfo(config.Msg.UserNoAccess, userId, bot, update)
		}

	// authorize user/chat using the id or by reply to the message
	case "authorize":
		if internal.CheckIfOwnerSudo(userId) {
			go authorizeHandler(update, bot, currentUserName, chatTitle, chatId, false)
		} else if internal.CheckFriends(userId) {
			go handler.SendInfo(config.Msg.SudoInfo, userId, bot, update)
		} else {
			go handler.SendInfo(config.Msg.UserNoAccess, userId, bot, update)
		}

	case "ping":
		timeSince := time.Since(config.Start)
		pingSince := time.Since(startPing)
		msg := "<strong>Ping Time</strong>: <code>%d.%dms</code>\n<strong>Service Is Alive Since</strong>: <code>%s</code>"
		go handler.SendMessage(bot, update, fmt.Sprintf(msg, pingSince.Milliseconds(), pingSince.Microseconds(), timeSince.Round(time.Second).String()))

	case "stats":
		if internal.CheckIfOwnerSudo(userId) {
			go stats(update, bot, currentUserName, chatTitle, chatId, userId)
		} else if internal.CheckFriends(userId) {
			handler.SendInfo(config.Msg.SudoInfo, userId, bot, update)
		} else {
			handler.SendInfo(config.Msg.UserNoAccess, userId, bot, update)
		}

	case "mode":
		if internal.CheckIfOwnerSudo(userId) {
			if mode := strings.Split(update.Message.Text, "/mode "); len(mode) == 2 {
				tempText := strings.TrimSpace(mode[1])
				if strings.ToLower(tempText) == "lazy" {
					drive.SearchType = false
					handler.SendMessage(bot, update, "<strong>Changed The mode to</strong><code> lazy</code>")
					drive.ModeText = tempText
				} else if strings.ToLower(tempText) == "full" {
					drive.SearchType = true
					handler.SendMessage(bot, update, "<strong>Changed The mode to</strong><code> full</code>")
					drive.ModeText = tempText
				} else {
					handler.SendMessage(bot, update, fmt.Sprintf("<strong>Invalid Mode.</strong>\n\n<strong>Usage:\n</strong>/mode [<code>lazy/full</code>]\n\n<strong>Current Search Mode: </strong><code>%s</code>", drive.ModeText))
				}
			} else {
				handler.SendMessage(bot, update, fmt.Sprintf("<strong>Invalid Mode.</strong>\n\n<strong>Usage:\n</strong>/mode [<code>lazy/full</code>]\n\n<strong>Current Search Mode: </strong><code>%s</code>", drive.ModeText))
			}
		} else if internal.CheckFriends(userId) {
			go handler.SendInfo(config.Msg.SudoInfo, userId, bot, update)
		} else {
			go handler.SendInfo(config.Msg.UserNoAccess, userId, bot, update)
		}
	case "wipe":
		if userId == config.EnvFields.OwnerID {
			if database.ClearRequests() {
				handler.SendMessage(bot, update, "<strong>Purged Every Requests</strong>")
			} else {
				handler.SendMessage(bot, update, "Could Not purge")
			}
		} else {
			handler.SendMessage(bot, update, "<strong>Owner Only Operation</strong>")
		}

	case "myrequests":
		showRequests(update, bot)
	}
}

// show requests
func showRequests(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	if update.Message.ReplyToMessage != nil && !update.Message.ReplyToMessage.From.IsBot {
		fetchRequests(update, bot, update.Message.ReplyToMessage.From.ID, update.Message.ReplyToMessage.From.UserName)

	}else if userIDToAuthorize := strings.Split(update.Message.Text, "/myrequests "); len(userIDToAuthorize) == 2 {
		userid, err := strconv.ParseInt(strings.TrimSpace(userIDToAuthorize[1]), 10, 64)
		if err == nil {
			fetchRequests(update, bot, userid, "")
		}
	}else {
		fetchRequests(update, bot, update.Message.From.ID, update.Message.From.UserName)
	}
}


func fetchRequests(update tgbotapi.Update, bot *tgbotapi.BotAPI, userId int64, userName string) {
	myRequests, err1 := database.GetPendingRequests(userId, true, false)
	pendingRequests, err2 := database.GetPendingRequests(userId, false, false)
	completedRequests, err3 := database.GetPendingRequests(userId, false, true)
	var userStatus string = fmt.Sprintf("<strong><code>%s</code> Request Stats</strong>\n\n", userName)
	// var userName string
	createMSG(update, bot, myRequests, err1, "Total", userName, &userStatus)
	createMSG(update, bot, pendingRequests, err2, "Pending", userName, &userStatus)
	createMSG(update, bot, completedRequests, err3, "fulfilled", userName, &userStatus)
	handler.SendMessage(bot, update, userStatus)
	// func createMSG(requests []database.RequestData, err, requestType string) {
	// 	if err != nil {
	// 		handler.SendMessage(bot, update, "Probably Some Horrible Things has occured while getting the request. Maybe world is about to end")
	
	// 	} else {
	// 		totalRequests := len(requests)
	// 		if totalRequests >= 1 {
	// 			userStatus += fmt.Sprintf("<code> %d </code><strong>%s Requests</strong>\n" ,requestType)
	// 			// msg := fmt.Sprintf(fmt.Sprintf("<strong>User %s Has</strong><code> %d </code><strong>Pending Requests</strong>\n", results[0].RequestedBy.UserName, totalRequests))
	// 			for i, request := range results {
	// 				// print(request)
	// 				userStatus += fmt.Sprintf("\n<strong>%d: %s</strong>", i+1, request.RequestedBy.MessageLink)
	// 			}
	// 		} else {
	// 			handler.SendMessage(bot, update, fmt.Sprintf("<strong>User %s Has</strong> No <strong>%s Requests</strong>\n", userName, requestType))
	// 		}
	// 	}
	// }


	// if err1 != nil {
	// 	handler.SendMessage(bot, update, "Probably Some Horrible Things has occured while getting the request. Maybe world is about to end")

	// } else {
	// 	totalRequests := len(myRequests)
	// 	if totalRequests == 1 {
	// 		handler.SendMessage(bot, update, fmt.Sprintf("<strong>User %s Has</strong><code> 1 </code><strong>Pending Requests</strong>\n<strong>1: %s</strong>", results[0].RequestedBy.UserName, results[0].RequestedBy.MessageLink))
	// 	} else if totalRequests > 1 {
	// 		msg := fmt.Sprintf(fmt.Sprintf("<strong>User %s Has</strong><code> %d </code><strong>Pending Requests</strong>\n", results[0].RequestedBy.UserName, totalRequests))
	// 		for i, request := range results {
	// 			// print(request)
	// 			msg += fmt.Sprintf("\n<strong>%d: %s</strong>", i+1, request.RequestedBy.MessageLink)
	// 		}
	// 		handler.SendMessage(bot, update, msg)
	// 	} else {
	// 		handler.SendMessage(bot, update, fmt.Sprintf("<strong>User %s Has</strong> No <strong>Pending Requests</strong>\n", userName))
	// 	}
	// }
}
func createMSG(update tgbotapi.Update, bot *tgbotapi.BotAPI, 
			requests []database.RequestData, err error, requestType string, userName string, userStatus *string) {
	if err != nil {
		handler.SendMessage(bot, update, "Probably Some Horrible Things has occured while getting the request. Maybe world is about to end")

	} else {
		totalRequests := len(requests)
		if totalRequests >= 1 {
			

			*userStatus += fmt.Sprintf("<code> (%d) </code><i>%s Requests</i>\n" ,totalRequests,requestType)
			// msg := fmt.Sprintf(fmt.Sprintf("<strong>User %s Has</strong><code> %d </code><strong>Pending Requests</strong>\n", results[0].RequestedBy.UserName, totalRequests))
			if requestType != "Total" {
				for i, request := range requests {
					// print(request)
					*userStatus += fmt.Sprintf("\n<strong>%d: %s</strong>", i+1, request.RequestedBy.MessageLink)
				}
			}
			*userStatus += "\n--------------------------------------------------\n"
			// handler.SendMessage(bot, update, userStatus)

		} else {
			*userStatus += fmt.Sprintf("\n<strong>Zero %s Requests</strong>\n-------------------------------------------------\n", requestType)
		}
	}
}

// func showUserRequests(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
// 	repliedId := update.Message.ReplyToMessage.From.ID
// 	results, err := database.GetPendingRequests(repliedId)
// 	if err != nil {
// 		handler.SendMessage(bot, update, "Probably Some Horrible Things has occured while getting the request. Maybe world is about to end")

// 	} else {
// 		totalRequests := len(results)
// 		if totalRequests == 1 {
// 			handler.SendMessage(bot, update, fmt.Sprintf("<strong>User %s Has</strong><code> 1 </code><strong>Pending Requests</strong>\n<strong>1: %s</strong>", results[0].RequestedBy.UserName, results[0].RequestedBy.MessageLink))
// 		} else if totalRequests > 1 {
// 			msg := fmt.Sprintf(fmt.Sprintf("<strong>User %s Has</strong><code> %d </code><strong>Pending Requests</strong>\n", results[0].RequestedBy.UserName, totalRequests))
// 			for i, request := range results {
// 				// print(request)
// 				msg += fmt.Sprintf("\n<strong>%d: %s</strong>", i+1, request.RequestedBy.MessageLink)
// 			}
// 			handler.SendMessage(bot, update, msg)
// 		} else {
// 			handler.SendMessage(bot, update, fmt.Sprintf("<strong>User %s Has</strong> No <strong>Pending Requests</strong>\n", update.Message.ReplyToMessage.From.UserName))
// 		}
// 	}

// }

// show stats
func stats(update tgbotapi.Update, bot *tgbotapi.BotAPI, username string, chatTitle string, chatID int64, userid int64) {
	var stats = fmt.Sprintf("<strong>Statistics:</strong>\n\n<strong>User ID: </strong><code>%d</code><strong>  (%s)</strong>\n<u>Permission: </u>", userid, username)

	if internal.CheckIfOwnerSudo(userid) {
		stats += "<code>Sudo</code>\n\n"
	} else {
		stats += "<code>User</code>\n"
		if internal.Authorized(userid) {
			stats += "<u>Authorized: </u><code>True</code>\n\n"
		} else {
			stats += "<u>Authorized: </u><code>False</code>\n\n"
		}
	}

	if update.Message.ReplyToMessage != nil {
		repliedID := update.Message.ReplyToMessage.From.ID
		repliedUserName := update.Message.ReplyToMessage.From.UserName
		stats += fmt.Sprintf("<strong>Replied To: </strong><code>%d</code><strong>  (%s)</strong>\n<u>Permission: </u>", repliedID, repliedUserName)
		if internal.CheckIfOwnerSudo(repliedID) {
			stats += "<code>Sudo</code>\n\n"
		} else {
			stats += "<code>User</code>\n"
			if internal.Authorized(repliedID) {
				stats += "<u>Authorized :</u><code>True</code>\n\n"
			} else {
				stats += "<u>Authorized :</u><code>False</code>\n\n"
			}
		}
	}

	chatType := update.Message.Chat.Type
	stats += fmt.Sprintf("<strong>Chat: </strong><code>%s</code>\n<u>Type: </u><code>%s</code>\n<u>ID: </u><code>%d</code>\n", chatTitle, chatType, chatID)

	if internal.Authorized(chatID) {
		stats += "<u>Authorized: </u><code>True</code>"
	} else {
		stats += "<u>Authorized: </u><code>False</code>"
	}
	handler.SendMessage(bot, update, stats)
}

func prepareSudo(update tgbotapi.Update, bot *tgbotapi.BotAPI, userId int64) {
	if update.Message.ReplyToMessage != nil {
		if update.Message.ReplyToMessage.From.ID == userId {
			addSudo(userId, update, bot)
		}
	} else if userIDToAuthorize := strings.Split(update.Message.Text, "/sudo "); len(userIDToAuthorize) == 2 {
		userid, err := strconv.ParseInt(strings.TrimSpace(userIDToAuthorize[1]), 10, 64)
		if err == nil && internal.CheckFriends(userid) {
			addSudo(userid, update, bot)
		} else {
			handler.SendMessage(bot, update, "<strong> Wrong ID: Not eligible to become a sudo</strong>")
		}

	} else {
		handler.SendMessage(bot, update, "<strong>First Make Yourself a Sudo</strong><code>/sudo</code>")
	}
}

func sudoRemove(id int64, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if internal.CheckIfOwnerSudo(id) && !(config.EnvFields.OwnerID == id) {
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
	if update.Message.ReplyToMessage != nil && !update.Message.ReplyToMessage.From.IsBot {

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

	} else if success := matchCommandsID(update.Message.Text) ; success != "err" {
		userid, err := strconv.ParseInt(strings.TrimSpace(success), 10, 64)
		if err == nil && isSudo {
			addSudo(userid, update, bot)
			return
		}
		if err == nil && !internal.Authorized(userid) {
			config.EnvFields.AuthorizedUsers[userid] = len(config.EnvFields.AuthorizedUsers) + 1
			handler.SendMessage(bot, update, fmt.Sprintf("<strong><code>%d</code> Is Authorized! By @%s</strong>", userid, currentUserName))
		} else {
			handler.SendMessage(bot, update, fmt.Sprintf("<code>%d</code> <strong>Invalid User ID! Or Chat Is Already Authorized.</strong>", userid))
		}

	} else if internal.Authorized(chatId) {
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

func matchCommandsID(text string) string{
	re := regexp.MustCompile(`(?i)/(authorize|sudo|sudorm|unauthorize)\s+(\d+)`)
	match := re.FindStringSubmatch(text)
	// var err string = -1
	if match != nil {
        command := match[1]
        id := match[2]
        fmt.Printf("Command: %s, ID: %s", command, id)
		return id
    } 
	return "err"
}

func unauthorizeHandler(update tgbotapi.Update, bot *tgbotapi.BotAPI, currentUserName string, chatTitle string, chatId int64, isSudo bool) {
	if update.Message.ReplyToMessage != nil && !update.Message.ReplyToMessage.From.IsBot {
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

	} else if success := matchCommandsID(update.Message.Text) ; success != "err" {
		userid, err := strconv.ParseInt(strings.TrimSpace(success), 10, 64)
		if isSudo && (err == nil) {
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

	} else if internal.Authorized(chatId) {
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
	if strings.HasPrefix(messageText, botName) && strings.HasSuffix(messageText, botName) {
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
	} else if strings.HasPrefix(messageText, "/search ") {
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
			// startTime := time.Now()
			result := drive.Search(queryText)
			if result != "err"{//fmt.Sprintf(config.Msg.ResultText, queryText, result, timeTaken.Milliseconds(), returnedMSG.ReplyToMessage.From.ID, returnedMSG.ReplyToMessage.From.UserName)
				// timeTaken := time.Since(startTime)
				handler.ResultReply(bot, update, result, returnedMSG.Chat.ID, returnedMSG.MessageID)
			} else {
				handler.EditMessage(returnedMSG.Chat.ID, returnedMSG.MessageID, fmt.Sprintf("No Results Found for %s", queryText), bot)
			}
		}()
	}

}

func createRequest(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	handler.KeyboardButton(bot, update)
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

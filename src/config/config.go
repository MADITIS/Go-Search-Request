package config

import "time"

var Start = time.Now()
type config struct {
	BotToken, 
	TmdbAPI,
	RedisURL,
	RedisPassword,
	ServerURL,
	MongoUser,
	MongoPass,
	MongoDBURL   string             
	AuthorizedUsers, SudoOwner map[int64]int
	DriveLists 		map[string]string
	Friends        []int64
	OwnerID,         
	RequestSource, 
	RequestDest, 
	SearchChat int64
}

type help struct {
	Start,
	Help, 
	Search,
	SudoInfo,
	NoAccess,
	UserNoAccess,
	SudoAccess, 
	GotSudoAccess,
	Sudo,
	Sudorm,
	ResultText,
	RequestText,
	TelegraphFormat string
}

var Msg help = help{
	RequestText:"<strong>Make a request</strong> 📢\n\n<code>#request [item name] [link (optional)]</code>\n\n<strong>Description:</strong>\nUse the #request command to request an item. If you have a link related to the item, you can provide it as well.\n\n<strong>Example:</strong>\n<code>#request iPhone 12</code>\n<code>#request MacBook Pro https://www.apple.com/macbook-pro/</code>\n\n📢 <i>Don't see the item you're looking for? Request it using the #request command and we'll try our best to find it for you. Be sure to provide a link if you have one, it'll help us in our search.</i> 📢",
	Help: "🦅 <strong>Welcome to Phoenix!</strong> 🦅\n\n<strong>Commands:</strong>\n<strong>/help</strong> - <code>Show this help text</code>\n<strong>/ping</strong> - <code>Check if the bot is online</code>\n<strong>/search</strong> - <code>Perform searches on cloud drives</code>\n<strong>/start</strong> - <code>Start a new conversation with the bot</code>\n\n🔎 <i>Phoenix</i> is a powerful search bot that helps you find what you need on different cloud drives. 🔎\n\n<strong>With the /start command:</strong> 🚀 <i>Start a new conversation with Phoenix and access its features.</i>\n\n<strong>With the /ping command:</strong> 🔵 <i>Check if Phoenix is online.</i>\n\n<strong>With the /search command:</strong> 🔍 <i>Perform searches on cloud drives and find what you need quickly.</i>\n\n<strong>Sudo commands:</strong>\n<strong>/authorize</strong> - <code>Authorize a user</code>\n<strong>/sudo</strong> - <code>Add users to the sudo list</code>\n<strong>/sudorm</strong> - <code>Remove users from the sudo list</code>\n<strong>/unauthorize</strong> - <code>Unauthorize a user</code>\n\n<strong>If you have any questions or need further assistance, please contact the bot administrator:</strong>",
	Search: "🔍 <strong>Search on Google Drive with Phoenix</strong> 🔍\n\n<code>/search [-f] [-d] [query]</code>\n\n<strong>Description:</strong>\nPerform a search on Google Drive using Phoenix. The maximum number of results is 150 for better performance.\n\n<strong>Optional Flags:</strong>\n<code> -f</code> <i>Only search for files</i>\n<code> -d</code> <i>Only search for folders/directories</i>\n\n<strong>Example:</strong>\n<code>/search -f report</code>\n<code>/search -d meeting notes</code>\n\n🔎 <i>Phoenix</i> makes it easy to find what you need on Google Drive. With the /search command you can search for files and folders, narrowing down your results to find exactly what you're looking for. 🔎\n\n<strong>Note:</strong> For better performance, try to be specific with your search query and use the flags to filter your results.",
	Sudo:"<strong>/sudo</strong> - <code>Add users to the sudo list</code>\n\n<strong>Usage:</strong>\n/sudo [user ID] - Add user with ID [user ID] to the sudo list.\n/sudo (replying to a message from a user) - Add the replied user to the sudo list.\n\n<i>Note:</i> Only available for users with existing sudo access.",
	Sudorm:"<strong>/sudorm</strong> - <code>Remove users from the sudo list</code>\n\n<strong>Usage:</strong>\n/sudorm [user ID] - Remove user with ID [user ID] from the sudo list.\n/sudorm (replying to a message from a user) - Remove the replied user from the sudo list.\n\n<i>Note:</i> Only available for users with existing sudo access.",
	SudoInfo: "🔒<strong>Access Denied</strong> 🔒\n\n<i>It seems %d don't have access to this feature.</i>\n\n<strong>But don't worry!</strong>\n<code>/sudo</code> - <i>You can authorize yourself or others to gain access. Usage: /sudo [userid or reply to user]</i>\n\n🔑 <i>Unlock Phoenix's full potential by authorizing yourself or others.</i>🔑",
	NoAccess: "🔒<strong>Access Restricted</strong> 🔒\n\n<i>It seems %d don't have access to Phoenix's features.</i>\n\n<strong>Contact the admin</strong>\n\n🔑 <i>Unlock Phoenix's full potential by getting authorized.</i>🔑",
	UserNoAccess: "<code>%d<code> 🔒<strong>Access Restricted</strong> 🔒\n\n<i>It seems you don't have access to all Phoenix's features.</i>\n\n<strong>Contact the admins:</strong>\n<a href='https://t.me/MADITIS'>MDS</a>\n\n🔑 <i>Unlock Phoenix's full potential by getting authorized.</i>🔑",
	SudoAccess: "<code>%d</code> Already a sudo user on Phoenix. And have full access to all the features.\n\n<strong>Commands:</strong>\n<strong>/authorize</strong> - <code>Authorize other users</code>\n<strong>/unauthorize</strong> - <code>Unauthorize other users</code>",
	GotSudoAccess: "Congratulations! <code>%d</code> have been granted sudo access on Phoenix. Now have full access to all the features.\n\n<strong>Commands:</strong>\n<strong>/authorize</strong> - <code>Authorize other users</code>\n<strong>/unauthorize</strong> - <code>Unauthorize other users</code>",
	ResultText: "<b>Query:</b> <code>%s</code>\n\n<b>Search Results:%s\n\n<b>Time Taken:</b> <code>%dms</code>\n\n<b>Requested by:</b> <a href='tg://user?id=%d'>@%s</a>",
}


var EnvKeys = map[string]bool{
	"BOT_TOKEN":        true,
	"OWNER_ID":         true,
	"AUTHORIZED_USERS": false,
	"TMDB_API":         false,
	"REQUEST_SOURCE": true,
	"REQUEST_DEST": true,
	"SUDO_OWNER": false,
	"SEARCH_CHAT": true,
	"REDIS_URL": true,
	"REDIS_PASSWORD": true,
	"MONGODB_URL": true,
	"SERVER_URL": true,
	"MONGODB_USER": true,
	"MONGODB_PASS": true,
	"FRIENDS": false,
}
var EnvFields = config{AuthorizedUsers: make(map[int64]int),
					   Friends: []int64{},
					   SudoOwner: make(map[int64]int),
					}

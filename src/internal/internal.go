package internal

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	config "github.com/maditis/search-go/src/config"
)

func Authorized(chatId int64) (bool) {
	groups := config.EnvFields.AuthorizedUsers 
	fmt.Println("Current Groups", groups)
	for value := range groups {
		if value == chatId {
			fmt.Println("Current Group Checking..", value)
			return true
		}
	}
	return false
}


func CheckEnv(load bool, required bool, key string, value string) {
	if value == "" && required{
		Error.Fatalf("The %s environment variable is not set\n", key)
	} else if !load && required{
		Error.Fatalf("The %s environment variable is not set\n", key)
	} else if !load && !required {
		Warning.Printf("The %s environment variable is not set\n", key)
	}
}

func StringToSlice(chatIDS string) ([]int64) {
	stringrgx(chatIDS)
	temp := strings.Split(chatIDS, ",")
	var IDS []int64

	if len(temp) == 1 {
		id := temp[0]
		i , err := strconv.ParseInt(strings.TrimSpace(id), 10, 64)
		ErrorExit(err, fmt.Sprintf("%v Not a Valid ID: Check ENV file", id))
		IDS = append(IDS, i)
		return IDS
	}

	for _, id := range temp {
		i , err := strconv.ParseInt(strings.TrimSpace(id), 10, 64)
		ErrorExit(err, fmt.Sprintf("%v Not a Valid ID: Check ENV file", id))
		IDS = append(IDS, i)	}
	return IDS
}

func stringrgx(chatid string) {
	match, err := regexp.MatchString(`^"-?\d+(?:\s*,\s*-?\d+)*\s*"$`, chatid)
	ErrorExit(err, fmt.Sprintf("Could Not Match the %s: Please correct the ENV value", chatid))

	if !match {
		ErrorExit(err, fmt.Sprintf("Could Not Match the %s: Please correct the ENV value", chatid))
	}
}

func ConvertToInt(id string) (int64) {
	i , err := strconv.ParseInt(strings.TrimSpace(id), 10, 64)
	ErrorExit(err, fmt.Sprintf("%v Not a Valid ID: Check ENV file", id))
	return i
}

// func ConvertToID(id string) (int64) {
// 	i , err := strconv.ParseInt(strings.TrimSpace(id), 10, 64)
// 	WarningLog(err, fmt.Sprintf("%v Not a Valid ID: Check ENV file", id))
// 	return i
// }

func CheckIfOwnerSudo(id int64) (bool) {
	if id == config.EnvFields.OwnerID {
		return true
	} else {
		for key := range config.EnvFields.SudoOwner {
			if key == id {
				return true
			}
		}
	}
	return false
}

func CheckFriends(id int64) (bool) {
	for _, item := range config.EnvFields.Friends {
		if id == item {
			return true
		}
	}
	return false
}

func CleanQuery(query *string) {
	*query = strings.ReplaceAll(*query, "'", "\\'")
}
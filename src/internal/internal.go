package internal

import (
	"fmt"
	"os"
	"path/filepath"
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
		Error.Fatalf("More OR %s environment variable is not set\n", key)
	} else if !load && required{
		Error.Fatalf("More OR %s environment variable is not set\n", key)
	} else if !load && !required {
		Warning.Printf("More OR %s environment variable is not set\n", key)
	}
}

func StringToSlice(chatIDS string, IDS *[]int64) {
	idstring := stringrgx(chatIDS)
	if len(idstring) >= 1 {
			for _, id := range idstring {
			i , err := strconv.ParseInt(strings.TrimSpace(id), 10, 64)
			ErrorExit(err, fmt.Sprintf("%v Not a Valid ID: Check ENV file", id))
			*IDS = append(*IDS, i)
		}
	}

	// if len(temp) == 1 {
	// 	id := temp[0]
	// 	i , err := strconv.ParseInt(strings.TrimSpace(id), 10, 64)
	// 	ErrorExit(err, fmt.Sprintf("%v Not a Valid ID: Check ENV file", id))
	// 	IDS = append(IDS, i)
	// 	return IDS
	// }

	// for _, id := range temp {
	// 	i , err := strconv.ParseInt(strings.TrimSpace(id), 10, 64)
	// 	ErrorExit(err, fmt.Sprintf("%v Not a Valid ID: Check ENV file", id))
	// 	IDS = append(IDS, i)	}
	// return IDS
}

func stringrgx(chatid string) []string{
	re := regexp.MustCompile(`([-?0-9]+\\s*,\\s*)*[-?0-9]+`)
    ids := re.FindAllString(chatid, -1)
	return ids
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

func GetBaseFolder(folder string) string {
	println(os.Getwd())
	currectWD, err := os.Getwd()
	if err != nil {
		ErrorExit(err, "Could not find the working directory")
	}
	// baseDir := filepath.Dir(currectWD)
	Dir := filepath.Join(currectWD, folder)
	println(Dir)
	_, err = os.Stat(Dir)
	ErrorExit(err, fmt.Sprintf("%s does not exist", folder))
	return Dir
}
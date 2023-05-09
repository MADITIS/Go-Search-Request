package init

import (
	"encoding/json"
	"os"
	"sync"

	env "github.com/joho/godotenv"
	config "github.com/maditis/search-go/src/config"
	"github.com/maditis/search-go/src/database"
	internal "github.com/maditis/search-go/src/internal"
)


func init() {
	println(os.Getwd())
	// if len(os.Args) > 1 && os.Args[1] == "start" {
	// 	internal.Info.Println("App Started!")
	// } else {
	// 	internal.Error.Fatal("Wrong Command: type start")
	// }
	internal.Info.Println("Started Init!")

	var wg sync.WaitGroup

	loadEnv()
	wg.Add(2)
	internal.Info.Println("Getting env variables")
	go func () {
		getEnv()
		wg.Done()
	}()

	go func() {
		getDrives()
		wg.Done()
	}()
	
	// go func() {
	// 	graph.Graph()
	// 	wg.Done()
	// }()

	wg.Wait()
	internal.Info.Println("Starting database")
	database.InitRedis()
	database.InitMongo()
	
}

func loadEnv() {
	const folder string = "src/config/config.env"
	dir := internal.GetBaseFolder(folder)
	err := env.Load(dir)
	internal.ErrorExit(err, "Error loading config.env file")
}

func getEnv() {
	for envKey, isRequired := range config.EnvKeys{
		envValue, ok := os.LookupEnv(envKey)
		internal.CheckEnv(ok, isRequired, envKey, envValue)

		if envValue == "" {
			continue
		}

		if envKey == "AUTHORIZED_USERS" {
			var authorizedUsers []int64
			internal.StringToSlice(envValue, &authorizedUsers)
			for i, value := range authorizedUsers {
				config.EnvFields.AuthorizedUsers[value] = i+1
			}
		}else if envKey == "OWNER_ID" {
			ownerID := internal.ConvertToInt(envValue)
			config.EnvFields.OwnerID = ownerID
		}else if envKey == "BOT_TOKEN" {
			config.EnvFields.BotToken = envValue
		}else if envKey == "TMDB_API" {
			config.EnvFields.TmdbAPI = envValue
		}else if envKey == "REQUEST_SOURCE" {
			ID := internal.ConvertToInt(envValue)
			config.EnvFields.RequestSource = ID
		}else if envKey == "REQUEST_DEST" {
			ID := internal.ConvertToInt(envValue)
			config.EnvFields.RequestDest = ID
		} else if envKey == "SUDO_OWNER" {
			var sudoOwners []int64 
			internal.StringToSlice(envValue, &sudoOwners)
			for i, value := range sudoOwners {
				config.EnvFields.SudoOwner[value] = i+1
			}
		} else if envKey == "SEARCH_CHAT" {
			ID := internal.ConvertToInt(envValue)
			config.EnvFields.SearchChat = ID
		} else if envKey == "REDIS_URL" {
			config.EnvFields.RedisURL = envValue
		} else if envKey == "REDIS_PASSWORD" {
			config.EnvFields.RedisPassword = envValue
		} else if envKey == "MONGODB_URL" {
			config.EnvFields.MongoDBURL = envValue
		} else if envKey == "SERVER_URL" {
			config.EnvFields.ServerURL = envValue
		}else if envKey == "MONGODB_USER" {
			config.EnvFields.MongoUser = envValue
		} else if envKey == "MONGODB_PASS" {
			config.EnvFields.MongoPass = envValue
		}else if envKey == "FRIENDS" {
			var friends []int64
			internal.StringToSlice(envValue, &friends)
			copy(config.EnvFields.Friends, friends)
		}

	}
	internal.Info.Printf("ENV Successfully Set! %v", config.EnvFields.AuthorizedUsers)
}

func getDrives() {
	internal.Info.Println("Fetching the drives!")
	// currentPath, _ := os.Executable()
	const folder string = "src/config/driveList.json"
	drive_file := internal.GetBaseFolder(folder)
	fileBytes, _ := os.ReadFile(drive_file)
	err := json.Unmarshal(fileBytes, &config.EnvFields.DriveLists)
	internal.ErrorExit(err, "Error reading Drive JSON file")
	internal.Info.Println("Successfully Fetched the drives: ", config.EnvFields.DriveLists)
}
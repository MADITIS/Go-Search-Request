package init

import (
	"encoding/json"
	"os"
	"sync"

	env "github.com/joho/godotenv"
	config "github.com/maditis/search-go/src/config"
	internal "github.com/maditis/search-go/src/internal"
	graph "github.com/maditis/search-go/src/telegraph"
)


func init() {
	if len(os.Args) > 1 && os.Args[1] == "start" {
		internal.Info.Println("App Started!")
	} else {
		internal.Error.Fatal("Wrong Command: type start")
	}
	internal.Info.Println("Started Init!")

	var wg sync.WaitGroup

	loadEnv()
	wg.Add(3)
	go func () {
		getEnv()
		wg.Done()
	}()

	go func() {
		getDrives()
		wg.Done()
	}()
	
	go func() {
		graph.Graph()
		wg.Done()
	}()

	wg.Wait()
}

func loadEnv() {
	err := env.Load("src/config/config.env")
	internal.ErrorExit(err, "Error loading .env file")
}

func getEnv() {
	for envKey, isRequired := range config.EnvKeys{
		envValue, ok := os.LookupEnv(envKey)
		internal.CheckEnv(ok, isRequired, envKey, envValue)

		if envValue == "" {
			continue
		}

		if envKey == "AUTHORIZED_USERS" {
			authorizedUsers := internal.StringToSlice(envValue)
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
		}

	}
	internal.Info.Printf("ENV Successfully Set! %v", config.EnvFields)
}

func getDrives() {
	internal.Info.Println("Getting the drives!")
	fileBytes, _ := os.ReadFile("/src/config/driveList.json")
	err := json.Unmarshal(fileBytes, &config.EnvFields.DriveLists)
	internal.ErrorExit(err, "Error reading JSON file")
	internal.Info.Println("Successfuly Fetched the drives: ", config.EnvFields.DriveLists)
}

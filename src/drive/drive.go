package drive

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"

	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"github.com/maditis/search-go/src/config"
	"github.com/maditis/search-go/src/database"
	internal "github.com/maditis/search-go/src/internal"
	graph "github.com/maditis/search-go/src/telegraph"
	"github.com/maditis/search-go/src/web"
)

type driveService struct {
	service *drive.Service
}

func getService(s driveService) *drive.Service {
	return s.service
}
var ModeText = "full"
var middleware = driveService{}

const telegraphLimit int = 45


var wg sync.WaitGroup

var SearchType = true

func CreateService() {
	// get sa file
	saFile := internal.GetSA()
	// get necessary configs to create service
	config := getConfig(saFile)
	ctx := context.Background()
	client := config.Client(ctx)

	// create service
	service, err := drive.NewService(ctx, option.WithHTTPClient(client))
	internal.ErrorExit(err, "Unable to Create Service")

	internal.Info.Println("Service Successfully Created", service)
	middleware.service = service
}

// Return JsonKey for the JWTConfigFromJSON
func getJsonKey(saFile string) []byte {
	// currentPath, _ := os.Executable()
	const folder string = "accounts"
	accountDir := internal.GetBaseFolder(folder)
	jsonKey, err := os.ReadFile(accountDir + "/" + saFile)
	internal.ErrorExit(err, "Invalid SA file")

	return jsonKey
}

// Returns config to create client
func getConfig(saFile string) *jwt.Config {
	driveScope := "https://www.googleapis.com/auth/drive"
	jsonKey := getJsonKey(saFile)
	config, err := google.JWTConfigFromJSON(jsonKey, driveScope)
	internal.ErrorExit(err, "Could Not create Service")
	return config
}

func searchHelper(query *string, q *string) {
	*query = strings.ToLower(*query)
	if strings.Contains(*query, "-f") {
		*query = strings.TrimSpace(strings.Split(*query, "-f")[1])
		if SearchType {
			*q = fmt.Sprintf("fullText contains '%s' and mimeType != 'application/vnd.google-apps.folder' and trashed=false", *query)
		} else {
			*q = fmt.Sprintf("name contains '%s' and mimeType != 'application/vnd.google-apps.folder' and trashed=false", *query)
		}
	} else if strings.Contains(*query, "-d") {
		*query = strings.TrimSpace(strings.Split(*query, "-d")[1])
		if SearchType {
			*q = fmt.Sprintf("fullText contains '%s' and mimeType = 'application/vnd.google-apps.folder' and trashed=false", *query)
		} else {
			*q = fmt.Sprintf("name contains '%s' and mimeType = 'application/vnd.google-apps.folder' and trashed=false", *query)
		}
	}
	if SearchType {
		*q = fmt.Sprintf("fullText contains '%s' and trashed=false", *query)
	} else {
		*q = fmt.Sprintf("name contains '%s' and trashed=false", *query)
	}

}

func Search(query string) string {
	var q = ""
	internal.CleanQuery(&query)
	searchHelper(&query, &q)
	fmt.Printf("Starting get value %s", query)
	_, err := database.GetValue(query)
	url := fmt.Sprintf("%s?q=%s", config.EnvFields.ServerURL,  url.PathEscape(query))
	if err {
		web.Results.Query = query
		return url
	}

	errc, results := cacheSearch(q)
	if errc == -1 {
		return "err"
	}
	web.Results.AllFiles = results
	web.Results.Query = query
	database.SetValue(query, results)
	return url
}

func cacheSearch(q string) (int,  []map[string]string){
	var allFiles []map[string]string

	totalDrives := len(config.EnvFields.DriveLists)
	i := 0
	for _, id := range config.EnvFields.DriveLists {
		i++
		
		internal.Info.Println("Searching in ", id)
		wg.Add(1)
		go func() {
			getAllFiles(id, q, &allFiles)
			wg.Done()
		}()
		internal.Info.Printf("Progress Go Routines: %d/%d", i, totalDrives)
	}
	wg.Wait()
	numberPages := len(allFiles)
	fmt.Println("\nAll goroutines finished!")
	if numberPages == 0 {
		internal.Warning.Println("No Results Found")
		return -1, allFiles
	}

	// web.ServeResults()
	internal.Info.Println("Results: ", numberPages)
	// return setupTelegraph(allFiles, query), numberPages, nil
	return 1, allFiles
}

func setupTelegraph(allFiles []map[string]string, query string) string {
	var telegraphContent []string
	var telegraphPath []string
	msg := fmt.Sprintf("<h4>Query</h4><strong>Search Results For: </strong><code>%s</code><br>", query)
	for i, item := range allFiles {
		msg += fmt.Sprintf(`📁<code>%s</code><b>(%s)</b><br><b><a href='%s'>Drive Link</a></b><br><br>`, item["name"], item["mimeType"], item["link"])
		if i == telegraphLimit {
			telegraphContent = append(telegraphContent, msg)
			msg = ""
		}
	}

	telegraphContent = append(telegraphContent, msg)
	totalPages := len(telegraphContent)
	accounts := graph.Conf.AccessToken
	totalAccounts := len(accounts)

	accNo := -1
	pagePerAcc := 2
	for i := 0; i < totalPages; i++ {

		if i%pagePerAcc == 0 {
			accNo = (accNo + 1) % totalAccounts
		}

		if i != 0 {
			telegraphContent[i] += fmt.Sprintf(`<br><b><a href="https://graph.org/%s">Previous</a> | Page %d/%d</b>`, telegraphPath[i-1], i+1, totalPages)
		} else {
			telegraphContent[i] += fmt.Sprintf(`<br><b>Page %d/%d</b>`, i+1, totalPages)
		}
		path := graph.CreatePage("Drive Search", telegraphContent[i], accounts[accNo])
		if path != "" {
			telegraphPath = append(telegraphPath, path)
		}

		if i != 0 {
			telegraphContent[i-1] += fmt.Sprintf(`<b><a href="https://graph.org/%s"> Next</a></b>`, telegraphPath[i])
			temp := accNo
			if i%pagePerAcc == 0 {
				temp = accNo - 1
			}
			graph.EditPage(telegraphPath[i-1], "Test", telegraphContent[i-1], accounts[temp])
		}
	}

	// handler.SendMessage(bot, update, fmt.Sprintf("Results: https://graph.org/%s", telegraphPath[0]))
	return telegraphPath[0]
}

func getAllFiles(driveID string, q string, allFiles *[]map[string]string) {
	var token string
	service := getService(middleware)
	for {
		result, err := service.Files.List().
			DriveId(driveID).
			Q(q).
			Corpora("drive").
			SupportsAllDrives(true).
			IncludeItemsFromAllDrives(true).
			PageSize(500).
			Fields("nextPageToken, files(id, name, webViewLink, mimeType, size)").
			PageToken(token).
			Do()
		if err != nil {
			msg := fmt.Sprintf("<strong>Unable to retrieve files:</strong> <code>%v: %s</code>", err, q)
			internal.Error.Printf(msg)
			// handler.SendMessage(bot, update, msg)
			continue
		}

		files := result.Files
		var wg sync.WaitGroup
		for _, file := range files {
			// if len(*allFiles) >= contentLimit {
			// 	return
			// }
			fmt.Printf("type: %s, size: %d\n", file.MimeType, file.Size)

			*allFiles = append(*allFiles, map[string]string{
				"name": file.Name,
				// "id": file.Id,
				"link":     file.WebViewLink,
				"mimeType": file.MimeType,
			})
		}
		wg.Wait()
		token = result.NextPageToken
		if token == "" {
			break
		}
		fmt.Println("Found Next Token: ", token)
	}
	// return true
}

// func getStartToken(service *drive.Service) string{
// 	results, err := service.Changes.GetStartPageToken().
// 									DriveId("0ALuOVeNOGJ8CUk9PVA").
// 									SupportsAllDrives(true).
// 									Do()
// 	if err != nil {
// 		fmt.Println(err)
		
// 	}
// 	return results.StartPageToken
// }


// func WatchDrive() {
// 	fmt.Printf("Watching is going on")
// 	service := getService(middleware)

// 	var pageToken string = getStartToken(service)
// 	for {
// 		fmt.Println("started watichgn")
// 		results, err := service.Changes.List(pageToken).
// 			DriveId("0ALuOVeNOGJ8CUk9PVA").
// 			IncludeItemsFromAllDrives(true).
// 			SupportsAllDrives(true).
// 			PageSize(1000).
// 			Do()
	
// 		if err != nil {
// 			fmt.Println(err)
// 			continue
// 		}
		
// 		fmt.Println("Changes:")
// 		for _, change := range results.Changes {
// 			if change.File != nil {
// 				if !database.GetValue(change.FileId) {
// 					fmt.Println("Change Type:", change.ChangeType)
// 					fmt.Println("File ID:", change.FileId)
// 					fmt.Println("File Name:", change.File.Name)
// 					database.SetValue(change.FileId, change.FileId)
// 				}
// 			}

// 		}

// 		if results.NextPageToken!= ""{
// 			pageToken = results.NewStartPageToken
// 		}

	
// 			// Add a delay before the next iteration
// 		time.Sleep(10 * time.Second)
// 	}
// }

// func getSize(file *drive.File) string {
// 	var size string = ""
// 	if file.MimeType != "application/vnd.google-apps.folder" {
// 		size = getReadableSize(file.Size)

// 	} else {
// 		size = "Folder"
// 	}

// 	return size
// }

// func getReadableSize(size int64) string {
// 	var oneGb int64 = 1024 * 1024 * 1024
// 	var realSize string = ""
// 	if size <= oneGb {
// 		realSize = fmt.Sprintf("%0.2f", math.Round(float64(size / (1024 * 1024))))
// 		return realSize
// 	}
// 	realSize = fmt.Sprintf("%0.2f", math.Round(float64(size / (1024 * 1024 * 1024))))
// 	return realSize
// }
// func getDrives(service *drive.Service) {
// 	drives, err := service.Drives.List().Do()

// 	if err != nil {
// 		internal.Fatal("No Drives available")
// 	}
// 	if len(drives.Drives) > 0 {
// 		for _, i := range drives.Drives {
// 			fmt.Printf("%s (%v)\n", i.Name, i)
// 		}
// 	} else {
// 		fmt.Print("No drives found.")
// 	}
// }
// sort.Slice(result.Files, func (i, j int) bool {
// 	return result.Files[i].Size > result.Files[j].Size
// })

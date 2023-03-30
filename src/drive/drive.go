package drive

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	// "encoding/json"
	// "fmt"
	// "internal"
	// "net/http"
	// "os"

	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	// "google.golang.org/api/option"
	"github.com/maditis/search-go/src/config"
	internal "github.com/maditis/search-go/src/internal"
	graph "github.com/maditis/search-go/src/telegraph"
	// sa "github.com/maditis/search-go/src/internal"
)

type driveService struct {
	service *drive.Service
}

func getService(s driveService) *drive.Service {
	return s.service
}

var middleware = driveService{}
const telegraphLimit int = 45
const contentLimit int = 150
var wg sync.WaitGroup

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
func getJsonKey(saFile string) ([]byte) {
	jsonKey, err := os.ReadFile("D:\\Programming\\GO\\search app\\accounts\\"+saFile)
	internal.ErrorExit(err, "Invalid SA file")

	return jsonKey
}

// Returns config to create client
func getConfig(saFile string) (*jwt.Config){
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
		*q = fmt.Sprintf("mimeType != 'application/vnd.google-apps.folder' and (name contains '%s' or fullText contains '%s') and trashed = false", *query, *query)
	} else if strings.Contains(*query, "-d") {
		*query = strings.TrimSpace(strings.Split(*query, "-d")[1])
		*q = fmt.Sprintf("mimeType = 'application/vnd.google-apps.folder' and (name contains '%s' or fullText contains '%s') and trashed = false", *query, *query)
	}
	*q = fmt.Sprintf("(name contains '%s' or fullText contains '%s') and trashed = false", *query, *query)
 
}

func Search(query string) (string, int, error) {
	var q = ""
	internal.CleanQuery(&query)
	searchHelper(&query, &q) 


	var allFiles []map[string]string
	totalDrives := len(config.EnvFields.DriveLists)
	i := 0
	for _, id := range config.EnvFields.DriveLists {
		i++
		wg.Add(1)
		internal.Info.Println("Searching in ", id)
		go func() {
			getAllFiles(id, q, &allFiles)	
			wg.Done()
			}()
		internal.Info.Printf("Progress Go Routines: %d/%d", i, totalDrives)
		}
	wg.Wait()
	fmt.Println("\nAll goroutines finished!")
	numberPages := len(allFiles)
	if numberPages == 0 {
		internal.Warning.Println("No Results Found")
		return "No results",numberPages, errors.New("no results found")
	}
	internal.Info.Println("Results: ", numberPages)
	return setupTelegraph(allFiles, query), numberPages, nil
}

func setupTelegraph(allFiles []map[string]string, query string) string {
	var telegraphContent []string
	var telegraphPath []string
	msg := fmt.Sprintf("<h4>Query</h4><strong>Search Results For: </strong><code>%s</code><br>", query)
	for i, item := range allFiles {
		msg += fmt.Sprintf(`📁<code>%s</code><b>(%s)</b><br><b><a href='%s'>Drive Link</a></b><br><br>`, item["name"],item["mimeType"], item["link"])
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
	for i:= 0; i < totalPages; i++ {
		
		if i % pagePerAcc == 0{
			accNo = (accNo+1) % totalAccounts
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
			telegraphContent[i - 1] +=  fmt.Sprintf(`<b><a href="https://graph.org/%s"> Next</a></b>`, telegraphPath[i])
			temp := accNo
			if i % pagePerAcc == 0 {
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
					PageSize(200).
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
			if len(*allFiles) == contentLimit {
				fmt.Println("Found More than the limit, Exiting.")
				return
			}

			fmt.Printf("type: %s, size: %d\n", file.MimeType, file.Size)

			*allFiles = append(*allFiles, map[string]string{
				"name": file.Name,
				// "id": file.Id,
				"link": file.WebViewLink,
				"mimeType": file.MimeType,
			})
		}
		wg.Wait()
		token = result.NextPageToken
		if token == ""{
			break
		}
		fmt.Println("Found Next Token: ", token)
	}
	// return true
}

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
package graph

// Depreciated package

import (
	"net/http"
	"time"

	telegraph "github.com/anonyindian/telegraph-go/v2"
	internal "github.com/maditis/search-go/src/internal"
)

type graphConfig struct {
	client *telegraph.TelegraphClient
	AccessToken []string
}

var Conf graphConfig

func createAccount(client *telegraph.TelegraphClient) {
	account, err := client.CreateAccount("MDS", &telegraph.CreateAccountOpts{
		AuthorName: "The Alchemist",
	})

	if err != nil {
		internal.Error.Println("Could Not Create Telegraph Account", err.Error())
		time.Sleep(5 * time.Second)
		createAccount(client)
	}

	time.Sleep(1 * time.Second)
	internal.Info.Println("Created The Telegraph Account!")
	Conf.AccessToken = append(Conf.AccessToken, account.AccessToken)

}

func Graph() {
	client := telegraph.GetTelegraphClient(&telegraph.ClientOpt {
		HttpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	})
	for i := 0; i < 8; i++ {
		createAccount(client)
	}

	Conf.client = client

	// tet := client.GetAccountInfo()
	// fmt.Println("acesstokenn",account.AccessToken)
}
func CreatePage(title string, content string, accessToken string) string {
	_, err := Conf.client.CreatePage(accessToken, title, content, nil)
	if err != nil {
		internal.Error.Println("Could Not Create Page!", err.Error())
		time.Sleep(1 * time.Second)
		CreatePage(title, content, accessToken)
	}
	internal.Info.Println("Successfully Created The page")


	plist, _ := Conf.client.GetPageList(accessToken, &telegraph.PageListOpts{
		Limit: 1,
	})
	path := ""
	for _, page := range plist.Pages {
		path = page.Path
		// you can print all pages with the help of loop
	}
	return path
}

func EditPage(path string, title string, content string, accessToken string) {
	_, err := Conf.client.EditPage(accessToken, path, title, content, nil)
	if err != nil {
		if err != nil {
		internal.Error.Println("Could Not Edit Page!", err.Error())
		time.Sleep(1 * time.Second)
		EditPage(accessToken, path, title, content)
		}
	}
	internal.Info.Println("Successfully Edited The page")
}
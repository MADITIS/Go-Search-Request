package main

import (
	drive "github.com/maditis/search-go/src/drive"
	_ "github.com/maditis/search-go/src/init"
	telegram "github.com/maditis/search-go/src/telegram"
	"github.com/maditis/search-go/src/web"
)

func main() {
	drive.CreateService()
	go web.StartServer()
	botStart()
}

func botStart() {
	telegram.Start()
}
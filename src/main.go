package main

import (
	drive "github.com/maditis/search-go/src/drive"
	_ "github.com/maditis/search-go/src/init"
	telegram "github.com/maditis/search-go/src/telegram"
)

func main() {
	drive.CreateService()
	botStart()
}

func botStart() {
	telegram.Start()
}
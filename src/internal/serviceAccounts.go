package internal

import (
	"os"
)

type State struct {
	count int
}

var state State = State{
	count: 0,
}

func GetSA() (string){
	dir := GetBaseFolder("accounts")
	files, err := os.ReadDir(dir)
	ErrorExit(err, "Invalid Accounts folder.")
	
	sa := files[state.count].Name()
	return sa
}
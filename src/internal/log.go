package internal

import (
	"log"
	"os"
)

var (
	Warning *log.Logger = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	Info    *log.Logger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error   *log.Logger = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
)

func ErrorExit(err error, errMSG string){
	if err != nil {
		Error.Fatalf("%v: %s", err, errMSG)
	}
}

func WarningLog(err error, errMSG string) {
	if err != nil {
		Warning.Printf("%v: %s\n", err, errMSG)
	}
}


package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/maditis/search-go/src/database"
	"github.com/maditis/search-go/src/internal"
)


type results struct {
	AllFiles []map[string]string
	Query string
}

var Results results
var router *gin.Engine

func StartServer() {
	gin.SetMode(gin.DebugMode)
	router = gin.Default()
	staticDIR := internal.GetBaseFolder("src/web")
	router.Static("/css", staticDIR + "/templates/css/")
	router.Static("/img", staticDIR + "/templates/img/")
	tmp, _ := template.ParseGlob(staticDIR + "/templates/public/results.html")
	router.SetHTMLTemplate(tmp)
	go router.GET("/results", func(ctx *gin.Context) {
		
		query := ctx.Query("q")
		fmt.Print("started first get", query)
			

		val, err := database.GetValue(query)
		if err {
			var result []map[string]string
			er := json.Unmarshal([]byte(val), &result) 
			if er != nil {
				fmt.Print("Can't convert")
			} else {
				ctx.HTML(http.StatusOK,"results.html", gin.H{
					"AllFiles": result,
					"Query": query,
				})
			}
		} else {
			ctx.HTML(http.StatusOK,"results.html", gin.H{
				"AllFiles": Results.AllFiles,
				"Query": query,
			})
		}				
	})
	router.Run(":8080")
	fmt.Printf("Started Webserver")

	
}

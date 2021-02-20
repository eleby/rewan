package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"rewan/models"
)

//Function used to pass global variables easier, and to simplify the code
func Render(c *gin.Context, args gin.H) gin.H {
	log.Print("http.go > Render")
	//The title still has to be passed in the args list, so we get it from there
	title := fmt.Sprintf("%v", args["title"])
	//We check if the cookie banner has to be included
	_, err := c.Cookie("acceptCookies")
	//Init the result map with global variables and the title
	result := gin.H{
		//The title variable might need to be translated, so we use a key to get the message in the correct language
		"title": MsgStr("title."+title, GetLang(c), nil) + " | Rewan",
		//We get the language from the context, the language being set beforehand by the global middleware
		"lang": GetLang(c),
		//We get the version from the configuration to display it in the footer
		"version":            models.Conf.Version,
		"domain":             models.Conf.BaseUrl,
		"hasAcceptedCookies": err == nil,
	}

	//We then add every additional argument which is passed to the args variable
	//The title being always set, we can assume the args map will never be nil
	for key, element := range args {
		//We don't want to set the title twice so we ignore it here
		if key != "title" {
			result[key] = element
		}
	}
	return result
}

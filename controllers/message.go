package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"html/template"
	"log"
	"rewan/models"
)

//Create the bundle var, initialized at the default language
var bundle = i18n.NewBundle(language.English)

//Initialize the bundle with format and references to the files
func InitBundle() {
	log.Print("message.go > InitBundle")
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	bundle.MustLoadMessageFile("fr.json")
	bundle.MustLoadMessageFile("en.json")
}

//GetLang fetches the language from the context, and returns only the first two letters
func GetLang(c *gin.Context) string {
	log.Print("message.go > GetLang")
	lang, err := c.Get("lang")
	if err != false {
		return lang.(string)[:2]
	} else {
		return ""
	}
}

//Returns the message as an HTML template - Called in templates
//Takes a message id, the current language and an extra parameter as entries
func Msg(id string, lang string, param interface{}) template.HTML {
	log.Printf("message.go > Msg [ %v / %v / %v ]", id, lang, param)
	//Get the localizer
	loc := i18n.NewLocalizer(bundle, lang)
	//Get the message by it's id, the localizer and an eventual parameter
	translation := loc.MustLocalize(&i18n.LocalizeConfig{
		MessageID: id,
		TemplateData: map[string]interface{}{
			"param": param,
		},
	})
	//Returns the translation as an HTML template
	return template.HTML(translation)
}

//Returns the message as a string - Called server-side
//Takes a message id, the current language and an extra parameter as entries
func MsgStr(id string, lang string, param interface{}) string {
	log.Printf("message.go > MsgStr [ %v / %v / %v ]", id, lang, param)
	//Get the localizer
	loc := i18n.NewLocalizer(bundle, lang)
	//Get the message by it's id, the localizer and an eventual parameter
	translation := loc.MustLocalize(&i18n.LocalizeConfig{
		MessageID: id,
		TemplateData: map[string]interface{}{
			"param": param,
		},
	})
	//Returns the translation as a string
	return translation
}

//Function creating the cookies storing the flash type and message
func Flashvar(typeFlash string, messageFlash string, param interface{}, c *gin.Context) {
	log.Printf("message.go > Flashvar [ %v / %v / %v ]", typeFlash, messageFlash, param)
	isSecure := gin.Mode() == gin.ReleaseMode
	message := MsgStr(messageFlash, GetLang(c), param)
	c.SetCookie("flashType", typeFlash, 3600, "/", models.Conf.BaseUrl, isSecure, false)
	c.SetCookie("flashMessage", message, 3600, "/", models.Conf.BaseUrl, isSecure, false)
}

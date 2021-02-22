package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"rewan/controllers"
	"rewan/controllers/routines"
	"rewan/models"
	"rewan/models/memorygame"
	"rewan/models/persistence"
	"strconv"
)

/** AUTHOR : Erwan Le Bihan **/
/** The source code is open for anyone to see it. **/

//Functions we can use in templates
var funcMap = template.FuncMap{
	"msg":         controllers.Msg,
	"timeMachine": controllers.TimeMachine,
	"year":        controllers.GetCurrentYear,
	"pieceColor":  memorygame.GetColorPiece,
	"cornerClass": memorygame.GetCornerClass,
}

func main() {
	releaseMode := os.Args[2] != "dev"
	//The app goes release mode when the configuration user is not dev
	if releaseMode {
		gin.SetMode(gin.ReleaseMode)
	}
	//Set logger behavior
	//Define the log files, open or create them with full permissions granted
	f, _ := os.OpenFile(models.Conf.LogDir+models.Conf.LogFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	ferr, _ := os.OpenFile(models.Conf.LogDir+models.Conf.ErrLogFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	//Set the gin logger to also write in the file
	multiWriter := io.MultiWriter(f, os.Stdout)
	errWriter := io.MultiWriter(ferr, os.Stdout)
	gin.DefaultWriter = multiWriter
	gin.DefaultErrorWriter = errWriter
	//Add a prefix to the alternative logger
	log.SetPrefix("[LOG] ")
	//Set the other logger to also write in the file
	log.SetOutput(multiWriter)
	//Init gin router
	r := gin.New()
	//Init language files
	controllers.InitBundle()
	//This path is used for resources (css, img, js). The sitemap can also be placed here
	r.Static("/resources", "./resources")
	//Declares which functions we can use in templates
	r.SetFuncMap(funcMap)
	//This path is used for templates.
	//** fetches every directory in /views, and * fetches every template in each directory
	r.LoadHTMLGlob("views/**/*")
	//Global middleware, used before every request. Allows to always get the preferred language
	r.Use(func(c *gin.Context) {
		//Get the preferred language from the "Accept-Language" request header
		lang := c.Request.Header.Get("Accept-Language")
		//Set the lang parameter to be used by the render function
		c.Set("lang", lang)
		c.Next()
	})
	//Recovers from any Panic and render a 500 error page
	//while logging an error id to allow simple retrieval via
	//the id we communicate on the page
	r.Use(gin.CustomRecovery(func(c *gin.Context, err interface{}) {
		errorCode := persistence.GenerateUID()
		var logger *log.Logger
		logger = log.New(gin.DefaultErrorWriter, "[ERR] ", log.LstdFlags)
		errStr := []byte(fmt.Sprintf("%v", err))
		logger.Print("Error '" + string(errStr) + "' logged with " + errorCode)
		lang := controllers.GetLang(c)
		title := controllers.MsgStr("500.title", lang, string(errStr))
		content := controllers.MsgStr("500.content", lang, errorCode)
		c.HTML(http.StatusInternalServerError, "error.html", controllers.Render(c, gin.H{"title": "500",
			"titleError": template.HTML(title), "contentError": template.HTML(content)}))
	}))

	//Set the Gin logger to ignore logging the fetching of resources (ex: images)
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{SkipPaths: []string{"/resources/*"}}))
	//Routes to be handled by the router
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", controllers.Render(c, gin.H{"title": "index"}))
	})
	r.GET("/career", func(c *gin.Context) {
		c.HTML(http.StatusOK, "career.html", controllers.Render(c, gin.H{"title": "career"}))
	})
	r.GET("/projects", func(c *gin.Context) {
		c.HTML(http.StatusOK, "projects.html", controllers.Render(c, gin.H{"title": "projects"}))
	})
	r.GET("/contact", func(c *gin.Context) {
		c.HTML(http.StatusOK, "contact.html", controllers.Render(c, gin.H{"title": "contact"}))
	})
	r.GET("/shelf", func(c *gin.Context) {
		c.HTML(http.StatusOK, "shelf.html", controllers.Render(c, gin.H{"title": "shelf"}))
	})
	r.GET("/about/bootstrap", func(c *gin.Context) {
		lang := controllers.GetLang(c)
		titleContent := controllers.MsgStr("bootstrap.main.title", lang, nil)
		textContent := controllers.MsgStr("bootstrap.main.content", lang, nil)
		c.HTML(http.StatusOK, "wiki.html", controllers.Render(c,
			gin.H{"title": "bootstrap", "image": "bootstrap.svg",
				"credits": "", "titleContent": titleContent, "textContent": textContent,
				"website": "https://getbootstrap.com/", "wiki": "https://en.wikipedia.org/wiki/Bootstrap_(framework)"}))
	})
	r.GET("/about/play", func(c *gin.Context) {
		lang := controllers.GetLang(c)
		titleContent := controllers.MsgStr("play.main.title", lang, nil)
		textContent := controllers.MsgStr("play.main.content", lang, nil)
		c.HTML(http.StatusOK, "wiki.html", controllers.Render(c,
			gin.H{"title": "play", "image": "play.svg",
				"credits": "", "titleContent": titleContent, "textContent": textContent,
				"website": "https://www.playframework.com/", "wiki": "https://en.wikipedia.org/wiki/Play_Framework"}))
	})
	r.GET("/about/letsencrypt", func(c *gin.Context) {
		lang := controllers.GetLang(c)
		titleContent := controllers.MsgStr("letsencrypt.main.title", lang, nil)
		textContent := controllers.MsgStr("letsencrypt.main.content", lang, nil)
		c.HTML(http.StatusOK, "wiki.html", controllers.Render(c,
			gin.H{"title": "letsencrypt", "image": "letsencrypt.svg",
				"credits": "", "titleContent": titleContent, "textContent": textContent,
				"website": "https://letsencrypt.org/", "wiki": "https://en.wikipedia.org/wiki/Let%27s_Encrypt"}))
	})
	r.GET("/about/gandi", func(c *gin.Context) {
		lang := controllers.GetLang(c)
		titleContent := controllers.MsgStr("gandi.main.title", lang, nil)
		textContent := controllers.MsgStr("gandi.main.content", lang, nil)
		c.HTML(http.StatusOK, "wiki.html", controllers.Render(c,
			gin.H{"title": "gandi", "image": "gandi.svg",
				"credits": "", "titleContent": titleContent, "textContent": textContent,
				"website": "https://www.gandi.net/en", "wiki": "https://en.wikipedia.org/wiki/Gandi"}))
	})
	r.GET("/about/go", func(c *gin.Context) {
		lang := controllers.GetLang(c)
		titleContent := controllers.MsgStr("go.main.title", lang, nil)
		textContent := controllers.MsgStr("go.main.content", lang, nil)
		c.HTML(http.StatusOK, "wiki.html", controllers.Render(c,
			gin.H{"title": "go", "image": "go.svg",
				"credits": "", "titleContent": titleContent, "textContent": textContent,
				"website": "https://golang.org/", "wiki": "https://fr.wikipedia.org/wiki/Go_(langage)"}))
	})
	r.GET("/about/gin", func(c *gin.Context) {
		lang := controllers.GetLang(c)
		titleContent := controllers.MsgStr("gin.main.title", lang, nil)
		textContent := controllers.MsgStr("gin.main.content", lang, nil)
		c.HTML(http.StatusOK, "wiki.html", controllers.Render(c,
			gin.H{"title": "gin", "image": "gin.png",
				"credits": "", "titleContent": titleContent, "textContent": template.HTML(textContent),
				"website": "https://gin-gonic.com/"}))
	})
	r.GET("/about/rouen", func(c *gin.Context) {
		lang := controllers.GetLang(c)
		credits := controllers.MsgStr("rouen.credits", lang, nil)
		titleContent := controllers.MsgStr("rouen.main.title", lang, nil)
		textContent := controllers.MsgStr("rouen.main.content", lang, nil)
		c.HTML(http.StatusOK, "wiki.html", controllers.Render(c,
			gin.H{"title": "rouen", "image": "rouen.jpg",
				"credits": credits, "titleContent": titleContent, "textContent": textContent,
				"website": "https://rouen.fr/", "wiki": "https://fr.wikipedia.org/wiki/Rouen"}))
	})
	r.GET("/tools/memory", func(c *gin.Context) {
		c.HTML(http.StatusOK, "memory.html", controllers.Render(c, gin.H{"title": "memory"}))
	})
	r.GET("/tools/memory/game", func(c *gin.Context) {
		game := memorygame.GetSetMemoryGame(c)
		c.HTML(http.StatusOK, "memorygame.html", controllers.Render(c, gin.H{"title": "memory", "board": game, "pieceType": memorygame.PieceType, "gameState": memorygame.GameState}))
	})
	r.GET("/tools/memory/game/reset", memorygame.Reset())
	r.GET("/tools/memory/game/finish", memorygame.Finish())
	r.GET("/tools/time", func(c *gin.Context) {
		c.HTML(http.StatusOK, "time.html", controllers.Render(c, gin.H{"title": "time"}))
	})
	r.GET("/tools/memory/game/row/:rowId/piece/:pieceId/type", memorygame.GetPieceType())
	r.POST("/contact/mail", controllers.SendMail())
	r.GET("/tools/time/get/:query/:year/:month/:day/:hour/:minute", controllers.GetTimeMachine())
	r.GET("/about/2.0", func(c *gin.Context) {
		c.HTML(http.StatusOK, "version.html", controllers.Render(c, gin.H{"title": "version"}))
	})
	r.GET("/sitemap.xml", controllers.RenderSitemap())
	r.GET("/robots.txt", func(c *gin.Context) {
		c.File("robots.txt")
	})
	//Display a page for routes not found (error 404)
	r.NoRoute(func(c *gin.Context) {
		lang := controllers.GetLang(c)
		title := controllers.MsgStr("404.title", lang, c.Request.URL)
		content := controllers.MsgStr("404.content", lang, nil)
		c.HTML(http.StatusInternalServerError, "error.html", controllers.Render(c, gin.H{"title": "404",
			"titleError": template.HTML(title), "contentError": template.HTML(content)}))
	})
	//Generate the sitemap at each startup
	controllers.GenerateSitemap(r)
	//Launch the logs management routine
	routines.ManageLogsRoutine()
	//Launch the server at the configured TCP port
	_ = r.Run(":" + strconv.Itoa(models.Conf.Port))
}

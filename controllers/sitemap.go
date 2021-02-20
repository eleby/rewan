package controllers

import (
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"regexp"
	"rewan/models"
	"time"
)

//Sitemap file path
var FILE string

//Base URL of the website
var BASEURL string

//Header tag for the XML content
var HEADER = "<?xml version='1.0' encoding='UTF-8'?>"

//List of tags used for the XML content
var URLSETOPEN = "<urlset xmlns='http://www.sitemaps.org/schemas/sitemap/0.9'>"
var URLOPEN = "<url>"
var LOCOPEN = "<loc>"
var LOCCLOSE = "</loc>"
var MODOPEN = "<lastmod>"
var MODCLOSE = "</lastmod>"
var URLCLOSE = "</url>"
var URLSETCLOSE = "</urlset>"

//XML content we append elements to
var txt string

//Called from the main function, it generates the sitemap.
//The Engine param allows to get a list of the routes, and the
//Config parameter allows to get the file path
func GenerateSitemap(e *gin.Engine) {
	log.Print("sitemap.go > GenerateSitemap")
	//We first use the conf parameter to init the FILE and BASEURL variables
	FILE = models.Conf.SitemapBasePath + "/sitemap.xml"
	BASEURL = models.Conf.SitemapBaseUrl

	//Then we write the file, calling getSitemapStr to create and get the content
	err := ioutil.WriteFile(FILE, []byte(getSitemapStr(e)), 0755)
	if err != nil {
		log.Print(MsgStr("sitemap.error.filewrite", "en", err))
	}
}

//Function used to create the content for the file
func getSitemapStr(e *gin.Engine) string {
	log.Print("sitemap.go > getSitemapStr")
	//First of all, we add the header to specify the XML nature of this file
	txt = HEADER
	//The sitemap only needs one URLSET
	txt += URLSETOPEN
	//Then every page will generate an element in the XML
	//So, we iterate over the routes
	for _, tree := range e.Routes() {
		//Every page render function is declared in the main.main function
		//so we use that as a filter
		m, _ := regexp.MatchString("main\\.main\\.", tree.Handler)
		//Then we check the method : We only want to reference pages
		//that the users access with GET http requests, also matching our
		//function handler filter
		if tree.Method == "GET" && m {
			appendElement(tree.Path)
		}
	}
	//Then we close the URLSET tag and we return the result to be written in the file
	txt += URLSETCLOSE
	return txt
}

//Function called for each URL, which writes all necessary XML tags to the txt variable
func appendElement(page string) {
	log.Printf("sitemap.go > appendElement [ %v ]", page)
	//Getting the current time
	lastmod := time.Now()
	//Getting a pointer on the txt variable to edit it directly here
	str := &txt
	//UrlOpen tag
	*str += URLOPEN
	//Adding the path to the page in the loc tag
	*str += LOCOPEN
	*str += BASEURL + page
	*str += LOCCLOSE
	//Adding the current time in the lastmod tag
	*str += MODOPEN
	//Using yyyy-mm-dd format
	*str += lastmod.Format("2006-01-02")
	*str += MODCLOSE
	*str += URLCLOSE
}

//Sitemap renderer function to make the sitemap accessible to /sitemap.xml
func RenderSitemap() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.File(models.Conf.SitemapBasePath + "/sitemap.xml")
	}
}

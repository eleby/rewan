package models

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	//Version of the website
	Version string
	//Port used by the router
	Port int
	//Log files
	LogFile    string
	ErrLogFile string
	LogDir     string
	//SMTP connection variables
	SmtpSender   string
	SmtpPassword string
	SmtpHost     string
	SmtpChannel  string
	SmtpPort     int
	//Key used to connect to the Recaptcha service
	CaptchaKey string
	//Base url of the website
	BaseUrl string
	//Sitemap url and path
	SitemapBaseUrl  string
	SitemapBasePath string
}

//Get the configuration to use for the next actions
//and put it in a global variable to only decode the conf file once
var Conf = GetConf()

func GetConf() Config {
	log.Print("conf.go > GetConf")
	//Path to where the conf files are stored (ex: /var with user dev will fetch /var/dev.conf.json)
	path := os.Args[1]
	//Used to fetch the correct conf file (ex: dev will use dev.conf.json and prod will use prod.conf.json)
	usr := os.Args[2]
	//Value containing the configuration
	var conf Config
	//Opening and decoding the json file
	file, _ := os.Open(path + "/" + usr + ".conf.json")
	decoder := json.NewDecoder(file)
	_ = decoder.Decode(&conf)
	return conf
}

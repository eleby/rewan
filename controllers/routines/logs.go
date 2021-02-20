package routines

import (
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"os"
	"rewan/models"
	"time"
)

//Size of a megabyte in bytes
const MB = 1000000

//Limit of 10 megabytes
var MAXLOGSIZE = int64(MB * 10)

//Function creating a goroutine to manage log files
func ManageLogsRoutine() {
	log.Print("ManageLogsRoutine > Launching goroutine..")
	//The check is made every day
	tick := time.Tick(24 * time.Hour)
	//Goroutine launch
	go func() {
		for {
			select {
			//If 24 hours have passed
			case <-tick:
				//Process log files
				processLogFile(models.Conf.LogFile, false)
				processLogFile(models.Conf.ErrLogFile, true)
			//Else wait for an hour and check again
			default:
				time.Sleep(time.Second)
			}
		}
	}()
}

//Function to check the size of a log file, and manage it if it is too big
//[name] is the path of the file, [isErr] is the type of logger it is connected to
func processLogFile(name string, isErr bool) {
	//Open the file
	f, _ := os.Open(models.Conf.LogDir + name)
	//Get its file info
	s, _ := f.Stat()
	//Log the current size of the file
	log.Printf("ManageLogsRoutine > %v > Checked : size = %vMB", name, float32(s.Size())/float32(MB))
	//Set the close function to be called at the end of the process
	defer f.Close()
	//If the size is bigger than the MAXLOGSIZE variable, we need some management
	if s.Size() > MAXLOGSIZE {
		//Log the fact that we are going to copy it to an history version of it
		log.Printf("ManageLogsRoutine > %v > Copying to history file history.%v", name, name)
		//Create the history file
		f2, _ := os.Create(models.Conf.LogDir + "history." + name)
		//Defer its closing
		defer f2.Close()
		//Copy the first one to this one
		io.Copy(f2, f)
		//Delete the first one by name
		os.Remove(models.Conf.LogDir + name)
		//Re-create it as an empty file
		f, _ = os.OpenFile(models.Conf.LogDir+name, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
		//Then reassign the writers
		//Create the multiWriter
		multiWriter := io.MultiWriter(f, os.Stdout)
		//Then assign it to the correct loggers depending on the type of the log file
		if !isErr {
			//Gin's default logger
			gin.DefaultWriter = multiWriter
			//As well as a simple logger
			log.SetOutput(multiWriter)
		} else {
			//Gin's default error logger
			gin.DefaultErrorWriter = multiWriter
		}
	}
}

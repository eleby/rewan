package routines

import (
	"github.com/gin-gonic/gin"
	"log"
	"rewan/models"
	"sync"
	"time"
)

//Function creating a goroutine to manage stored variables
func IndexRemovalRoutine(array *sync.Map, index string, duration time.Duration, c *gin.Context, conf models.Config, isSecure bool) {
	log.Print("IndexRemovalRoutine > Launching goroutine..")
	//Goroutine launch
	go func() {
		time.Sleep(duration)
		_, ok := array.Load(index)
		if ok {
			array.Delete(index)
			c.SetCookie("token", "", 1, "/", conf.BaseUrl, isSecure, isSecure)
			log.Print("IndexRemovalRoutine > Deleted persisted variable")
		}
	}()
}

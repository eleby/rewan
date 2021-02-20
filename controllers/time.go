package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//Function we call from the template to get a time difference between a set date and today
func GetTimeMachine() gin.HandlerFunc {
	log.Print("time.go > GetTimeMachine")
	//We need to access the gin context to get the parameters so we use a HandlerFunc
	return func(c *gin.Context) {
		//The Query parameter will tell the function what time measurement it must return
		query := c.Param("query")
		//Then we fetch the date parameters
		years, _ := strconv.Atoi(c.Param("year"))
		months, _ := strconv.Atoi(c.Param("month"))
		days, _ := strconv.Atoi(c.Param("day"))
		hours, _ := strconv.Atoi(c.Param("hour"))
		minutes, _ := strconv.Atoi(c.Param("minute"))
		//And we return the result of the TimeMachine function
		c.String(http.StatusOK, TimeMachine(query, years, months, days, hours, minutes))
	}
}

//Function returning a time string difference between a set date and today
func TimeMachine(query string, year int, month int, day int, hour int, minute int) string {
	log.Printf("time.go > TimeMachine [ %v / %v / %v / %v / %v / %v ]", query, year, month, day, hour, minute)
	//Location to use to get the correct time
	loc, _ := time.LoadLocation("UTC")
	//Get the time from the parameters and the location
	date := time.Date(year, time.Month(month), day, hour, minute, 0, 0, loc)
	//Get the current time
	now := time.Now().In(loc).Add(time.Hour)
	//Variable containing the seconds between the two dates
	secondsBetween := float64(now.Unix() - date.Unix())
	//Declare the result variable
	var result string

	//We check the query parameter, which defines in which format we return the time
	//between the set date and now
	//Important : Since this date calculator has to be fairly accurate, we need to
	//multiply days by 365.25 instead of the classic 365 number, else remote dates
	//would give weird results,
	//for example 2021 years difference between 0001-01-01 and 2021-01-01
	switch query {
	case "years":
		//Years
		result = fmt.Sprintf("%d", int(secondsBetween/3600/24/365.25))
		break
	case "months":
		//Months
		result = fmt.Sprintf("%d", int(secondsBetween/3600/24/365.25*12))
		break
	case "days":
		//Days
		result = fmt.Sprintf("%d", int(secondsBetween/3600/24))
		break
	case "hours":
		//Hours
		result = fmt.Sprintf("%d", int(secondsBetween/3600))
		break
	case "minutes":
		//Minutes
		result = fmt.Sprintf("%d", int(secondsBetween/60))
		break
	case "seconds":
		//Seconds
		result = fmt.Sprintf("%d", int(secondsBetween))
		break
	case "decYears":
		//Years as a decimal
		result = fmt.Sprintf("%f", secondsBetween/3600/24/365.25)
		break
	default:
		result = "0"
		break
	}
	//If we have an integer number, we just return it with a space at every 4rd character
	if !strings.Contains(result, ".") {
		//So we pass it to the fmtNbStr function, which adds the spaces
		return fmtNbStr(result)
	}
	//Else we want to get a comma separator instead of a dot, and 2 characters after it
	return fmtDecStr(result, 2)
}

//Simple function to return the current year as an int
func GetCurrentYear() int {
	//Self explanatory
	now := time.Now()
	return now.Year()
}

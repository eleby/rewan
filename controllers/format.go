package controllers

import (
	"log"
	"strings"
)

//Returns a formatted version of the string,
//with a space at every 4th character
func fmtNbStr(nb string) string {
	log.Printf("format.go > fmtNbStr [ %v ]", nb)
	nbStr := ""
	nbInt := 0
	for i := len(nb); i > 0; i-- {
		nbInt++
		nbStr += nb[i-1 : i]
		if nbInt%3 == 0 {
			nbStr += " "
		}
	}
	//We then need to mirror the result, and then
	//remove an eventual space at the end
	//(= at the beginning of the mirrored result) of it
	return removeSpace(revertStr(nbStr))
}

//Returns a formatted version of the string,
//with [limit] characters after the comma which replaces the dot
func fmtDecStr(nb string, limit int) string {
	log.Printf("format.go > fmtDecStr [ %v / %v ]", nb, limit)
	return nb[:strings.Index(nb, ".")] + "," + nb[strings.Index(nb, ".")+1:strings.Index(nb, ".")+limit+1]
}

//Returns the mirrored version of the string
func revertStr(str string) string {
	log.Printf("format.go > revertStr [ %v ]", str)
	result := ""
	for i := len(str); i > 0; i-- {
		result += str[i-1 : i]
	}
	return result
}

//Returns the string without the eventual space at the beginning of it
func removeSpace(str string) string {
	log.Printf("format.go > removeSpace [ %v ]", str)
	if str[:1] == " " {
		return str[1:]
	}
	return str
}

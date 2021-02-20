package persistence

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

//Variable holding the maximum size of the tokens used to find the value in Storage
const sizeTokens = 15

//Variable holding the duration of the token and the persisted variable
const Duration = 3600 * 24 * 3

//Array holding persisted variables
//The type sync.Map avoids concurrent access
//crashing the app
var Storage = sync.Map{}

//Function used to generate a UID, used to identify the user and be able to fetch
//the correct persisted variable with it
func GenerateUID() string {
	log.Print("persist.go > GenerateUID")
	var result string
	//List of allowed characters to use for the generation
	allowedChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-"
	//Random number to choose a random index in allowedChars
	sourceRand := rand.NewSource(time.Now().UnixNano())
	randomObj := rand.New(sourceRand)
	for i := 0; i < sizeTokens; i++ {
		random := int(randomObj.Float32() * float32(len(allowedChars)))
		result += string(allowedChars[random])
	}
	return result
}

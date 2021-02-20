package memorygame

import (
	"github.com/gin-gonic/gin"
	"log"
	"math/rand"
	"net/http"
	"rewan/controllers"
	"rewan/controllers/routines"
	"rewan/models"
	"rewan/models/persistence"
	"strconv"
	"strings"
	"time"
)

//Variable holding an entire game
type MemoryGame struct {
	Time              int
	Cols              int
	Rows              int
	FontSize          float32
	NbQueensGoal      int
	NbQueens          int
	Difficulty        int
	Board             []Row
	ConvertiblePieces []Piece
	State             gameStateValue
	PiecesTried       string
	Tries             int
	Success           int
}

//Enum containing a value for each possible game state
type gameStateValue int

const (
	RUNNING = iota
	FAILED  = iota
	SUCCESS = iota
)

//An object to be able to compare values by enum name
type gameStateObj struct {
	RUNNING gameStateValue
	FAILED  gameStateValue
	SUCCESS gameStateValue
}

//Minimum number of queen pieces
const minQueensGoal = 2

//Number of cols and rows at difficulty 0
const minCols = 3
const minRows = 3

//Object containing the game state enum
var GameState = gameStateObj{RUNNING, FAILED, SUCCESS}

//Var containing all the Piece types which can be converted to queens
var convertibleTypes = []PieceTypeValue{NOTHING, PAWN, KNIGHT, BISHOP, ROOK}

//Generate and return a MemoryGame
func GenerateBoard(difficulty int) MemoryGame {
	log.Printf("memorygame.go > GenerateBoard [ %v ]", difficulty)
	//This var will contain, after an initial generation, all the
	//pieces which could be converted to queens
	var convertiblePieces []Piece
	//Difficulty variable
	diff := difficulty
	//Time before hiding the cards
	timeHide := 2000 + (diff * 1000)
	//Cols and rows
	cols := minCols + diff
	rows := minRows + diff
	//Used to resize the piece icons, else they
	//become too big with the number of cards increasing
	//and their size decreasing
	fontSize := float32(0)
	//Number of queens in the board
	queens := 0
	//The number of queens we want to have in the board
	queensGoal := int(minQueensGoal + float32(diff)*1.5)

	//A higher difficulty means a lower size for each card.
	//So we set the font size according to the difficulty
	switch difficulty {
	case 1:
		fontSize = 1.5
		break
	case 2:
		fontSize = 1.25
		break
	case 3:
		fontSize = 1.1
		break
	case 4:
		fontSize = .95
		break
	case 5:
		fontSize = .85
		break
	case 6:
		fontSize = .7
		break
	case 7:
		fontSize = .55
		break
	case 8:
		fontSize = .5
		break
	case 9:
	case 10:
		fontSize = .45
		break
	case 11:
		fontSize = .4
		break
	case 12:
		fontSize = .35
		break
	case 13:
		fontSize = .3
		break
	case 14:
		fontSize = .25
		break
	default:
		fontSize = 2
		break
	}
	//Here we get the array of rows to insert in our MemoryGame
	board, queens := getCompleteBoard(queensGoal, cols, rows)
	//Then the MemoryGame is created and returned
	return MemoryGame{timeHide, cols, rows, fontSize, int(queensGoal), queens, diff, board, convertiblePieces, RUNNING, ";", 0, 0}
}

//Returns an array of rows with all the pieces of the game, used for the Board variable of the MemoryGame
func getCompleteBoard(queensGoal int, nbCols int, nbRows int) ([]Row, int) {
	log.Printf("memorygame.go > getCompleteBoard [ %v / %v / %v ]", queensGoal, nbCols, nbRows)
	//Generate a random array of rows with random pieces in each row
	rows := getGeneratedBoard(nbCols, nbRows)
	//Each piece is added to the convertiblePieces array if its type is contained in the convertibleTypes array
	convertiblePieces := getConvertiblePieces(rows)
	//Convertible pieces are then randomly converted to Queen pieces.
	convertedPieces, newQueens := getConvertedPieces(convertiblePieces, queensGoal)
	//Then each piece of the converted array of pieces is added to its row.
	for piece := range convertedPieces {
		rowIndex := convertedPieces[piece].RowIndex
		pieceIndex := convertedPieces[piece].PieceIndex
		rows[rowIndex].Pieces[pieceIndex].PieceType = convertedPieces[piece].PieceType
	}
	return rows, newQueens
}

//Returns all the pieces, with random ones changed to queen pieces.
func getConvertedPieces(convertiblePieces []Piece, queensGoal int) ([]Piece, int) {
	log.Printf("memorygame.go > getConvertedPieces [ %v ]", queensGoal)
	//First a slice of ints is declared; it will hold the index of each piece to convert
	randomIntArray := make([]int, queensGoal)
	//Queens var to be incremented
	queens := 0
	//For each queen we need to place
	for i := 0; i < queensGoal; i++ {
		randomInt := 0
		//This loop sets a random int and will run until this number corresponds to
		//a non-queen piece which was not already set.
		for j := false; !j || convertiblePieces[randomInt].PieceType == QUEEN; {
			if !j {
				j = true
			}
			sourceRand := rand.NewSource(time.Now().UnixNano())
			randomObj := rand.New(sourceRand)
			random := randomObj.Float32() * float32(len(convertiblePieces))
			randomInt = int(random)
			for _, v := range randomIntArray {
				if v == randomInt {
					j = false
				}
			}
		}
		//This number is added to our slice for comparison purposes
		randomIntArray[i] = randomInt
		//Then we modify the concerned piece and increment the queens number var
		convertiblePieces[randomInt].PieceType = QUEEN
		queens++
	}
	//And the modified vars are returned
	return convertiblePieces, queens
}

///Function returning every piece with its type being included in the convertibleTypes array
func getConvertiblePieces(rows []Row) []Piece {
	log.Print("memorygame.go > getConvertiblePieces")
	var convertiblePieces []Piece
	//For each row
	for i := 0; i < len(rows); i++ {
		row := rows[i]
		//For each piece in the row
		for j := 0; j < len(row.Pieces); j++ {
			piece := row.Pieces[j]
			//If the type of this piece is included in the convertiblePieces array
			if contains(convertibleTypes, piece.PieceType) {
				//We add it to the convertiblePieces array
				convertiblePieces = append(convertiblePieces, piece)
			}
		}
	}
	//At the end we return all the convertiblePieces
	return convertiblePieces
}

//Generate a random array of rows with random pieces in each row
func getGeneratedBoard(nbCols int, nbRows int) []Row {
	log.Printf("memorygame.go > getGeneratedBoard [ %v / %v ]", nbCols, nbRows)
	var rows []Row
	for i := 0; i < nbCols; i++ {
		rows = append(rows, generateRow(nbRows, i))
	}
	return rows
}

//Called for a card when the user clicks on it, or for every card
//at the end of the game, this function gets the type of the
//card's piece and sets the game's state to rely on the server and avoid
//any possibility of cheating.
func GetPieceType() gin.HandlerFunc {
	log.Print("memorygame.go > GetPieceType")
	return func(c *gin.Context) {
		//Get all the parameters
		rowId, _ := strconv.Atoi(c.Param("rowId"))
		pieceId, _ := strconv.Atoi(c.Param("pieceId"))
		cookie, _ := c.Cookie("token")
		variable, _ := persistence.Storage.Load(cookie)
		//If the cookie is set and it's Storage entry exists
		if variable != nil {
			game := variable.(MemoryGame)
			piece := game.Board[rowId].Pieces[pieceId]
			//If the piece wasn't already revealed
			if !strings.Contains(game.PiecesTried, ";"+strconv.Itoa(piece.Id)+";") {
				//The number of tries is incremented, the piece is added to the list of
				//tried pieces. Then we save the modified game.
				game.PiecesTried += strconv.Itoa(piece.Id) + ";"
				game.Tries++
				persistence.Storage.Store(cookie, game)
				//If the game is ended, we return the piece type without any other operation
				if game.State != RUNNING {
					c.JSON(http.StatusOK, piece.PieceType.String())
					return
				} else {
					//If the piece is a queen we add one success point
					if piece.PieceType == PieceType.QUEEN {
						game.Success++
						//If we found all the queens we change the game state
						//to SUCCESS and we save the game object
						if game.Success == game.NbQueens {
							game.State = SUCCESS
							persistence.Storage.Store(cookie, game)
							c.JSON(http.StatusOK, GameState.SUCCESS.String())
							return
						}
						persistence.Storage.Store(cookie, game)
						//If the number of tries is superior to 1.5 * the number of queens,
						//all tries have been used. Game state is set to FAILED and we save the game.
					} else if game.Tries > int(float64(game.NbQueens)*1.5) {
						game.State = FAILED
						persistence.Storage.Store(cookie, game)
						c.JSON(http.StatusOK, GameState.FAILED.String())
						return
					}
					c.JSON(http.StatusOK, piece.PieceType.String())
					return
				}
			} else {
				c.JSON(http.StatusOK, GameState.FAILED.String())
				return
			}
		} else {
			//Else we panic a cookie not found error
			log.Panic(controllers.MsgStr("cookie.error.notfound", "en", nil))
		}
	}
}

//The two next functions are used to get enum values by their names
func (gs gameStateValue) String() string {
	return [...]string{"RUNNING", "FAILED", "SUCCESS"}[gs]
}

func (gs gameStateValue) StringLower() string {
	return [...]string{"running", "failed", "success"}[gs]
}

//Resets the game to level 0, flashes a message and redirects to the game
func Reset() gin.HandlerFunc {
	log.Print("memorygame.go > Reset")
	return func(c *gin.Context) {
		SetMemoryGame(c, 0)
		controllers.Flashvar("info", "memory.game.flash.reset", nil, c)
		c.Redirect(http.StatusFound, "/tools/memory/game")
	}
}

//Finishes the game and sets the difficulty accordingly; also flashes a message
func Finish() gin.HandlerFunc {
	log.Print("memorygame.go > Finish")
	return func(c *gin.Context) {
		game := GetMemoryGame(c)
		var flashType string
		var flashMessage string
		if game.State == SUCCESS {
			game.Difficulty++
			flashType = "success"
			flashMessage = "memory.game.flash.end.success"
		} else {
			flashType = "error"
			flashMessage = "memory.game.flash.end.failed"
		}
		controllers.Flashvar(flashType, flashMessage, game.Difficulty+1, c)
		SetMemoryGame(c, game.Difficulty)
		c.Redirect(http.StatusFound, "/tools/memory/game")
	}
}

//Check if the array contains the value
func contains(i []PieceTypeValue, j PieceTypeValue) bool {
	for _, v := range i {
		if v == j {
			return true
		}
	}

	return false
}

//Sets and returns a memory game
func GetSetMemoryGame(c *gin.Context) MemoryGame {
	log.Print("memorygame.go > GetSetMemoryGame")
	isSecure := gin.Mode() == gin.ReleaseMode
	uid, err := c.Cookie("token")
	variable, _ := persistence.Storage.Load(uid)
	var game MemoryGame
	//If the game already exists in Storage
	if variable != nil {
		game = variable.(MemoryGame)
		//Create the memory game with the same difficulty
		game = GenerateBoard(game.Difficulty)
		//Store it with the token as the key
		persistence.Storage.Store(uid, game)
	} else if err == nil {
		//Else if the cookie is set we store a new game in Storage
		game = GenerateBoard(0)
		//Store it with the token as the key
		persistence.Storage.Store(uid, game)
		//The goroutine removes the Storage map entry
		routines.IndexRemovalRoutine(&persistence.Storage, uid, time.Second*persistence.Duration, c, models.Conf, isSecure)
	}
	//If the cookie is not set, we set it and the Storage value
	if err != nil {
		uid = persistence.GenerateUID()
		c.SetCookie("token", uid, persistence.Duration, "/", models.Conf.BaseUrl, isSecure, isSecure)
		//Init the variable passed to the template
		game = GenerateBoard(0)
		//Store it with the token as the key
		persistence.Storage.Store(uid, game)
		//Launch a goroutine to delete the stored variable
		routines.IndexRemovalRoutine(&persistence.Storage, uid, time.Second*persistence.Duration, c, models.Conf, isSecure)
	}
	return game
}

//Gets the current memory game; if it does not exist it creates a new one at difficulty 1
func GetMemoryGame(c *gin.Context) MemoryGame {
	log.Print("memorygame.go > GetMemoryGame")
	uid, err := c.Cookie("token")
	variable, _ := persistence.Storage.Load(uid)
	var game MemoryGame
	if variable != nil {
		game = variable.(MemoryGame)
	} else if err == nil {
		game = GenerateBoard(0)
		persistence.Storage.Store(uid, game)
		isSecure := gin.Mode() == gin.ReleaseMode
		routines.IndexRemovalRoutine(&persistence.Storage, uid, time.Second*persistence.Duration, c, models.Conf, isSecure)
	}
	return game
}

//Creates and stores a memorygame with the indicated difficulty
func SetMemoryGame(c *gin.Context, difficulty int) {
	log.Print("memorygame.go > SetMemoryGame")
	uid, err := c.Cookie("token")
	if err != nil {
		uid = persistence.GenerateUID()
		isSecure := gin.Mode() == gin.ReleaseMode
		c.SetCookie("token", uid, persistence.Duration, "/", models.Conf.BaseUrl, isSecure, isSecure)
		routines.IndexRemovalRoutine(&persistence.Storage, uid, time.Second*persistence.Duration, c, models.Conf, isSecure)
	}
	persistence.Storage.Store(uid, GenerateBoard(difficulty))
}

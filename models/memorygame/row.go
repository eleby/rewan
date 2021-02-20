package memorygame

import (
	"log"
	"math/rand"
	"time"
)

//The Piece array contained in the row
type Row struct {
	Pieces []Piece
}

//Function used to generate a row of the Memorygame board.
//It uses a number of columns and the index of the row.
func generateRow(cols int, index int) Row {
	log.Print("row.go > generateRow")
	//A row is essentially an array of pieces
	var pieces []Piece
	//For each column we generate one piece
	for i := 0; i < cols; i++ {
		//Then we generate a random int, the number of
		//Piece types *2
		sourceRand := rand.NewSource(time.Now().UnixNano())
		randomObj := rand.New(sourceRand)
		random := randomObj.Float32() * countTypes * 2
		//If it is > than the number of Pieces types (1 on 2 chances)
		//the piece will be an empty spot to free space on the board
		//Else if will take the type associated with the random int
		var pieceType PieceTypeValue
		if random > countTypes {
			pieceType = 0
		} else {
			pieceType = PieceTypeValue(random)
		}
		//Then we add the piece to the array
		pieces = append(pieces, generatePiece(pieceType, index, i, index*cols+i))
	}
	//Return the result
	return Row{pieces}
}

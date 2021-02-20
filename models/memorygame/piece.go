package memorygame

import (
	"log"
)

//A piece has an unique id in its row and in the board,
//it also stored the index of its row and its index in the row,
//and of course its type
type Piece struct {
	Id         int
	RowIndex   int
	PieceIndex int
	PieceType  PieceTypeValue
}

//All piece types, defined with iota which is an iteration keyword
type PieceTypeValue int

const (
	NOTHING = iota
	PAWN    = iota
	KNIGHT  = iota
	BISHOP  = iota
	ROOK    = iota
	QUEEN   = iota
)

//Used to compare values with types by type name
//Mainly used in templates
type pieceTypeObj struct {
	NOTHING PieceTypeValue
	PAWN    PieceTypeValue
	KNIGHT  PieceTypeValue
	BISHOP  PieceTypeValue
	ROOK    PieceTypeValue
	QUEEN   PieceTypeValue
}

//Number of non-queen types = value of the highest type
const countTypes = QUEEN

//The PieceType object we pass to templates
//to compare piece types by their names
var PieceType = pieceTypeObj{NOTHING, PAWN, KNIGHT, BISHOP, ROOK, QUEEN}

//Creates a piece with the passed parameters
func generatePiece(pieceType PieceTypeValue, rowIndex int, pieceIndex int, index int) Piece {
	log.Printf("piece.go > generatePiece [ %v / %v / %v / %v ]", pieceType, rowIndex, pieceIndex, index)
	piece := Piece{index + 1, rowIndex, pieceIndex, pieceType}
	return piece
}

//Functions used to get the name of the types as strings
//Uppercase
func (pt PieceTypeValue) String() string {
	return [...]string{"NOTHING", "PAWN", "KNIGHT", "BISHOP", "ROOK", "QUEEN"}[pt]
}

//Lowercase
func (pt PieceTypeValue) StringLower() string {
	return [...]string{"nothing", "pawn", "knight", "bishop", "rook", "queen"}[pt]
}

//If the piece is in the corner of the board,
//this method will return the right class to add to it
//to set the correct border-radius
func GetCornerClass(p Piece, cols int, rows int) string {
	log.Printf("piece.go > GetCornerClass [ %v / %v / %v ]", p.Id, cols, rows)
	if p.PieceIndex == 0 && p.RowIndex == 0 {
		return "chessTopLeft"
	} else if p.RowIndex == 0 && p.PieceIndex == cols-1 {
		return "chessTopRight"
	} else if p.RowIndex == rows-1 && p.PieceIndex == 0 {
		return "chessBotLeft"
	} else if p.RowIndex == rows-1 && p.PieceIndex == cols-1 {
		return "chessBotRight"
	}
	return ""
}

//Every second piece must have a different color than the last one
//Since the number of columns can be odd as well as even, this function
//also needs the index of the row
func GetColorPiece(p Piece, cols int) string {
	log.Printf("piece.go > GetColorPiece [ %v / %v ]", p.Id, cols)
	if cols%2 == 0 {
		if (p.Id%2 == 0 && p.RowIndex%2 == 0) || (p.Id%2 != 0 && p.RowIndex%2 != 0) {
			return "chess-white"
		} else {
			return "chess-black"
		}
	} else {
		if p.Id%2 == 1 {
			return "chess-white"
		} else {
			return "chess-black"
		}
	}
}

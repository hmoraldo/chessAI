package main

import "fmt"

const PieceStatusBits = 3
const BitsPerSquare = PieceStatusBits + 2

/*

Every board square is numbered this way:

 0  1  2  3  4  5  6  7
 8  9 10 11 12 13 14 15
16 17 ...
...

A uint64 in Board contains one bit for each of the 64 squares,
in that order.

The first PieceStatusBits bits for each square represent the current square status.
This status uses the values in PieceStatus.

The remaining bits are used for:

- one bit for storing the PieceStatus value (can the king do a castling move? etc.)
- one bit for storing the PieceColor value

*/

type Board [BitsPerSquare]uint64;

type Piece uint8

const (
	Piece_Empty Piece = iota
	Piece_Pawn
	Piece_Rock
	Piece_Knight
	Piece_Bishop
	Piece_King
	Piece_Queen
)

var pieceNamesMap = map[Piece]string {
	Piece_Empty : "Empty", Piece_Pawn : "Pawn", Piece_Rock : "Rock", Piece_Bishop : "Bishop", Piece_Knight : "Knight", Piece_Queen : "Queen", Piece_King : "King",
}

func (p Piece) String() string {
	return pieceNamesMap[p]
}

type PieceStatus bool

const (
	PieceStatus_Default PieceStatus = false // initial status: pawn can't be captured in en-passant, rock / king can do castling
	PieceStatus_EnPassantAllowed = true // pawn can be captured using en-passant move
	PieceStatus_CastlingNotAllowed = true // rock or king not allowed to do castling
)

type PieceColor bool

const (
	PieceColor_White PieceColor = true
	PieceColor_Black = false
)

type PieceInfo struct {
	piece Piece
	status PieceStatus
	color PieceColor
}

var EmptyPieceInfo = PieceInfo{ Piece_Empty, PieceStatus_Default, PieceColor_White }

func (p PieceColor) String() string {
	if p == PieceColor_White { return "White" }
	return "Black"
}

type SquareColor bool

const (
	SquareColor_White SquareColor = true
	SquareColor_Black = false
)

// we assume background is white, otherwise the colors will look reverted
var pieceCharMap = map[PieceColor]map[Piece]string {
	PieceColor_White :
		{ Piece_Empty : ` `, Piece_Pawn : `♙`, Piece_Rock : `♖`, Piece_Knight : `♘`, Piece_Bishop : `♗`, Piece_King : `♔`, Piece_Queen : `♕`, },
	PieceColor_Black :
		{ Piece_Empty : ` `, Piece_Pawn : `♟`, Piece_Rock : `♜`, Piece_Knight : `♞`, Piece_Bishop : `♝`, Piece_King : `♚`, Piece_Queen : `♛`, },
}

var squareCharMap  = map[SquareColor]string { SquareColor_White : ` `, SquareColor_Black : `▨`, }

func BoolToInt(b bool) uint64 {
	if b {
		return 1
	}
	return 0
}

func GetBitValue(bits uint64, pos uint64) uint64 {
	return (bits >> pos) & 1
}

func SetBitValue(bits, pos, value uint64) uint64 {
	return (bits &^ (1 << pos)) | (value << pos)
}

// GetBoardAt gives information about a piece in a given position
func GetBoardAt(board Board, pos Position) (info PieceInfo) {
	var value, bitidx uint64

	if !PositionInBoard(pos) { panic("wrong position") }

	bitidx = uint64(pos.x + pos.y * 8)

	// value[bit0] = board[0][bit pos], value[bit1] = board[1][bit pos], ...
	for i := uint64(0); i < PieceStatusBits; i ++ {
		value |= GetBitValue(board[i], bitidx) << i
	}

	info.piece = Piece(value)
	info.status = 1 == GetBitValue(board[PieceStatusBits], bitidx)
	info.color = 1 == GetBitValue(board[PieceStatusBits + 1], bitidx)

	return
}

// SetBoardAt modifies a specific position of a board
func SetBoardAt(board *Board, pos Position, info PieceInfo) {

	if !PositionInBoard(pos) { panic("wrong position") }

	bitidx := uint64(pos.x + pos.y * 8)

	for i := uint64(0); i < PieceStatusBits; i ++ {
		(*board)[i] = SetBitValue((*board)[i], bitidx, GetBitValue(uint64(info.piece), i))
	}

	(*board)[PieceStatusBits] = SetBitValue((*board)[PieceStatusBits], bitidx, BoolToInt(bool(info.status)))
	(*board)[PieceStatusBits + 1] = SetBitValue((*board)[PieceStatusBits + 1], bitidx, BoolToInt(bool(info.color)))
}

func PositionInBoard(pos Position) bool {
	if pos.x < 0 || pos.x > 7 || pos.y < 0 || pos.y > 7 { return false }
	return true
}

func DrawPiece(info PieceInfo, square SquareColor) {
	printSquares := true
	debugStatus := false

	fmt.Print(" ")
	if debugStatus {
		if info.piece == Piece_Pawn && info.status == PieceStatus_EnPassantAllowed {
				fmt.Print("P")
				return
		}
		if info.piece == Piece_Rock && info.status == PieceStatus_CastlingNotAllowed {
				fmt.Print("R")
				return
		}
		if info.piece == Piece_King && info.status == PieceStatus_CastlingNotAllowed {
				fmt.Print("K")
				return
		}
	}

	if info.piece == Piece_Empty {
		if printSquares {
			fmt.Print(squareCharMap[square])
		} else {
			fmt.Print(" ")
		}
	} else {
		fmt.Print(pieceCharMap[info.color][info.piece])
	}
}

func DrawBoard(board Board) {
	squareColor := SquareColor_White
	lineCount := 0

	fmt.Println("  0 1 2 3 4 5 6 7")

	for y := 0; y < 8; y ++ {
		fmt.Print(lineCount)
		
		for x := 0; x < 8; x ++ {
			info := GetBoardAt(board, Position{x, y})
			DrawPiece(info, squareColor)
			squareColor = !squareColor
		}
		squareColor = !squareColor
		lineCount ++
		fmt.Println("")
	}
}

func GetPieces(board Board, piece Piece, color PieceColor) []Position {
	posl := make([]Position, 0, 4)

	for x := 0; x < 8; x ++ {
		for y := 0; y < 8; y ++ {
			pos := Position{x, y}
			infoHere := GetBoardAt(board, pos)
			if piece == infoHere.piece && color == infoHere.color {
				posl = append(posl, pos)
			}
		}
	}

	return posl
}

func GetPiecesByColor(board Board, color PieceColor) []Position {
	posl := make([]Position, 0, 4)

	for x := 0; x < 8; x ++ {
		for y := 0; y < 8; y ++ {
			pos := Position{x, y}
			infoHere := GetBoardAt(board, pos)
			if color == infoHere.color && infoHere.piece != Piece_Empty {
				posl = append(posl, pos)
			}
		}
	}

	return posl
}

func fillInitialBoardSide(board Board, piecesRow, pawnsRow int, color PieceColor, testBoard bool) Board {
	for i := 0; i < 8; i ++ {
		SetBoardAt(&board, Position{i, pawnsRow}, PieceInfo{ Piece_Pawn, PieceStatus_Default, color })
	}

	SetBoardAt(&board, Position{0, piecesRow}, PieceInfo{ Piece_Rock, PieceStatus_Default, color })
	SetBoardAt(&board, Position{7, piecesRow}, PieceInfo{ Piece_Rock, PieceStatus_Default, color })

	SetBoardAt(&board, Position{4, piecesRow}, PieceInfo{ Piece_King, PieceStatus_Default, color })

	if !testBoard {
		SetBoardAt(&board, Position{1, piecesRow}, PieceInfo{ Piece_Knight, PieceStatus_Default, color })
		SetBoardAt(&board, Position{6, piecesRow}, PieceInfo{ Piece_Knight, PieceStatus_Default, color })

		SetBoardAt(&board, Position{2, piecesRow}, PieceInfo{ Piece_Bishop, PieceStatus_Default, color })
		SetBoardAt(&board, Position{5, piecesRow}, PieceInfo{ Piece_Bishop, PieceStatus_Default, color })

		SetBoardAt(&board, Position{3, piecesRow}, PieceInfo{ Piece_Queen, PieceStatus_Default, color })
	}

	return board
}

func InitialBoard(testBoard bool) Board {
	var board Board

	board = fillInitialBoardSide(board, 0, 1, PieceColor_Black, testBoard)
	board = fillInitialBoardSide(board, 7, 6, PieceColor_White, testBoard)

	return board
}


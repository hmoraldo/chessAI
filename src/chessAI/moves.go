package main

import "fmt"
import "math"

type Position struct {
	x, y int
}

type Move struct {
	x, y int // move relative to current position
}

type FullMove struct {
	pos Position
	move Move
}

// the result of applying a move
type Turn struct {
	board Board
	lastMove FullMove
}

// MoveSeq contains a list of moves such that, moves[n] is valid only if moves[n - 1] is valid as well.
// This allows us to tell easily that you can't, for example, move a rock two places away if you can't do it one time
// away in the same direction.
type MoveSeq []Move

// MoveSeqs contains a list of MoveSeq; each MoveSeq is independent from the others
type MoveSeqs []MoveSeq

// movesMap stores the relative movement for each piece
var movesMap = map[PieceColor]map[Piece]MoveSeqs {}

func PositionAdd(pos Position, move Move) Position {
	return Position{ pos.x + move.x, pos.y + move.y }
}

func PositionDiff(pos Position, move Move) Position {
	return Position{ pos.x - move.x, pos.y - move.y }
}

// ApplyCastling applies the castling move in one specific direction; it assumes castling is valid
func ApplyCastling(board Board, kingPos Position, kingInfo PieceInfo, direction int) (newBoard Board) {
	var rockMove, kingMove FullMove
	newBoard = board

	if direction < 0 {
		rockMove = FullMove{ Position{0, kingPos.y}, Move{3, 0} }
	} else {
		rockMove = FullMove{ Position{7, kingPos.y}, Move{-2, 0} }
	}

	SetBoardAt(&newBoard, rockMove.pos, PieceInfo{ Piece_Rock, PieceStatus_CastlingNotAllowed, kingInfo.color })
	SetBoardAt(&newBoard, kingPos, PieceInfo{ kingInfo.piece, PieceStatus_CastlingNotAllowed, kingInfo.color })

	updateStates := true
	kingMove = FullMove{ kingPos, Move{direction * 2, 0} }
	newBoard = ApplyMove(newBoard, rockMove, updateStates)
	newBoard = ApplyMove(newBoard, kingMove, updateStates)

	return
}

// ApplyEnPassant applies en-passant move; it assumes the move is valid
func ApplyEnPassant(board Board, fullMove FullMove, updateStates bool) Board {
	newPos := PositionAdd(fullMove.pos, fullMove.move)

	board = ApplyMove(board, fullMove, updateStates)
	SetBoardAt(&board, Position{ newPos.x, fullMove.pos.y }, EmptyPieceInfo)
	return board
}

// ApplyPawnPromotion applies promotion move for one selected promotion type; it assumes the move is valid
func ApplyPawnPromotion(board Board, fullMove FullMove, selectedPiece Piece, updateStates bool) Board {
	info := GetBoardAt(board, fullMove.pos)
	newPos := PositionAdd(fullMove.pos, fullMove.move)

	board = ApplyMove(board, fullMove, updateStates)

	status := PieceStatus_Default
	if selectedPiece == Piece_Rock { status = PieceStatus_CastlingNotAllowed }
	
	SetBoardAt(&board, newPos, PieceInfo{ selectedPiece, status, info.color })

	return board
}

// addCastlingMove computes the board for a left or right castling move for the given king.
// direction is either -1 (left) or 1 (right)
func addCastlingMove(board Board, kingPos Position, kingInfo PieceInfo, direction int) (newBoard Board, ok bool) {
	var rockPos Position

	rockPos = Position{ 0, kingPos.y }
	if direction == 1 { rockPos.x = 7 }

	rockInfo := GetBoardAt(board, rockPos)
	if rockInfo.piece != Piece_Rock || rockInfo.status != PieceStatus_Default || rockInfo.color != kingInfo.color { return }

	// all squares between king and rock must be empty
	for xi := kingPos.x + direction; xi != rockPos.x; xi += direction {
		newPos := Position{ xi, kingPos.y }
		newInfo := GetBoardAt(board, newPos)

		if newInfo.piece != Piece_Empty { return }
	}

	// neither the king square nor the two squares in the direction of the rock can be under attack
	for xi := kingPos.x; xi != kingPos.x + 3 * direction; xi += direction {
		newPos := Position{ xi, kingPos.y }
		if isUnderAttack(board, newPos, kingInfo.color) { return }
	}


	// apply move to king & rock
	ok = true
	newBoard = ApplyCastling(board, kingPos, kingInfo, direction)
	return
}

func addCastlingMoves(board Board, kingPos Position, kingInfo PieceInfo, moves []Board) (newMoves []Board) {

	newMoves = moves
	if kingInfo.status != PieceStatus_Default { return }

	dirs := []int { -1, 1 }

	for _, dir := range dirs {
		move, ok := addCastlingMove(board, kingPos, kingInfo, dir)
		if ok { newMoves = append(newMoves, move) }
	}

	return
}

// addPawnMove adds either the pawn move, or all available promotions if the move is a promotion
func addPawnMove(board Board, info PieceInfo, move FullMove, isEnPassant bool, moves []Board) (newMoves []Board) {

	newMoves = moves
	updateStates := true
	newPos := PositionAdd(move.pos, move.move)
	
	if newPos.y != 0 && newPos.y != 7 {
		var newMove Board
		
		if isEnPassant {
			newMove = ApplyEnPassant(board, move, updateStates)
		} else {
			newMove = ApplyMove(board, move, updateStates)
		}
		newMoves = append(newMoves, newMove)
		return
	}

	availablePromotions := []Piece{ Piece_Queen, Piece_Rock, Piece_Bishop, Piece_Knight }
	for _, newPiece := range availablePromotions {
		status := PieceStatus_Default
		if newPiece == Piece_Rock { status = PieceStatus_CastlingNotAllowed }

		newBoard := ApplyPawnPromotion(board, move, newPiece, updateStates)
		SetBoardAt(&newBoard, newPos, PieceInfo{ newPiece, status, info.color })
		newMoves = append(newMoves, newBoard)
	}
	
	return
}

// addPawnSpecialMoves adds to list, the captures that can be done by a given pawn (including en-passant) and
// the promotion
func addPawnSpecialMoves(board Board, pos Position, info PieceInfo, moves []Board) (newMoves []Board) {

	newMoves = moves

	yDirection := 1
	if info.color == PieceColor_White { yDirection = -1 }
	newy := pos.y + yDirection
	if newy < 0 || newy > 7 { return }

	xDirections := []int{ -1, 1 }

	for _, xDirection := range xDirections {
		newx := pos.x + xDirection
		fullMove := FullMove{ pos, Move{ xDirection, yDirection } }

		if newx < 0 || newx > 7 { continue }
		enemyInfo := GetBoardAt(board, Position{ newx, newy })

		isEnPassant := false

		if enemyInfo.piece != Piece_Empty {	
			// normal capture
			if enemyInfo.color != info.color {
				newMoves = addPawnMove(board, info, fullMove, isEnPassant, newMoves)
			}
		} else {
			// try en-passant
			isEnPassant = true
			
			enPassantPos := Position{ newx, pos.y }
			enPassantInfo := GetBoardAt(board, enPassantPos)
			if enPassantInfo.color != info.color && enPassantInfo.piece == Piece_Pawn && enPassantInfo.status == PieceStatus_EnPassantAllowed {
				tmpMoves := []Board{}
				tmpMoves = addPawnMove(board, info, fullMove, isEnPassant, tmpMoves)
				newMoves = append(newMoves, tmpMoves...)
			}
		}
	}
	
	return
}

// removeCheckMoves gets rid of any moves that put the king under attack
func removeCheckMoves(boards []Board, color PieceColor) []Board {
	newBoards := make([]Board, 0, len(boards))

	for _, b := range boards {
		if len(GetPieces(b, Piece_King, color)) == 0 {
			// TODO: remove this, only here to debug
			fmt.Println("DEBUG BOARD!!")
			DrawBoard(b)
		}
		
		kingPos := GetPieces(b, Piece_King, color)[0]
		if !isUnderAttack(b, kingPos, color) {
			newBoards = append(newBoards, b)
		}
	}
	
	return newBoards
}


// GetPossibleMoves returns the list of moves that can be done by a single piece.
// It doesn't take checks into account, except for castling.
// Params:
// - filterCheckMoves = true forces the removal of any moves that puts the king under attack.
// - quickMode = true skips some steps that aren't necessary for secondary uses of this
//   function: computing castling and updating state info.
func GetPossibleMoves(board Board, pos Position, info PieceInfo, filterCheckMoves bool, quickMode bool) []Board {
	seqs := movesMap[info.color][info.piece]

	moves := []Move{}

	for _, seq := range seqs {
		for _, move := range seq {
			newPos := PositionAdd(pos, move)
			if !PositionInBoard(newPos) { break }

			infoHere := GetBoardAt(board, newPos)
			if infoHere.piece == Piece_Empty {
				moves = append(moves, move)
			} else {
				if infoHere.color != info.color && info.piece != Piece_Pawn {
					moves = append(moves, move)
				}
				break
			}
		}
	}


	boards := []Board{}
	updateStates := !quickMode
	for _, m := range moves {
		boards = append(boards, ApplyMove(board, FullMove{pos, m}, updateStates))
	}

	// we assume first move is one step, second move is two steps... this is always correct because
	// of the MoveSeq definition
	pawnWithMoves := info.piece == Piece_Pawn && len(moves) != 0
	if pawnWithMoves && ((info.color == PieceColor_Black && pos.y != 1) || (info.color == PieceColor_White && pos.y != 6)) {
		moves = moves[:1]
	}
	if info.piece == Piece_Pawn {
		if len(moves) == 1 {
			isEnPassant := false
			boards = addPawnMove(board, info, FullMove{ pos, moves[0] }, isEnPassant, []Board{})
		}
		boards = addPawnSpecialMoves(board, pos, info, boards)
	}


	if !quickMode && info.piece == Piece_King {
		boards = addCastlingMoves(board, pos, info, boards)
	}

	if filterCheckMoves {
		boards = removeCheckMoves(boards, info.color)
	}

	return boards
}

// GetAllPossibleMoves returns all possible moves for pieces of a given color
// (more details about arguments in GetPossibleMoves)
func GetAllPossibleMoves(board Board, color PieceColor, filterCheckMoves bool, quickMode bool) []Board {
	positions := GetPiecesByColor(board, color)
	allMoves := []Board{}

	for _, pos := range positions {
		info := GetBoardAt(board, pos)
		allMoves = append(allMoves, GetPossibleMoves(board, pos, info, filterCheckMoves, quickMode)...)
	}
	
	return allMoves
}

// isUnderAttack tells whether a piece with color=color is under attack by any enemy piece.
// This is the slow, but easy implementation.
func isUnderAttack(board Board, pos Position, color PieceColor) bool {

	var enemies []Position = GetPiecesByColor(board, !color)
	filterCheckMoves := false
	quickMode := true

	for _, ePos := range enemies {
		enemyInfo := GetBoardAt(board, ePos)
		enemyMoves := GetPossibleMoves(board, ePos, enemyInfo, filterCheckMoves, quickMode)

		for _, enemyMove := range enemyMoves {
			infoHere := GetBoardAt(enemyMove, pos)
			if infoHere.piece != Piece_Empty && infoHere.color != color { return true }
		}
	}

	return false
}

func IsValidMove(board Board, piecePos Position, newBoard Board) bool {

	quickMode := false
	filterCheckMoves := true
	info := GetBoardAt(board, piecePos)
	moves := GetPossibleMoves(board, piecePos, info, filterCheckMoves, quickMode)

	for _, m := range moves {
		if m == newBoard { return true }
	}

	return false
}

func GetPossibleMoveCount(board Board, color PieceColor, filterCheckMoves bool) int {
	count := 0
	quickMode := true
	
	for _, pos := range GetPiecesByColor(board, color) {
		info := GetBoardAt(board, pos)
		count += len(GetPossibleMoves(board, pos, info, filterCheckMoves, quickMode))
	}
	
	return count
}

// resetPawnsStatus resets the status of all pawns of a given color; this means no contrary pawn can capture
// any pawn using en-passant after this
func resetPawnsStatus(board Board, color PieceColor) Board {
	for _, pos := range GetPieces(board, Piece_Pawn, color) {
		info := GetBoardAt(board, pos)
		SetBoardAt(&board, pos, PieceInfo{ info.piece, PieceStatus_Default, info.color })
	}

	return board
}

// ApplyMove executes a move in a board; it assumes the move is a valid one, and it only applies simple moves
// (castling or en-passant can't use this function)
// Because the resetPawnsStatus call that updates states for the en-passant capture can be slow, we allow that
// to be disabled.
func ApplyMove(board Board, fullMove FullMove, updateStates bool) Board {
	info := GetBoardAt(board, fullMove.pos)

	// switch state changes for castling & en-passant
	if updateStates {
		board = resetPawnsStatus(board, info.color)
		if info.piece == Piece_King || info.piece == Piece_Rock {
			info.status = PieceStatus_CastlingNotAllowed
		}
		if info.piece == Piece_Pawn && math.Abs(float64(fullMove.move.y)) == 2. {
			info.status = PieceStatus_EnPassantAllowed
		}
		if info.piece == Piece_Pawn && math.Abs(float64(fullMove.move.y)) == 1. {
			info.status = PieceStatus_Default
		}
	}

	SetBoardAt(&board, fullMove.pos, EmptyPieceInfo)
	SetBoardAt(&board, PositionAdd(fullMove.pos, fullMove.move), info)
	return board
}

func initMovesMap(color PieceColor) map[Piece]MoveSeqs {
	m := make(map[Piece]MoveSeqs)

	// each sequence has to be in an order such that move n can only be done if
	// move n-1 is also possible (this takes care of collisions)

	// pawn
	dir := 1
	if color == PieceColor_White { dir = -1 }
	m[Piece_Pawn] = MoveSeqs{ MoveSeq{ Move{0, dir}, Move{0, 2 * dir} } }

	// rock
	rmoves := MoveSeqs{ MoveSeq{}, MoveSeq{}, MoveSeq{}, MoveSeq{} }
	for i := 1; i < 8; i ++ {
		rmoves[0] = append(rmoves[0], Move{0, i})
		rmoves[1] = append(rmoves[1], Move{0, -i})
		rmoves[2] = append(rmoves[2], Move{i, 0})
		rmoves[3] = append(rmoves[3], Move{-i, 0})
	}
	m[Piece_Rock] = MoveSeqs{ MoveSeq(rmoves[0]), MoveSeq(rmoves[1]), MoveSeq(rmoves[2]), MoveSeq(rmoves[3]) }

	// bishop
	bmoves := MoveSeqs{ MoveSeq{}, MoveSeq{}, MoveSeq{}, MoveSeq{} }
	for i := 1; i < 8; i ++ {
		bmoves[0] = append(bmoves[0], Move{i, i})
		bmoves[1] = append(bmoves[1], Move{-i, -i})
		bmoves[2] = append(bmoves[2], Move{i, -i})
		bmoves[3] = append(bmoves[3], Move{-i, i})
	}
	m[Piece_Bishop] = MoveSeqs{ MoveSeq(bmoves[0]), MoveSeq(bmoves[1]), MoveSeq(bmoves[2]), MoveSeq(bmoves[3]) }

	// queen
	qmoves := MoveSeqs{}
	qmoves = append(qmoves, rmoves...)
	qmoves = append(qmoves, bmoves...)
	m[Piece_Queen] = qmoves

	// king
	kmoves := MoveSeqs{}
	for i := 0; i < len(qmoves); i ++ {
		kmoves = append(kmoves, MoveSeq{ qmoves[i][0] })
	}
	m[Piece_King] = kmoves

	// knight
	m[Piece_Knight] = MoveSeqs{
		MoveSeq{ Move{-2, -1} }, MoveSeq{ Move{-1, -2} },
		MoveSeq{ Move{2, 1} }, MoveSeq{ Move{1, 2} },
		MoveSeq{ Move{-2, 1} }, MoveSeq{ Move{-1, 2} },
		MoveSeq{ Move{2, -1} }, MoveSeq{ Move{1, -2} },
	 }

	return m
}

// isCheckMate tells whether the king of a color is in checkmate.
func isCheckMate(board Board, availableMoveCount int, color PieceColor) bool {
	kingPos := GetPieces(board, Piece_King, color)[0]
	if availableMoveCount == 0 && isUnderAttack(board, kingPos, color) { return true }
	return false
}

// GetGameStatus tells whether game is finished or not, and who wins if it is finished
func GetGameStatus(board Board, nextTurnColor PieceColor, availableMoveCount int) (finished bool, draw bool, winningColor PieceColor) {
	finished = true

	if isCheckMate(board, availableMoveCount, nextTurnColor) {
		winningColor = !nextTurnColor
		return
	}

	if availableMoveCount == 0 {
		draw = true
		return
	}
	
	finished = false
	return
}

func init() {
	movesMap[PieceColor_Black] = initMovesMap(PieceColor_Black)
	movesMap[PieceColor_White] = initMovesMap(PieceColor_White)
}


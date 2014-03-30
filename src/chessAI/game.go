package main

import "fmt"
import "math"
import "time"

func ComputerTurn(board Board, color PieceColor) (finalBoard Board, canMove bool) {

	filterCheckMoves := true
	if GetPossibleMoveCount(board, color, filterCheckMoves) == 0 { return }

	negamaxDepth := 3
	bestMove, bestScore := Negamax(board, color, negamaxDepth)
	
	fmt.Println("Best score found", bestScore)
	return bestMove, true
}

func sign(x int) int {
	if x > 0 { return 1 }
	return -1
}

// PlayerTurn asks the player for a move, and applies it
func PlayerTurn(board Board, color PieceColor) Board {
	var fullMove FullMove
	var newBoard Board
	valid := false

	for {
		fmt.Println("Insert your move: x y diffx diffy")
		fmt.Scanln(&fullMove.pos.x, &fullMove.pos.y, &fullMove.move.x, &fullMove.move.y)

		info := GetBoardAt(board, fullMove.pos)
		if !PositionInBoard(fullMove.pos) {
			fmt.Println("Must select square inside of board")
			continue			
		}
		if info.piece == Piece_Empty {
			fmt.Println("Can't select empty piece")
			continue
		}
		if info.color != color {
			fmt.Println("Wrong piece color!")
			continue
		}
		newPos := PositionAdd(fullMove.pos, fullMove.move)
		if !PositionInBoard(newPos) {
			fmt.Println("Can't make move outside of the board!")
			continue			
		}
		
		isCastling := info.piece == Piece_King && math.Abs(float64(fullMove.move.x)) > 1.
		isPawnPromotion := info.piece == Piece_Pawn &&
			((info.color == PieceColor_White && newPos.y == 0)  || (info.color == PieceColor_Black && newPos.y == 7))
		isPawnCapture := info.piece == Piece_Pawn && math.Abs(float64(fullMove.move.x)) == 1.
		updateStates := true
		
		if isCastling {
			newBoard = ApplyCastling(board, fullMove.pos, info, sign(fullMove.move.x))
		} else if isPawnPromotion {
			var selectedPieceCode int
			fmt.Println("Select piece to promote to: 0 is queen, 1 is knight, 2 is bishop, 3 is rock")
			fmt.Scanln(&selectedPieceCode)

			promotionCodes := map[int]Piece { 0 : Piece_Queen, 1 : Piece_Knight, 2 : Piece_Bishop, 3 : Piece_Rock }
			selectedPiece, ok := promotionCodes[selectedPieceCode]
			
			if !ok {
				fmt.Println("Can't promote to selected piece")
				continue
			}
			
			newBoard = ApplyPawnPromotion(board, fullMove, selectedPiece, updateStates)
		} else if isPawnCapture {
			capturedInfo := GetBoardAt(board, newPos)
			enPassantInfo := GetBoardAt(board, Position{ newPos.x, fullMove.pos.y })
			
			if capturedInfo.color != color && capturedInfo.piece != Piece_Empty {
				newBoard = ApplyMove(board, fullMove, updateStates)
			} else if capturedInfo.piece == Piece_Empty && enPassantInfo.color != color && enPassantInfo.piece == Piece_Pawn {
				newBoard = ApplyEnPassant(board, fullMove, updateStates)
			} else {
				fmt.Println("Invalid pawn move")
				continue
			}
		} else {
			newBoard = ApplyMove(board, fullMove, updateStates)
		}

		valid = IsValidMove(board, fullMove.pos, newBoard)
		if valid {
			return newBoard
		}
		fmt.Println("Invalid move!")
	}
}

func DrawTurn(board Board, color PieceColor) {
	fmt.Println("Color", color, "turn:")
	DrawBoard(board)
	fmt.Println("===========================")
}

func gameEnded(board Board, colorNextTurn PieceColor) bool {
	filterCheckMoves := true
	availableMoveCount := GetPossibleMoveCount(board, colorNextTurn, filterCheckMoves)
	finished, draw, winningColor := GetGameStatus(board, colorNextTurn, availableMoveCount)
	
	if finished && draw {
		fmt.Println("Game over, result: draw")
	}
	if finished && !draw {
		fmt.Println("Game over, result:", winningColor, "wins")
	}
	
	return finished
}

// players can be 0 (computer - computer), 1 (computer - player) or 2 (computer - computer)
func PlayGame(players int) {
	useTestBoard := false
	board := InitialBoard(useTestBoard)
	color := PieceColor_White
	turnCount := 0

	DrawTurn(board, color)

	for {
		var ok bool

		fmt.Println("Turn:", turnCount)

		if players < 2 {
			t := time.Now()
			
			board, ok = ComputerTurn(board, color)
			
			fmt.Println("Time spent by computer", time.Since(t))
			
			if !ok { break }
			DrawTurn(board, color)
			color = !color
		}
		if gameEnded(board, color) { return }

		if players > 0 {
			board = PlayerTurn(board, color)
			DrawTurn(board, color)
			color = !color
		}
		if gameEnded(board, color) { return }

		turnCount ++
	}
}

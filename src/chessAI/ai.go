package main

import "math"

type BoardScore struct {
	board Board
	score int
}

var pieceScoreMap = map[Piece]int {
	Piece_King : 1000, Piece_Queen : 9, Piece_Knight : 3, Piece_Bishop : 3, Piece_Rock : 5, Piece_Pawn : 1,
}

func getPiecesScore(board Board, color PieceColor) int {
	var info PieceInfo
	positions := GetPiecesByColor(board, color)
	
	score := 0
	for _, pos := range positions {
		info = GetBoardAt(board, pos)
		score += pieceScoreMap[info.piece]
	}
	
	return score
}

func EvaluateBoard(board Board, color PieceColor) int {
	
	drawScore := - 100
	checkMateScore := 1000
	
	filterCheckMoves := true

	moveCount := GetPossibleMoveCount(board, color, filterCheckMoves)
	enemyMoveCount := GetPossibleMoveCount(board, !color, filterCheckMoves)
	moveScore := moveCount - enemyMoveCount

	pieceScore := getPiecesScore(board, color)
	enemyPieceScore := getPiecesScore(board, !color)
	combinedPieceScore := pieceScore - enemyPieceScore

	finished, draw, winningColor := GetGameStatus(board, color, moveCount)
	if finished {
		if draw { return drawScore }
		if winningColor == color {
			return checkMateScore
		} else {
			return - checkMateScore
		}
	}
	
	return moveScore + combinedPieceScore * 2
}

var biggestScore = 100000
var lowestScore = - biggestScore

func NegamaxAlphaBeta(board Board, color PieceColor, alpha, beta int, transpositionTable map[Board]int, maxDepth int) (bestMove Board, bestScore int) {


	if maxDepth == 0 {
		bestMove = board
		bestScore = EvaluateBoard(board, color)
		return
	}
	
	filterCheckMoves := true
	quickMode := false
	moves := GetAllPossibleMoves(board, color, filterCheckMoves, quickMode)

	var score int
	bestScore = lowestScore
	for _, move := range moves {
		
		cached, ok := transpositionTable[move]
		if ok {
			score = cached
		} else {
			_, score = NegamaxAlphaBeta(move, !color, -beta, -alpha, transpositionTable, maxDepth - 1)
			transpositionTable[move] = score
		}
		
		score = - score
		if score > bestScore {
			bestScore = score
			bestMove = move
		}
		
		alpha = int(math.Max(float64(alpha), float64(score)))
		if alpha > beta { break }
	}
	
	return
}

func Negamax(board Board, color PieceColor, maxDepth int) (bestMove Board, bestScore int) {

	var transpositionTable = make(map[Board]int)

	alpha := lowestScore
	beta := biggestScore
	if color == PieceColor_Black {
		alpha, beta = -beta, -alpha
	}
	bestMove, bestScore = NegamaxAlphaBeta(board, color, alpha, beta, transpositionTable, maxDepth)
	return
}



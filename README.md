chessAI
=======

Chess AI is a chess player written in the Go language. It also contains the chess game to allow playing against the AI.

The chess program currently supports:

- 0, 1 and 2 player modes: computer against computer, player against computer, player against player
- move search is based in Negamax (a zero sum version of Minimax) with Alpha-Beta pruning and transposition tables

I am also currently working on:

- a parallel version of Negamax
- sorting moves for making Negamax AB much faster
- other misc. changes

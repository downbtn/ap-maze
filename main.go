package main

import "github.com/downbtn/ap-maze/maze"

var AVAILABLE_MAZES = []string{"maze_1", "maze_2"}

func main() {
	game := maze.CreateGame(AVAILABLE_MAZES)
	game.MainMenu()
}

package maze

import (
	"math/rand"
)

type Direction uint8

const POS_Y Direction = 0
const POS_X Direction = 1
const NEG_Y Direction = 2
const NEG_X Direction = 3

// GenerateMaze uses a depth-first approach to generate a maze.
// The parameters width and height are NOT the dimensions of the resulting map,
// but rather the dimensions of the maze grid that generates them. The
// dimension of the generated maze will always be 2n+1.
func GenerateMaze(width int, height int, seed int64) (*Maze, error) {

	// Start by creating a 2w+1 x 2h+1 board of all walls.
	// This is to have the cells separated by walls at the end.

	board := make([][]Tile, 0, (2*height + 1))
	for i := 0; i < (2*height + 1); i++ {
		board = append(board, make([]Tile, (2*width+1), (2*width+1)))
		for j, _ := range board[i] {
			board[i][j] = TILE_WALL
		}
	}

	// The caller needs to supply a seed to use the builtin PRNG. If the
	// user doesn't input one, just read 8 bytes from /dev/urandom or
	// equivalent.
	rng := rand.New(rand.NewSource(seed))

	toVisit := width * height
	x := rng.Intn(width)
	y := rng.Intn(height)
	backtrack := make([]Coords, 0, toVisit)
	endpoints := make([]Coords, 1)
	endpoints = append(endpoints, Coords{X: x, Y: y})

	for toVisit > 0 {
		// Randomly traverse board and mark path until a square with no
		// unmarked neighbors is reached.

		// check directions
		var directions []Direction
		if y != height-1 && board[1+2*(y+1)][1+2*x] != TILE_EMPTY {
			directions = append(directions, POS_Y)
		}
		if y != 0 && board[1+2*(y-1)][1+2*x] != TILE_EMPTY {
			directions = append(directions, NEG_Y)
		}
		if x != width-1 && board[1+2*y][1+2*(x+1)] != TILE_EMPTY {
			directions = append(directions, POS_X)
		}
		if x != 0 && board[1+2*y][1+2*(x-1)] != TILE_EMPTY {
			directions = append(directions, NEG_X)
		}

		if len(directions) == 0 {
			// this is a dead end
			endpoints = append(endpoints, Coords{X: x, Y: y})
			// backtrack
			for len(directions) == 0 {
				x = backtrack[len(backtrack)-1].X
				y = backtrack[len(backtrack)-1].Y
				backtrack = backtrack[:len(backtrack)-1]

				if y != height-1 && board[1+2*(y+1)][1+2*x] != TILE_EMPTY {
					directions = append(directions, POS_Y)
				}
				if y != 0 && board[1+2*(y-1)][1+2*x] != TILE_EMPTY {
					directions = append(directions, NEG_Y)
				}
				if x != width-1 && board[1+2*y][1+2*(x+1)] != TILE_EMPTY {
					directions = append(directions, POS_X)
				}
				if x != 0 && board[1+2*y][1+2*(x-1)] != TILE_EMPTY {
					directions = append(directions, NEG_X)
				}
			}
		} else {
			move := directions[rand.Intn(len(directions))]
			switch move {
			case POS_X:
				board[2*y+1][2*x+2] = TILE_EMPTY
				x++
			case POS_Y:
				board[2*y+2][2*x+1] = TILE_EMPTY
				y++
			case NEG_X:
				board[2*y+1][2*x] = TILE_EMPTY
				x--
			case NEG_Y:
				board[2*y][2*x+1] = TILE_EMPTY
				y--
			}
			toVisit--
			board[1+2*y][1+2*x] = TILE_EMPTY
			backtrack = append(backtrack, Coords{X: x, Y: y})
		}

	}

	// Place down the entrance and exit
	// We don't want them to be too close together, but "closeness" in a
	// maze is really dictated by the distance of the shortest path between
	// two points and not the actual distance. So, I need a way to find the
	// two points with the longest "shortest possible path" between them.

	// Because of the way we generate a maze, there *should* be no "loops"
	// because the algorithm will refuse to visit a space it's already
	// visited. Therefore, if we pick the start of the generation as a
	// starting node, then the longest possible shortest path is either
	// from this point or between two of the ends of the branches.

	// We can use a greedy algorithm.
	var src Coords
	var dest Coords
	dist := -1

	tmpMaze := &Maze{Board: board}
	for _, p1 := range endpoints {
		spt, err := tmpMaze.CreateSpt(Coords{p1.X*2 + 1, p1.Y*2 + 1})
		if err != nil {
			return nil, err
		}

		longest := -1
		var p2 Coords
		for j, line := range spt {
			for k, val := range line {
				if val > longest {
					longest = val
					p2 = Coords{X: k, Y: j}
				}
			}
		}

		if longest > dist {
			dist = longest
			src = p1
			dest = p2
		}

	}

	board[src.Y*2+1][src.X*2+1] = TILE_START
	board[dest.Y*2+1][dest.X*2+1] = TILE_END

	return &Maze{
		Board:   board,
		Start:   Coords{X: src.X*2 + 1, Y: src.Y*2 + 1},
		End:     Coords{X: dest.X*2 + 1, Y: dest.Y*2 + 1},
		PathLen: dist,
		Width:   width*2 + 1,
		Height:  height*2 + 1,
	}, nil
}

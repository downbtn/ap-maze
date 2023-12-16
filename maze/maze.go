package maze

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type Tile rune

const TILE_EMPTY Tile = '.'
const TILE_WALL Tile = '#'
const TILE_START Tile = '>'
const TILE_END Tile = '<'

type Coords struct {
	X int
	Y int
}

type Maze struct {
	Board   [][]Tile
	Start   Coords
	End     Coords
	PathLen int
	Width   int
	Height  int
}

func LoadMazeFromString(s string) (*Maze, error) {
	lines := strings.Split(s, "\n")

	var board [][]Tile
	var startX int
	var startY int
	var endX int
	var endY int

	starts := 0
	ends := 0
	width := -1
	for i, l := range lines {
		row := []Tile(l)

		if len(row) == 0 {
			continue
		} else if width == -1 {
			width = len(row)
		} else if width != len(row) {
			return nil, fmt.Errorf("All rows in a maze must have the same length. Expected width: %d Got width: %d", width, len(row))
		}

		for j, tile := range row {
			if tile == TILE_START {
				if starts > 0 {
					return nil, errors.New("Maze cannot have multiple start points")
				}
				startX = j
				startY = i
				starts++
			} else if tile == TILE_END {
				if ends > 0 {
					return nil, errors.New("Maze cannot have multiple end points")
				}
				endX = j
				endY = i
				ends++
			} else if rune(tile) == ' ' {
				row[j] = TILE_EMPTY
			} else if tile != TILE_EMPTY && tile != TILE_WALL {
				return nil, fmt.Errorf("Invalid maze tile: %c", tile)
			}
		}
		board = append(board, row)
	}

	return &Maze{
		Start:   Coords{X: startX, Y: startY},
		End:     Coords{X: endX, Y: endY},
		Board:   board,
		PathLen: -1,
		Height:  len(board),
		Width:   width,
	}, nil
}

func LoadMazeFromFile(filename string) (*Maze, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return LoadMazeFromString(string(content))
}

func (m *Maze) DisplayText(playerX int, playerY int) (string, error) {
	var sb strings.Builder
	for i, row := range m.Board {
		for j, tile := range row {
			if j == playerX && i == playerY {
				sb.WriteRune('@')
			} else {
				sb.WriteRune(rune(tile))
			}
		}
		sb.WriteRune('\n')
	}

	return sb.String(), nil
}

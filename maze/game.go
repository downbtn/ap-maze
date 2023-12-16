package maze

import (
	"errors"
	"fmt"
	"math"
	"runtime"
	"strings"
	"time"

	tcell "github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Score represents the result of playing the maze. The user can succeed or
// fail and if they succeed they take a certain number of steps. It's used to
// make other threads wait for a game to finish.
type Score struct {
	Score int
	Won   bool
	Map   string
}

func CalcScore(steps int, bestSteps int) float64 {
	diff := float64(steps - bestSteps)
	coef := (1 - math.Exp(-diff/15)) / (1 + math.Exp(-diff/15))
	return 1000000 * (1 - coef)
}

func CalcScoreEndless(steps int, bestSteps int, round int) float64 {
	multiplier := 1 + math.Pow(float64(round), 2)/32
	score := multiplier * CalcScore(steps, bestSteps)
	return score
}

// Game represents the running state of a game, both the board state and
// also the TUI state.
type Game struct {
	Application    *tview.Application
	Pages          *tview.Pages
	AvailMaps      []string
	CurrentMap     *Maze
	CurrentMapName string
	CurrentSteps   int
	Endless        bool
	EndlessRounds  int
	PlayerX        int
	PlayerY        int
	//ScoreChannel   chan *Score
}

// CreateGame creates a Game struct. You need to populate the data yourself
func CreateGame(levels []string) *Game {
	return &Game{
		Application:    tview.NewApplication(),
		Pages:          tview.NewPages(),
		CurrentMap:     nil,
		CurrentMapName: "none",
		AvailMaps:      levels,
		PlayerX:        -1,
		PlayerY:        -1,
	}
}

func (g *Game) LevelSelect() {
	if g.Pages.HasPage("map_select") {
		g.Pages.SwitchToPage("map_select")
	} else {
		selectModal := tview.NewModal().SetText("Which map would you like to play?").AddButtons(g.AvailMaps).AddButtons([]string{"Exit"})
		selectModal.SetDoneFunc(func(_ int, label string) {
			if label == "Exit" {
				g.Application.Stop()
				return
			}
			g.LoadFile(label)
			g.PlayMap()
		})
		g.Pages.AddAndSwitchToPage("map_select", selectModal, false)
	}

}

// MainMenu opens the main menu, allowing the user to choose between playing
// Endless and Level modes, viewing highscores, and exiting.
func (g *Game) MainMenu() {
	if g.Pages.HasPage("menu") {
		g.Pages.SwitchToPage("menu")
	} else {
		menu := tview.NewModal().SetText("The Labyrinth\n\nA simple roguelike maze game made by Daniel Ha")
		//menu = menu.AddButtons([]string{"Levels", "Endless", "Credits"})
		menu = menu.AddButtons([]string{"Levels", "Credits"}) // Endless doesn't work right now
		menu.SetDoneFunc(func(_ int, btn string) {
			switch btn {
			case "Credits":
				g.displayCopyright()
			case "Levels":
				g.LevelSelect()
			case "Endless":
				g.PlayEndless()
			}
		})

		g.Pages.AddAndSwitchToPage("menu", menu, true)
	}

	g.Application = g.Application.SetRoot(g.Pages, true)
	g.Application.Run()
}

func (g *Game) okModal(content string, temp_id string) {
	oldPageId, _ := g.Pages.GetFrontPage()

	modal := tview.NewModal().SetText(content).AddButtons([]string{"OK"})
	modal.SetDoneFunc(func(_ int, _ string) {
		g.Pages.RemovePage(temp_id)
		g.Pages.SwitchToPage(oldPageId)
	})

	g.Pages.AddAndSwitchToPage(temp_id, modal, false)

}

// DisplayError is used for displaying an error to the user in a modal.
// I think this is a nicer way of handling errors that won't just crash the
// game when some invalid data is encountered.
func (g *Game) DisplayError(err error) {
	_, file, line, ok := runtime.Caller(1)
	var errorText string
	if ok {
		errorText = fmt.Sprintf("error %s:%d\n%v", file, line, err)
	} else {
		errorText = fmt.Sprintf("unknown error\n%v", err)
	}

	g.okModal(errorText, "error")
}

func (g *Game) PauseMenu() {
	menu := tview.NewModal().SetText("GAME PAUSED\nWhat would you like to do?").AddButtons([]string{"Quit to menu", "Copyright", "Help"})
	menu.SetDoneFunc(func(_ int, label string) {
		switch label {
		case "Quit to menu":
			g.ClearGame()
			g.MainMenu()
		case "Help":
			help := `Welcome to my maze game!
Controls: arrow keys to move, ESC to open menu
Tiles: @ is your player. You start on >. Your goal is
to make it to the >. # is a wall, you can't run into walls.`
			g.okModal(help, "help")
		default:
			g.DisplayError(errors.New("Invalid option"))
		}

		g.Pages.RemovePage("menu")
	})

	g.Pages.AddAndSwitchToPage("menu", menu, true)

}

func (g *Game) ClearGame() {
	if g.CurrentMapName == "none" {
		// game is not running
		return
	}

	g.CurrentMapName = "none"
	g.CurrentMap = nil
	g.CurrentSteps = 0
	g.Endless = false
	g.EndlessRounds = 0
	g.Pages.RemovePage("game")
}

func (g *Game) LoadFile(mapId string) {
	// Load map and store pointer in the Game struct
	currentMap, err := LoadMazeFromFile("data/" + mapId)
	if err != nil {
		g.DisplayError(err)
		return
	}
	g.LoadMaze(currentMap, mapId)
}

func (g *Game) LoadMaze(m *Maze, name string) {
	g.CurrentMap = m
	g.PlayerX = g.CurrentMap.Start.X
	g.PlayerY = g.CurrentMap.Start.Y
	g.CurrentMapName = name
	g.CurrentSteps = 0
}

func (g *Game) EndGame(s *Score) {
	endScreen := tview.NewModal()
	if g.Endless {
		endScreen = endScreen.AddButtons([]string{"Continue"})
	}
	if s.Won {
		text := fmt.Sprintf(`STAGE CLEAR: %s
Congratulations!
Your score was: %d`, s.Map, s.Score)
		endScreen = endScreen.SetText(text).AddButtons([]string{"Main Menu"})
	} else {
		text := fmt.Sprintf("STAGE FAILED: %s", s.Map)
		endScreen = endScreen.SetText(text).AddButtons([]string{"Retry", "Main Menu"})
	}

	endScreen = endScreen.SetDoneFunc(func(_ int, id string) {
		switch id {
		case "Main Menu":
			g.ClearGame()
			g.MainMenu()
		case "Retry":
			g.LoadMaze(g.CurrentMap, g.CurrentMapName)
			g.PlayMap()
		case "Continue":
			return
		}
	})
	g.Pages.AddAndSwitchToPage("end", endScreen, true)
}

// PlayMap loads a map and runs the game on that map.
func (g *Game) PlayMap() {
	gameBox := tview.NewTextView().SetText("Press any key to begin...")
	gameBox.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		failed := false
		won := false
		switch event.Key() {
		case tcell.KeyEscape:
			g.PauseMenu()
			return nil
		case tcell.KeyUp:
			if g.PlayerY == 0 || g.CurrentMap.Board[g.PlayerY-1][g.PlayerX] == TILE_WALL {
				failed = true
			} else {
				g.PlayerY--
				g.CurrentSteps++
				if g.CurrentMap.Board[g.PlayerY][g.PlayerX] == TILE_END {
					won = true
				}
			}
		case tcell.KeyDown:
			if g.PlayerY == g.CurrentMap.Height-1 || g.CurrentMap.Board[g.PlayerY+1][g.PlayerX] == TILE_WALL {
				failed = true
			} else {
				g.PlayerY++
				g.CurrentSteps++
				if g.CurrentMap.Board[g.PlayerY][g.PlayerX] == TILE_END {
					won = true
				}
			}
		case tcell.KeyLeft:
			if g.PlayerX == 0 || g.CurrentMap.Board[g.PlayerY][g.PlayerX-1] == TILE_WALL {
				failed = true
			} else {
				g.PlayerX--
				g.CurrentSteps++
				if g.CurrentMap.Board[g.PlayerY][g.PlayerX] == TILE_END {
					won = true
				}
			}
		case tcell.KeyRight:
			if g.PlayerX == g.CurrentMap.Width-1 || g.CurrentMap.Board[g.PlayerY][g.PlayerX+1] == TILE_WALL {
				failed = true
			} else {
				g.PlayerX++
				g.CurrentSteps++
				if g.CurrentMap.Board[g.PlayerY][g.PlayerX] == TILE_END {
					won = true
				}
			}
		}

		display, err := g.CurrentMap.DisplayText(g.PlayerX, g.PlayerY)
		if err != nil {
			g.DisplayError(err)
			return nil
		}

		var update strings.Builder
		if failed {
			update.WriteString("Can't move there\n\n")
		} else if won {
			var score float64
			if g.Endless {
				score = CalcScoreEndless(g.CurrentSteps, g.CurrentMap.PathLen, g.EndlessRounds)
			} else {
				score = CalcScore(g.CurrentSteps, g.CurrentMap.PathLen)
			}

			scorePtr := &Score{
				Score: int(score),
				Won:   true,
				Map:   g.CurrentMapName,
			}
			//g.ScoreChannel <- scorePtr
			g.EndGame(scorePtr)

		} else {
			update.WriteString("\n\n")
		}

		update.WriteString(display)
		gameBox.SetText(update.String())
		return nil
	})

	g.Pages.AddAndSwitchToPage("game", gameBox, true)

	//result := <-g.ScoreChannel
	//g.EndGame(result)
}

// Endless mode keeps randomly generating mazes with more and more difficulty
// each time. You need to reach the exit within a certin amount of moves each
// time and your score is based on how many stages you can clear.
func (g *Game) PlayEndless() {
	g.Endless = true
	difficulty := 1

	for {
		// get dimensions based on difficulty
		width := 5 + difficulty
		height := width * 4 / 5
		m, err := GenerateMaze(width, height, time.Now().UnixNano())
		if err != nil {
			g.DisplayError(err)
			continue
		}
		g.LoadMaze(m, "Endless")
		// TODO: the function below doesn't block so it leads to an infinite loop
		// Endless mode will NOT WORK until it's fixed
		g.PlayMap()
		difficulty++
	}
	g.Endless = false
}

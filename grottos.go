package main

import (
	"github.com/gcatlin/gocurses"
)

func main() {
	g := Game{Width: 80, Height: 24}
	g.Init()
	defer g.Shutdown()
	g.Run()
}

type Window struct {
	*curses.Window
}

type Game struct {
	Window        Window
	Width, Height int
	Screen        Screen
	Quit          bool
}

func (g *Game) Init() {
	w := Window{curses.Initscr()}
	w.Clear()
	curses.Cbreak()
	curses.Noecho()
	curses.Cursor(0)
	w.Keypad(true)
	g.Window = w

	g.MainMenu()
}

func (g *Game) Run() {
	for !g.Quit {
		g.Window.Clear()
		g.Render()
		keyCode := g.WaitForInput()
		g.HandleInput(keyCode)
	}
}

func (g *Game) ExitGame() {
	g.Quit = true
}

func (g *Game) Shutdown() {
	curses.End()
	g.Window.Clear()
}

func (g *Game) Render() {
	g.Screen.Render(g)
}

func (g *Game) WaitForInput() KeyCode {
	return KeyCode(g.Window.Getch())
}

func (g *Game) HandleInput(kc KeyCode) {
	g.Screen.HandleInput(kc)
}

func (g *Game) MainMenu() {
	ms := MenuScreen{
		Title: "Grottos of Go",
		Items: []MenuItem{
			MenuItem{"New Game", func() { g.PlayGame() }},
			MenuItem{"Quit", func() { g.ExitGame() }},
		},
	}
	ms.KeyBindings = NewKeyBindingMap([]KeyBinding{
		KeyBinding{'k', func() { ms.PrevItem() }},
		KeyBinding{'j', func() { ms.NextItem() }},
		KeyBinding{10, func() { ms.ExecuteItem() }},
	})
	g.Screen = &ms
}

func (g *Game) PlayGame() {
	ps := PlayScreen{Game: g}

	ps.Map = NewGameMap(
		[][]int{
			[]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			[]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			[]int{0, 0, 1, 1, 1, 1, 1, 0, 0, 0, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0},
			[]int{0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0},
			[]int{0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0},
			[]int{0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0},
			[]int{0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0},
			[]int{0, 0, 1, 1, 2, 1, 1, 0, 0, 0, 1, 1, 2, 1, 1, 0, 0, 0, 0, 0},
			[]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			[]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			[]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			[]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			[]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			[]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			[]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}})

	ps.KeyBindings = NewKeyBindingMap([]KeyBinding{
		KeyBinding{'y', func() { ps.MovePlayerNorthWest() }},
		KeyBinding{'u', func() { ps.MovePlayerNorthEast() }},
		KeyBinding{'h', func() { ps.MovePlayerWest() }},
		KeyBinding{'j', func() { ps.MovePlayerSouth() }},
		KeyBinding{'k', func() { ps.MovePlayerNorth() }},
		KeyBinding{'l', func() { ps.MovePlayerEast() }},
		KeyBinding{'b', func() { ps.MovePlayerSouthWest() }},
		KeyBinding{'n', func() { ps.MovePlayerSouthEast() }},
		KeyBinding{'q', func() { ps.ExitToMainMenu() }},
	})
	g.Screen = &ps
}

type KeyCode int
type Command func()
type KeyBinding struct {
	KeyCode KeyCode
	Command Command
}

type KeyBindingMap struct {
	Bindings map[KeyCode]Command
}

func NewKeyBindingMap(bindings []KeyBinding) KeyBindingMap {
	km := KeyBindingMap{}
	km.Bindings = make(map[KeyCode]Command)
	for _, kb := range bindings {
		km.Bind(kb.KeyCode, kb.Command)
	}
	return km
}

func (km *KeyBindingMap) Lookup(kc KeyCode) Command {
	if cmd, hasKey := km.Bindings[kc]; hasKey {
		return cmd
	}
	return func() {} // no-op
}

func (km *KeyBindingMap) Bind(kc KeyCode, cmd Command) {
	km.Bindings[kc] = cmd
}

func (km *KeyBindingMap) Unbind(kc KeyCode) {
	delete(km.Bindings, kc)
}

type Screen interface {
	Render(*Game)
	HandleInput(KeyCode)
}

type MenuScreen struct {
	Screen
	Title       string
	Items       []MenuItem
	CurrentItem int
	KeyBindings KeyBindingMap
}

func (ms *MenuScreen) Render(g *Game) {
	g.Window.Clear()

	g.Window.Attron(curses.A_BOLD)
	g.Window.Mvaddstr(0, 0, ms.Title)
	g.Window.Attroff(curses.A_BOLD)

	offset := 2
	for i, item := range ms.Items {
		indicator := "  "
		if i == ms.CurrentItem {
			indicator = "> "
		}
		g.Window.Mvaddstr(i+offset, 0, indicator+item.Name)
	}
}

func (ms *MenuScreen) HandleInput(kc KeyCode) {
	ms.KeyBindings.Lookup(kc)()
}

func (ms *MenuScreen) PrevItem() {
	ms.SelectItem(ms.CurrentItem - 1)
}

func (ms *MenuScreen) NextItem() {
	ms.SelectItem(ms.CurrentItem + 1)
}

func (ms *MenuScreen) SelectItem(i int) {
	// wrap around
	max := len(ms.Items) - 1
	if i > max {
		i = 0
	} else if i < 0 {
		i = max
	}
	ms.CurrentItem = i
}

func (ms *MenuScreen) ExecuteItem() {
	ms.Items[ms.CurrentItem].Command()
}

type MenuItem struct {
	Name    string
	Command func()
}

type PlayScreen struct {
	Screen
	Game        *Game
	Map         GameMap
	KeyBindings KeyBindingMap
	Player      Player
	Abort       bool
}

func (ps *PlayScreen) Render(g *Game) {
	for y := 0; y < ps.Map.Height; y++ {
		g.Window.Move(y, 0)
		for x := 0; x < ps.Map.Width; x++ {
			c := int('.')
			switch ps.Map.Grid[y][x] {
			case 1:
				c = int('#')
			case 2:
				c = int('+')
			}
			g.Window.Addch(c)
		}
	}

	g.Window.Mvaddch(ps.Player.Y, ps.Player.X, '@')
}

func (ps *PlayScreen) HandleInput(kc KeyCode) {
	ps.KeyBindings.Lookup(kc)()

	if ps.Abort {
		ps.Game.MainMenu()
	}

	if ps.Player.X < 0 {
		ps.Player.X = 0
	} else if ps.Player.X >= ps.Map.Width {
		ps.Player.X = ps.Map.Width - 1
	}

	if ps.Player.Y < 0 {
		ps.Player.Y = 0
	} else if ps.Player.Y >= ps.Map.Height {
		ps.Player.Y = ps.Map.Height - 1
	}
}

func (ps *PlayScreen) MovePlayerNorthWest() {
	ps.Player.X--
	ps.Player.Y--
}

func (ps *PlayScreen) MovePlayerNorth() {
	ps.Player.Y--
}

func (ps *PlayScreen) MovePlayerNorthEast() {
	ps.Player.X++
	ps.Player.Y--
}

func (ps *PlayScreen) MovePlayerEast() {
	ps.Player.X++
}

func (ps *PlayScreen) MovePlayerSouthEast() {
	ps.Player.X++
	ps.Player.Y++
}

func (ps *PlayScreen) MovePlayerSouth() {
	ps.Player.Y++
}

func (ps *PlayScreen) MovePlayerSouthWest() {
	ps.Player.X--
	ps.Player.Y++
}

func (ps *PlayScreen) MovePlayerWest() {
	ps.Player.X--
}

func (ps *PlayScreen) ExitToMainMenu() {
	ps.Abort = true
}

type Player struct {
	X, Y int
}

type GameMap struct {
	Grid          [][]int
	Width, Height int
}

func NewGameMap(grid [][]int) GameMap {
	gm := GameMap{Grid: grid}
	gm.Height = len(grid)
	gm.Width = len(grid[0])
	return gm
}

package main

import (
	"github.com/gcatlin/gocurses"
	"log"
	"math/rand"
	"os"
)

func main() {
	f, err := os.OpenFile("log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Could not open log file: %v", err)
	}
	defer f.Close()

	g := Game{
		Width:  80,
		Height: 24,
		Logger: log.New(f, "", log.LstdFlags|log.Lshortfile|log.Lmicroseconds),
	}
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
	*log.Logger
}

func (g *Game) Init() {
	g.Logger.Println("Initializing")
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
	g.Logger.Println("Shutting down")
	curses.End()
	g.Window.Clear()
}

func (g *Game) Render() {
	g.Screen.Render(g)
}

func (g *Game) WaitForInput() KeyCode {
	return KeyCode(g.Window.Getch())
}

func (g *Game) HandleInput(key KeyCode) {
	g.Screen.HandleInput(key)
}

func (g *Game) MainMenu() {
	s := MenuScreen{
		Title: "Grottos of Go",
		Items: []MenuItem{
			MenuItem{"New Game", func() { g.PlayGame() }},
			MenuItem{"Quit", func() { g.ExitGame() }},
		},
	}
	s.KeyBindings = NewKeyBindingMap([]KeyBinding{
		KeyBinding{'n', func() { g.PlayGame() }},
		KeyBinding{'q', func() { g.ExitGame() }},
		KeyBinding{'k', func() { s.PrevItem() }},
		KeyBinding{'j', func() { s.NextItem() }},
		KeyBinding{10, func() { s.ExecuteItem() }},
	})
	g.Screen = &s
}

func (g *Game) PlayGame() {
	s := PlayScreen{Game: g}
	s.Map = NewRandomGameMap(g.Width, g.Height)
	s.KeyBindings = NewKeyBindingMap([]KeyBinding{
		KeyBinding{'y', func() { s.MovePlayerNorthWest() }},
		KeyBinding{'u', func() { s.MovePlayerNorthEast() }},
		KeyBinding{'h', func() { s.MovePlayerWest() }},
		KeyBinding{'j', func() { s.MovePlayerSouth() }},
		KeyBinding{'k', func() { s.MovePlayerNorth() }},
		KeyBinding{'l', func() { s.MovePlayerEast() }},
		KeyBinding{'b', func() { s.MovePlayerSouthWest() }},
		KeyBinding{'n', func() { s.MovePlayerSouthEast() }},
		KeyBinding{'q', func() { g.MainMenu() }},
		KeyBinding{10, func() { g.WinGame() }},
		KeyBinding{27, func() { g.LoseGame() }},
	})
	g.Screen = &s
}

func (g *Game) WinGame() {
	g.Screen = NewEndScreen(g, "You win!!!")
}

func (g *Game) LoseGame() {
	g.Screen = NewEndScreen(g, "You lose!!!")
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
	if cmd, ok := km.Bindings[kc]; ok {
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
	HandleInput(KeyCode)
	Render(*Game)
}

type MenuScreen struct {
	Screen
	Title       string
	Items       []MenuItem
	CurrentItem int
	KeyBindings KeyBindingMap
}

func (s *MenuScreen) HandleInput(kc KeyCode) {
	s.KeyBindings.Lookup(kc)()
}

func (s *MenuScreen) Render(g *Game) {
	g.Window.Clear()

	g.Window.Attron(curses.A_BOLD)
	g.Window.Mvaddstr(0, 0, s.Title)
	g.Window.Attroff(curses.A_BOLD)

	offset := 2
	for i, item := range s.Items {
		indicator := "  "
		if i == s.CurrentItem {
			indicator = "> "
		}
		g.Window.Mvaddstr(i+offset, 0, indicator+item.Name)
	}
}

func (s *MenuScreen) PrevItem() {
	s.SelectItem(s.CurrentItem - 1)
}

func (s *MenuScreen) NextItem() {
	s.SelectItem(s.CurrentItem + 1)
}

func (s *MenuScreen) SelectItem(i int) {
	// wrap around
	max := len(s.Items) - 1
	if i > max {
		i = 0
	} else if i < 0 {
		i = max
	}
	s.CurrentItem = i
}

func (s *MenuScreen) ExecuteItem() {
	s.Items[s.CurrentItem].Command()
}

type MenuItem struct {
	Name    string
	Command func()
}

type EndScreen struct {
	Screen
	Game        *Game
	Message     string
	KeyBindings KeyBindingMap
}

func (s *EndScreen) HandleInput(kc KeyCode) {
	s.KeyBindings.Lookup(kc)()
}

func NewEndScreen(g *Game, msg string) *EndScreen {
	s := EndScreen{Game: g, Message: msg}
	s.KeyBindings = NewKeyBindingMap([]KeyBinding{
		KeyBinding{10, func() { g.MainMenu() }},
	})
	return &s
}

func (s *EndScreen) Render(g *Game) {
	g.Window.Clear()
	g.Window.Mvaddstr(0, 0, s.Message)
}

type PlayScreen struct {
	Screen
	Game        *Game
	Map         *GameMap
	KeyBindings KeyBindingMap
	Player      Player
}

func (s *PlayScreen) Render(g *Game) {
	for y := 0; y < s.Map.Height; y++ {
		g.Window.Move(y, 0)
		for x := 0; x < s.Map.Width; x++ {
			c := s.Map.Tiles[y][x]
			g.Window.Addch(c)
		}
	}

	g.Window.Mvaddch(s.Player.Y, s.Player.X, '@')
}

func (s *PlayScreen) HandleInput(kc KeyCode) {
	s.KeyBindings.Lookup(kc)()

	if s.Player.X < 0 {
		s.Player.X = 0
	} else if s.Player.X >= s.Map.Width {
		s.Player.X = s.Map.Width - 1
	}

	if s.Player.Y < 0 {
		s.Player.Y = 0
	} else if s.Player.Y >= s.Map.Height {
		s.Player.Y = s.Map.Height - 1
	}
}

func (s *PlayScreen) MovePlayerNorthWest() {
	s.Player.X--
	s.Player.Y--
}

func (s *PlayScreen) MovePlayerNorth() {
	s.Player.Y--
}

func (s *PlayScreen) MovePlayerNorthEast() {
	s.Player.X++
	s.Player.Y--
}

func (s *PlayScreen) MovePlayerEast() {
	s.Player.X++
}

func (s *PlayScreen) MovePlayerSouthEast() {
	s.Player.X++
	s.Player.Y++
}

func (s *PlayScreen) MovePlayerSouth() {
	s.Player.Y++
}

func (s *PlayScreen) MovePlayerSouthWest() {
	s.Player.X--
	s.Player.Y++
}

func (s *PlayScreen) MovePlayerWest() {
	s.Player.X--
}

type Player struct {
	X, Y int
}

type GameMap struct {
	Tiles         [][]int
	Width, Height int
}

func (m *GameMap) GetTile(x, y int) (c int, ok bool) {
	ymax := len(m.Tiles) - 1
	if 0 <= y && y <= ymax && ymax > 0 && x < len(m.Tiles[0]) {
		c = m.Tiles[y][x]
	}
	return
}

func NewGameMap(tiles [][]int) *GameMap {
	m := GameMap{Tiles: tiles}
	m.Height = len(tiles)
	m.Width = len(tiles[0])
	return &m
}

func NewRandomGameMap(w, h int) *GameMap {
	chars := []int{'.', '#'}
	tiles := make([][]int, h)
	for y := 0; y < h; y++ {
		tiles[y] = make([]int, w)
		for x := 0; x < w; x++ {
			tiles[y][x] = chars[rand.Intn(len(chars))]
		}
	}
	return NewGameMap(tiles)
}

type Tile struct {
	Type  int // bitfield?
	Char  int
	Color int
}

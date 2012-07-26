package main

import (
	"fmt"
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
		Height: 25,
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
	s.Map = NewGameMap(10*g.Width, 10*g.Height)
	s.Map.Randomize()
	s.KeyBindings = NewKeyBindingMap([]KeyBinding{
		KeyBinding{'y', func() { s.MovePlayerNorthWest() }},
		KeyBinding{'u', func() { s.MovePlayerNorthEast() }},
		KeyBinding{'h', func() { s.MovePlayerWest() }},
		KeyBinding{'j', func() { s.MovePlayerSouth() }},
		KeyBinding{'k', func() { s.MovePlayerNorth() }},
		KeyBinding{'l', func() { s.MovePlayerEast() }},
		KeyBinding{'b', func() { s.MovePlayerSouthWest() }},
		KeyBinding{'n', func() { s.MovePlayerSouthEast() }},
		KeyBinding{'g', func() { s.Map.Grasslands() }},
		KeyBinding{'r', func() { s.Map.AddRabbits() }},
		KeyBinding{'t', func() { s.Map.AddTrees() }},
		KeyBinding{'c', func() { s.Map.MakeCaves() }},
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
	viewport_w, viewport_h := 80, 20
	mid_x, mid_y := viewport_w/2, viewport_h/2
	origin_x, origin_y := s.Player.X-mid_x, s.Player.Y-mid_y

	if origin_x < 0 {
		origin_x = 0
	} else if origin_x+viewport_w >= s.Map.Width {
		origin_x = s.Map.Width - viewport_w
	}
	if origin_y < 0 {
		origin_y = 0
	} else if origin_y+viewport_h >= s.Map.Height {
		origin_y = s.Map.Height  - viewport_h
	}

	for y := 0; y < viewport_h; y++ {
		for x := 0; x < viewport_w; x++ {
			if c, ok := s.Map.GetTile(origin_x+x, origin_y+y); ok {
				g.Window.Mvaddch(y, x, c)
			}
		}
	}

	px, py := mid_x, mid_y
	if s.Player.X < mid_x {
		px = s.Player.X
	} else if s.Player.X > origin_x+mid_x {
		px = s.Player.X - origin_x
	}
	if s.Player.Y < mid_y {
		py = s.Player.Y
	} else if s.Player.Y > origin_y+mid_y {
		py = s.Player.Y - origin_y
	}

	g.Window.Mvaddch(py, px, '@')
	g.Window.Mvaddstr(viewport_h+1, 0, fmt.Sprintf("[%d, %d]", s.Player.X, s.Player.Y))
}

func (s *PlayScreen) HandleInput(kc KeyCode) {
	px, py := s.Player.X, s.Player.Y

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

	if px != s.Player.X || py != s.Player.Y {
		if t, ok := s.Map.GetTile(s.Player.X, s.Player.Y); ok && t == '#' {
			s.Map.Tiles[s.Player.Y][s.Player.X] = '.'
			s.Player.X, s.Player.Y = px, py
		}
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

type Entity struct {
	Location Point
	Symbol int
	//Age
	//Material
}

func Eat(e *Entity) {
}

type Forager interface {
	Eat(*Entity)
}

type Eater interface {
	Eat(*Entity)
	// eaters have food preferences; e.g. grass, rabbits, rocks, etc.
	// eating yields energy
}

type Sleeper interface {
	Sleep()
	// sleepers MUST sleep
	// sleepers have different sleep duration requirements and behaviors
	// sleepers seek shelter
}

type Rabbit struct {
	Entity
	Eater
	// high agility
	//behavior: highly alert for predators, freezes/flees if alerted
	//Grazer look for food, go to food, eat food
	//Destructible
	//Skinnable
	//Edible
	//Breeder
	//Mover
	//Burrower
	//Evader
	//Attacker
	//material: flesh
	//Domesticable
}

type Fox struct {
	Entity
	Eater
	// behavior: searches for prey, sneaks when prey detected
	//Predator
	//Destructible
	//Skinnable
	//Edible
	//Breeder
	//Mover
	//Burrower
	//Evader
	//Attacker
	//material: flesh
}

type Tree struct {
	Entity
	//material: wood
	//Destructible
}

type Player struct {
	Point
}

type Point struct {
	X, Y int
}

type GameMap struct {
	Tiles         [][]int
	Width, Height int
	Entities      map[Point][]*Entity
}

func (m *GameMap) GetTile(x, y int) (t int, ok bool) {
	if 0 <= y && y < m.Height && 0 <= x && x < m.Width {
		t = m.Tiles[y][x]
		ok = true
	}
	return
}

func (m *GameMap) GetWallCount(x, y int) int {
	c := 0
	for _, p := range m.GetNeighbors(x, y) {
		if t, ok := m.GetTile(p.X, p.Y); !ok || t == '#' {
			c += 1
		}
	}
	if t, _ := m.GetTile(x, y); t == '#' {
		c += 1
	}
	return c
}

func (m *GameMap) GetNeighbors(x, y int) []Point {
	return []Point{
		Point{x - 1, y - 1}, Point{x, y - 1}, Point{x + 1, y - 1},
		Point{x - 1, y}, Point{x + 1, y},
		Point{x - 1, y + 1}, Point{x, y + 1}, Point{x + 1, y + 1},
	}
}

func (m *GameMap) PutEntity(x, y int, e interface{}) {
	e.(Entity)
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			p := Point{x, y}
			e.Location = p
			m.Entities[p] = append(m.Entities[p], e)
		}
	}
}

func (m *GameMap) Grasslands() {
	chars := []int{'.', ',', ';', '`', '\''}
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			m.Tiles[y][x] = chars[rand.Intn(len(chars))]
		}
	}
}

func (m *GameMap) AddFoxes() {
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			if rand.Intn(200) == 1 {
				m.PutEntity(x, y, Entity(&Fox{}))
			}
		}
	}
}

func (m *GameMap) AddRabbits() {
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			if rand.Intn(25) == 1 {
				m.PutEntity(x, y, &Rabbit{})
			}
		}
	}
}

func (m *GameMap) AddTrees() {
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			if rand.Intn(4) == 1 {
				m.Tiles[y][x] = 'T'
			}
		}
	}
}

func (m *GameMap) MakeCaves() {
	m.Randomize()
	for i := 0; i < 3; i++ {
		m.Smooth()
	}
}

func (m *GameMap) Smooth() {
	s := NewGameMap(m.Width, m.Height)
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			c := m.GetWallCount(x, y)
			if c >= 5 {
				s.Tiles[y][x] = '#'
			} else {
				s.Tiles[y][x] = '.'
			}
		}
	}
	m.Tiles = s.Tiles
}

func NewGameMap(w, h int) *GameMap {
	m := new(GameMap)
	m.Tiles = make([][]int, h)
	for y := 0; y < h; y++ {
		m.Tiles[y] = make([]int, w)
	}
	m.Height = h
	m.Width = w
	return m
}

func (m *GameMap) Randomize() {
	chars := []int{'.', '#'}
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			m.Tiles[y][x] = chars[rand.Intn(len(chars))]
		}
	}
}

type Tile struct {
	Type  int // bitfield?
	Char  int
	Color int
}

// const (
// 	Floor iota
// 	Wall
// )
// 
// materials (have properties: flamable, impervious, porous, etc.)
// 

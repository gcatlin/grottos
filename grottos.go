package main

import "github.com/gcatlin/curses-go"

// TODO accept cmd line args for screen size?

func main() {
	screen := curses.Initscr()
	defer curses.End()
	curses.Cbreak()
	curses.Noecho()

	screen.Addstr("Welcome to the Caves of Go!\n")
	screen.Addstr("Press any key to exit...")
	screen.Refresh()
	screen.Getch()
}

type World map[string]string

type Screen curses.Window

type Game struct {
	world  World
	screen Screen
}


package main

import (
        "github.com/gcatlin/gocurses"
        "math"
)

// TODO accept cmd line args for screen size?

func main() {
        // Initialize
	screen := curses.Initscr()
	defer curses.End()
        defer screen.Clear()
	curses.Cbreak()
	curses.Noecho()
        curses.Cursor(0)
        screen.Keypad(true)
        screen.Clear()

        px := 40
        py := 12
        for {
                // Output phase
                screen.Clear()
                screen.Mvaddch(py, px, '@')

                // Input phase
                c := screen.Getch()

                // Processing phase
                switch c {
                case 'y':
                        px--
                        py--
                case 'u':
                        px++
                        py--
                case 'h':
                        px--
                case 'j':
                        py++
                case 'k':
                        py--
                case 'l':
                        px++
                case 'b':
                        px--
                        py++
                case 'n':
                        px++
                        py++
                case 'q':
                        return
                }
                px = int(math.Min(math.Max(0.0, float64(px)), 80.0))
                py = int(math.Min(math.Max(0.0, float64(py)), 24.0))
        }
}

type World map[string]string

type Screen curses.Window

type Game struct {
	world  World
	screen Screen
}


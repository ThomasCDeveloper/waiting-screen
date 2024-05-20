package main

import (
	"fmt"
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/crazy3lf/colorconv"
)

type RGB struct {
	R uint8
	G uint8
	B uint8
}

type xy struct {
	x, y int
}

type Set struct {
	Char  string
	X     int
	Y     int
	Color RGB
}

func (set Set) getString() string {
	return fmt.Sprintf("\033[%d;%dH\033[38;2;%d;%d;%dm%s\033[0m", set.Y, set.X, set.Color.R, set.Color.G, set.Color.B, set.Char)
}

type pipe struct {
	x, y, x0, y0, targetX, targetY int
	lastDir                        int
}

func (p *pipe) update(m model) (int, int, string) {
	if p.x == p.targetX && p.y == p.targetY {
		newX, newY := p.x, p.y

		for newX == p.x || newY == p.y {
			newX = rand.Intn(m.width) + 2
			newY = rand.Intn(m.height) + 2
		}

		p.targetX = newX
		p.targetY = newY
	}

	p.x0, p.y0 = p.x, p.y

	char := " "
	newDir := 0
	if p.x < p.targetX {
		p.x++
		newDir = 0
	} else if p.x > p.targetX {
		p.x--
		newDir = 2
	} else if p.y < p.targetY {
		p.y++
		newDir = 1
	} else if p.y > p.targetY {
		p.y--
		newDir = 3
	}

	if (p.lastDir == 0 && newDir == 0) || (p.lastDir == 2 && newDir == 2) {
		char = "═"
	}
	if (p.lastDir == 1 && newDir == 1) || (p.lastDir == 3 && newDir == 3) {
		char = "║"
	}
	if (p.lastDir == 1 && newDir == 2) || (p.lastDir == 0 && newDir == 3) {
		char = "╝"
	}
	if (p.lastDir == 3 && newDir == 2) || (p.lastDir == 0 && newDir == 1) {
		char = "╗"
	}
	if (p.lastDir == 2 && newDir == 1) || (p.lastDir == 3 && newDir == 0) {
		char = "╔"
	}
	if (p.lastDir == 2 && newDir == 3) || (p.lastDir == 1 && newDir == 0) {
		char = "╚"
	}

	p.lastDir = newDir

	return p.x0, p.y0, char
}

func (m *model) set(x int, y int, rune string, color RGB) {
	set := Set{X: x, Y: y, Char: rune, Color: color}
	m.SetCommands[xy{x, y}] = set
}

type frameMsg struct{}

func animate() tea.Cmd {
	return tea.Tick(time.Second/fps, func(_ time.Time) tea.Msg {
		return frameMsg{}
	})
}

type model struct {
	SetCommands   map[xy]Set
	width, height int

	pipe         pipe
	currentColor int
}

func (m *model) reset(w, h int) {
	m.width = w
	m.height = h

	m.SetCommands = map[xy]Set{}

	for i := 0; i < h; i++ {
		fmt.Println()
	}

	m.pipe.x, m.pipe.y, m.pipe.x0, m.pipe.y0, m.pipe.targetX, m.pipe.targetY = w/2, h/2, w/2, h/2, w/2, h/2
	m.currentColor = 0
}

func (m model) Init() tea.Cmd {
	return animate()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	case tea.WindowSizeMsg:
		m.reset(msg.Width-2, msg.Height-2)
		return m, nil
	case frameMsg:
		x, y, char := m.pipe.update(m)

		R, G, B, _ := colorconv.HSVToRGB(float64(m.currentColor), 1, 1)
		m.set(x, y, char, RGB{R, G, B})
		m.currentColor = (m.currentColor + 1) % 360
		return m, animate()
	default:
		return m, nil
	}
}

func (m model) View() string {
	output := m.SetCommands[xy{m.pipe.x0, m.pipe.y0}].getString()

	return output
}

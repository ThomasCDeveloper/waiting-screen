package main

// A simple example demonstrating how to draw and animate on a cellular grid.
// Note that the cellbuffer implementation in this example does not support
// double-width runes.

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	fps = 60
)

type cellbuffer struct {
	cells  []string
	stride int
}

func (c *cellbuffer) init(w, h int) {
	if w == 0 {
		return
	}
	c.stride = w
	c.cells = make([]string, w*h)
	c.wipe()
}

func (c cellbuffer) set(x int, y int, rune string) {
	i := y*c.stride + x
	if i > len(c.cells)-1 || x < 0 || y < 0 || x >= c.width() || y >= c.height() {
		return
	}
	c.cells[i] = rune
}

func (c *cellbuffer) wipe() {
	for i := range c.cells {
		c.cells[i] = " "
	}
}

func (c cellbuffer) width() int {
	return c.stride
}

func (c cellbuffer) height() int {
	h := len(c.cells) / c.stride
	if len(c.cells)%c.stride != 0 {
		h++
	}
	return h
}

func (c cellbuffer) ready() bool {
	return len(c.cells) > 0
}

func (c cellbuffer) String() string {
	var b strings.Builder
	for i := 0; i < len(c.cells); i++ {
		if i > 0 && i%c.stride == 0 && i < len(c.cells)-1 {
			b.WriteRune('\n')
		}
		b.WriteString(c.cells[i])
	}
	return b.String()
}

type frameMsg struct{}

func animate() tea.Cmd {
	return tea.Tick(time.Second/fps, func(_ time.Time) tea.Msg {
		return frameMsg{}
	})
}

type xy struct {
	x, y int
}

type pipe struct {
	tail    []xy
	lastdir int
	char    string
}

func (p *pipe) reset(x, y int) {
	p.tail = []xy{{x, y}}
	p.char = "."
}

func (p *pipe) update(m model) {
	// tweak probabilities (prob by pipe ?)
	// 1/2 chances to go forward
	// 1/4 chances to turn either left or right
	lastxy := p.tail[len(p.tail)-1]
	nextxy := lastxy

	newdir := p.lastdir

	if rand.Intn(10) > 8 {
		if rand.Intn(2) == 0 {
			newdir++
		} else {
			newdir--
		}

		newdir = (newdir + 4) % 4
	}

	switch newdir {
	case 0:
		nextxy.x++
	case 1:
		nextxy.y++
	case 2:
		nextxy.x--
	default:
		nextxy.y--
	}

	if (p.lastdir == 0 && newdir == 1) || (p.lastdir == 3 && newdir == 2) {
		p.char = "╗"
	} else if (p.lastdir == 0 && newdir == 0) || (p.lastdir == 2 && newdir == 2) {
		p.char = "═"
	} else if (p.lastdir == 0 && newdir == 3) || (p.lastdir == 1 && newdir == 2) {
		p.char = "╝"
	} else if (p.lastdir == 1 && newdir == 1) || (p.lastdir == 3 && newdir == 3) {
		p.char = "║"
	} else if (p.lastdir == 2 && newdir == 3) || (p.lastdir == 1 && newdir == 0) {
		p.char = "╚"
	} else if (p.lastdir == 2 && newdir == 1) || (p.lastdir == 3 && newdir == 0) {
		p.char = "╔"
	} else {
		p.char = "."
	}

	p.lastdir = newdir

	h, w := m.cells.height(), m.cells.width()

	nextxy.x += w
	nextxy.x = nextxy.x % w

	nextxy.y += h
	nextxy.y = nextxy.y % h

	p.tail = append(p.tail, nextxy)

	if len(p.tail) > 10 {
		p.tail = p.tail[1:]
	}
}

type model struct {
	cells cellbuffer
	pipe  pipe
}

func (m model) Init() tea.Cmd {
	return animate()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	case tea.WindowSizeMsg:
		m.pipe.reset(msg.Width/2, msg.Height/2)
		m.cells.init(msg.Width, msg.Height)
		return m, nil

	case frameMsg:
		if !m.cells.ready() {
			return m, nil
		}

		m.pipe.update(m)

		if len(m.pipe.tail) <= 2 {
			return m, animate()
		}

		nextxy := m.pipe.tail[len(m.pipe.tail)-2]

		m.cells.set(nextxy.x, nextxy.y, m.pipe.char)

		return m, animate()

	default:

		return m, nil
	}
}

func (m model) View() string {
	return m.cells.String()
}

func main() {
	m := model{}
	m.pipe.reset(40, 40)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Uh oh:", err)
		os.Exit(1)
	}
}

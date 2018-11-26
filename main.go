package main

import (
	"container/list"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

var (
	fieldSize    = 50
	pointSize    = pixel.V(10, 10)
	padding      = pixel.V(1, 1)
	speedDefault = 0.1
	speedBoost   = 0.03
)

type game struct {
	field [][]point
	snake *snake
}

func (g *game) createField() {
	g.field = make([][]point, fieldSize)
	for x := range g.field {
		g.field[x] = make([]point, fieldSize)
		for y := 0; y < fieldSize; y++ {
			g.field[x][y] = point{
				kind: kindEmpty,
				pos:  pixel.V(float64(x), float64(y)),
			}
		}
	}
}

func (g *game) spawnSnake() {
	g.snake = &snake{
		dir:   right,
		parts: list.New(),
		speed: speedDefault,
	}
	g.snake.parts.PushFront(g.field[0][fieldSize-1].pos)
	g.snake.grows = 3
	g.snake.applyChanges(g)
}

func (g *game) newPoint() {
again:
	x := rand.Int31n(int32(fieldSize))
	y := rand.Int31n(int32(fieldSize))
	if g.field[x][y].kind != kindEmpty {
		goto again
	}
	g.field[x][y].kind = kindPoint
}

func (g *game) setPointKind(pos pixel.Vec, kind pointKind) {
	g.field[int(pos.X)][int(pos.Y)].kind = kind
}

func (g *game) getPointKind(pos pixel.Vec) pointKind {
	return g.field[int(pos.X)][int(pos.Y)].kind
}

func (g *game) end() {
	g.createField()
	g.newPoint()
	g.spawnSnake()
}

func (g *game) update(dt float64) {
	g.snake.update(dt, g)
}

func (g *game) getBounds() pixel.Rect {
	return pixel.Rect{
		Min: pixel.ZV,
		Max: pointSize.Scaled(float64(fieldSize)).Add(padding.Scaled(float64(fieldSize + 1))),
	}
}

func (g *game) render(dt float64, imd *imdraw.IMDraw) {
	imd.Color = colornames.White
	bounds := g.getBounds()
	imd.Push(bounds.Min, bounds.Max)
	imd.Rectangle(0)
	for x := 0; x < len(g.field); x++ {
		for y := 0; y < len(g.field[x]); y++ {
			g.field[x][y].draw(dt, imd)
		}
	}
}

type snake struct {
	dir    direction
	newDir direction
	parts  *list.List
	grows  int

	speed    float64
	lastMove float64
}

func (s *snake) head() pixel.Vec {
	return s.parts.Front().Value.(pixel.Vec)
}

func (s *snake) update(dt float64, g *game) {
	s.lastMove += dt
	if s.lastMove >= s.speed {
		s.lastMove = 0
	} else {
		return
	}

	if s.newDir != s.dir && s.dir.opposite() != s.newDir {
		s.dir = s.newDir
	}

	newHead := s.head().Add(s.dir.motion())
	if newHead.X < 0 || newHead.Y < 0 || int(newHead.X) >= fieldSize || int(newHead.Y) >= fieldSize {
		g.end()
		return
	}

	if g.getPointKind(newHead) == kindEmpty || g.getPointKind(newHead) == kindPoint {
		if g.getPointKind(newHead) == kindPoint {
			s.grows = 3
			g.newPoint()
		}
		s.parts.PushFront(newHead)
	} else {
		g.end()
		return
	}
	if s.grows == 0 {
		back := s.parts.Back()
		g.setPointKind(back.Value.(pixel.Vec), kindEmpty)
		s.parts.Remove(back)
	} else {
		s.grows--
	}
	s.applyChanges(g)
}

func (s *snake) applyChanges(g *game) {
	front := s.parts.Front()
	g.setPointKind(front.Value.(pixel.Vec), kindSnakeHead)
	for e := front.Next(); e != nil; e = e.Next() {
		g.setPointKind(e.Value.(pixel.Vec), kindSnake)
	}
}

func (s *snake) size() int {
	return s.parts.Len()
}

type point struct {
	kind pointKind
	pos  pixel.Vec
}

func (p point) draw(dt float64, imd *imdraw.IMDraw) {
	if p.kind == kindEmpty {
		return
	}
	tp := p.pos.ScaledXY(pointSize).Add(p.pos.ScaledXY(padding)).Add(padding)
	imd.Color = p.kind.color()
	imd.Push(tp, tp.Add(pointSize))
	imd.Rectangle(0)
}

func drawPoints(text *text.Text, word string, points int) {
	_, _ = fmt.Fprint(text, word+": ")
	text.Color = colornames.Lightgreen
	_, _ = fmt.Fprintln(text, points)
	text.Color = colornames.White
}

func run() {
	// create field
	game := game{}
	game.createField()
	game.spawnSnake()
	game.newPoint()
	bounds := game.getBounds()

	cfg := pixelgl.WindowConfig{
		Title:  "Змейка от бога",
		Bounds: pixel.R(0, 0, bounds.Max.X+100, bounds.Max.Y),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	textAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)

	best := 0
	imd := imdraw.New(nil)
	canvas := pixelgl.NewCanvas(win.Bounds())
	last := time.Now()
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()

		win.Clear(colornames.Black)
		canvas.SetBounds(win.Bounds())
		canvas.Clear(pixel.Alpha(0))

		if win.JustPressed(pixelgl.KeyLeft) {
			game.snake.newDir = left
		}
		if win.JustPressed(pixelgl.KeyRight) {
			game.snake.newDir = right
		}
		if win.JustPressed(pixelgl.KeyUp) {
			game.snake.newDir = up
		}
		if win.JustPressed(pixelgl.KeyDown) {
			game.snake.newDir = down
		}
		if win.Pressed(pixelgl.KeyTab) {
			game.snake.speed = speedBoost
		} else {
			game.snake.speed = speedDefault
		}

		game.update(dt)
		game.render(dt, imd)
		imd.Draw(canvas)

		best = int(math.Max(float64(game.snake.size()), float64(best)))
		txt := text.New(bounds.Max.Add(pixel.V(5, -13)), textAtlas)
		drawPoints(txt, "Points", game.snake.size())
		drawPoints(txt, "Best", best)
		_, _ = fmt.Fprintln(txt)
		_, _ = fmt.Fprintln(txt, "Controls:")
		_, _ = fmt.Fprintln(txt, "Arrows")
		_, _ = fmt.Fprintln(txt, "TAB - Boost")
		txt.Draw(win, pixel.IM)

		canvas.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}

package main

import (
	"github.com/faiface/pixel"
	"golang.org/x/image/colornames"
	"image/color"
)

type pointKind uint8

const (
	kindWall pointKind = iota
	kindEmpty
	kindSnake
	kindSnakeHead
	kindPoint
)

func (kind pointKind) color() color.Color {
	switch kind {
	case kindPoint:
		return colornames.Red
	case kindSnake:
		return colornames.Deepskyblue
	case kindSnakeHead:
		return colornames.Red
	case kindWall:
		return colornames.Gray
	}
	return colornames.White
}

type direction uint8

const (
	left direction = iota
	right
	up
	down
)

func (d direction) opposite() direction {
	switch d {
	case left:
		return right
	case right:
		return left
	case up:
		return down
	case down:
		return up
	}
	return down
}

func (d direction) motion() pixel.Vec {
	switch d {
	case left:
		return pixel.V(-1, 0)
	case right:
		return pixel.V(1, 0)
	case up:
		return pixel.V(0, 1)
	case down:
		return pixel.V(0, -1)
	}
	return pixel.ZV
}

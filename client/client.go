// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build js

package main

import (
	"log"
	"syscall/js"

	"github.com/googleforgames/space-agon/game"
)

func main() {
	js.Global().Call("whenLoaded", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		c, err := newClient()
		if err == nil {
			c.scheduleFrame()
		} else {
			js.Global().Get("document").Get("body").Set("innerHTML", js.ValueOf(err.Error()))
		}
		return nil
	}))

	<-make(chan struct{})
}

func newClient() (*client, error) {
	inp := &game.Input{
		IsRendered: true,
	}
	js.Global().Get("document").Call("addEventListener", "keydown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// log.Println("keydown", args[0].Get("code").String())
		switch args[0].Get("code").String() {
		case "ArrowUp":
			inp.Up.Down()
		case "ArrowLeft":
			inp.Left.Down()
		case "ArrowRight":
			inp.Right.Down()
		case "ArrowDown":
			inp.Down.Down()
		case "Space":
			inp.Fire.Down()
		}
		return nil
	}))
	js.Global().Get("document").Call("addEventListener", "keyup", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// log.Println("keyup", args[0].Get("code").String())
		switch args[0].Get("code").String() {
		case "ArrowUp":
			inp.Up.Up()
		case "ArrowLeft":
			inp.Left.Up()
		case "ArrowRight":
			inp.Right.Up()
		case "ArrowDown":
			inp.Down.Up()
		case "Space":
			inp.Fire.Up()
		}
		return nil
	}))

	log.Println("Initiating Graphics.")
	gr, err := NewGraphics()
	if err != nil {
		return nil, err
	}

	return &client{
		gr:            gr,
		g:             game.NewGame(),
		inp:           inp,
		lastTimestamp: js.Global().Get("performance").Call("now").Float(),
	}, nil
}

type client struct {
	gr            *graphics
	g             *game.Game
	inp           *game.Input
	lastTimestamp float64
}

func (c *client) scheduleFrame() {
	js.Global().Get("window").Call("requestAnimationFrame", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		now := args[0].Float()
		c.inp.Dt = float32((now - c.lastTimestamp) / 1000)
		c.lastTimestamp = now
		c.frame()
		return nil
	}))
}

var rotation = float32(0)

func (c *client) frame() {
	c.g.Step(c.inp)

	c.gr.Clear()

	{
		i := c.g.E.NewIter()
		i.Require(game.PosKey)
		i.Require(game.KeepInCameraKey)

		xMin := float32(-10)
		yMin := float32(-10)
		xMax := float32(10)
		yMax := float32(10)

		for i.Next() {
			x := (*i.Pos())[0]
			y := (*i.Pos())[1]
			boundary := float32(0)

			sprite := i.Sprite()
			if sprite != nil {
				boundary = spritemap[*sprite].size * 2
			}

			if x-boundary < xMin {
				xMin = x - boundary
			}
			if x+boundary > xMax {
				xMax = x + boundary
			}
			if y-boundary < yMin {
				yMin = y - boundary
			}
			if y+boundary > yMax {
				yMax = y + boundary
			}
		}

		c.gr.SetCamera(xMin, yMin, xMax, yMax)
	}

	// count := 0
	{
		i := c.g.E.NewIter()
		i.Require(game.PosKey)
		i.Require(game.SpriteKey)
		for i.Next() {
			// count++
			p := *i.Pos()
			rot := i.Rot()
			rotation := float32(0)
			if rot != nil {
				rotation = *rot
			}
			c.gr.Sprite(spritemap[*i.Sprite()], p[0], p[1], rotation)
		}
	}

	c.gr.Flush()

	c.g.FrameEnd()
	c.inp.FrameEndReset()

	c.scheduleFrame()
}

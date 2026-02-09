package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

// --- Engine Configuration ---
const (
	screenWidth  = 320
	screenHeight = 240
	toolbarWidth = 60
	gameWidth    = screenWidth - toolbarWidth
	gameHeight   = screenHeight
	renderScale  = 3
)

// --- Material Definitions ---
const (
	Empty = iota
	Wall
	Sand
	Water
	Plant
	Fire
	Smoke
	Stone
	Lava
	TNT
	Gas
	Acid
	Steam
	Glass
	Volt
	Wood
)

var materials = []int{
	Sand, Water, Plant, Wood, Fire, Smoke, Wall, Stone,
	Lava, Acid, TNT, Gas, Volt, Glass, Steam, Empty,
}

var matNames = map[int]string{
	Empty: "Eraser", Wall: "Static Wall", Sand: "Sand", Water: "Water",
	Plant: "Plant", Fire: "Fire", Smoke: "Smoke", Stone: "Stone",
	Lava: "Lava", TNT: "TNT", Gas: "Gas", Acid: "Acid",
	Steam: "Steam", Glass: "Glass", Volt: "VoltBolt", Wood: "Wood",
}

// getMatColor returns the RGBA color associated with a specific material ID.
func getMatColor(m int) color.RGBA {
	switch m {
	case Empty:
		return color.RGBA{10, 10, 15, 255}
	case Sand:
		return color.RGBA{235, 200, 75, 255}
	case Wall:
		return color.RGBA{100, 100, 110, 255}
	case Water:
		return color.RGBA{50, 130, 255, 255}
	case Plant:
		return color.RGBA{40, 180, 40, 255}
	case Fire:
		return color.RGBA{255, 100, 0, 255}
	case Smoke:
		return color.RGBA{120, 120, 130, 180}
	case Stone:
		return color.RGBA{80, 80, 90, 255}
	case Lava:
		return color.RGBA{255, 50, 0, 255}
	case TNT:
		return color.RGBA{150, 40, 40, 255}
	case Gas:
		return color.RGBA{180, 120, 200, 100}
	case Acid:
		return color.RGBA{150, 255, 0, 255}
	case Steam:
		return color.RGBA{200, 200, 255, 100}
	case Glass:
		return color.RGBA{180, 220, 255, 100}
	case Volt:
		return color.RGBA{255, 255, 200, 255}
	case Wood:
		return color.RGBA{100, 60, 20, 255}
	default:
		return color.RGBA{255, 0, 255, 255}
	}
}

type Game struct {
	grid       [gameWidth * gameHeight]uint8
	pixels     []byte
	frameCount int
	selected   int
	brushSize  int
}

func NewGame() *Game {
	return &Game{
		pixels:    make([]byte, screenWidth*screenHeight*4),
		selected:  Sand,
		brushSize: 4,
	}
}

func (g *Game) Update() error {
	// Global Key Handling
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.grid = [gameWidth * gameHeight]uint8{}
	}

	// Input Processing: Brush and Selection
	_, dy := ebiten.Wheel()
	if dy != 0 {
		g.brushSize += int(dy)
	}
	if g.brushSize < 1 {
		g.brushSize = 1
	}

	mx, my := ebiten.CursorPosition()
	if mx >= gameWidth {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			col, row := (mx-gameWidth)/25, my/25
			idx := row*2 + col
			if idx >= 0 && idx < len(materials) {
				g.selected = materials[idx]
			}
		}
	} else if mx >= 0 && my >= 0 && my < gameHeight {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			g.spawn(mx, my, g.selected)
		}
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
			g.spawn(mx, my, Empty)
		}
	}

	g.runPhysics()
	g.frameCount++
	return nil
}

// runPhysics iterates through the grid to apply material-specific behaviors.
func (g *Game) runPhysics() {
	for y := 0; y < gameHeight; y++ {
		// Sweep direction logic to minimize visual artifacts
		startX, endX, stepX := 0, gameWidth, 1
		if g.frameCount%2 == 0 {
			startX, endX, stepX = gameWidth-1, -1, -1
		}

		for x := startX; x != endX; x += stepX {
			idx := y*gameWidth + x
			cell := g.grid[idx]

			// Optimization: Skip static/empty cells
			if cell == Empty || cell == Wall || cell == Stone || cell == Glass || cell == Wood {
				continue
			}

			// --- BUOYANT GAS PHYSICS (Smoke, Steam, Gas) ---
			if cell == Smoke || cell == Steam || cell == Gas {
				if y > 0 {
					above := (y-1)*gameWidth + x
					// Primary upward movement
					if g.grid[above] == Empty {
						g.move(idx, above)
					} else {
						// Secondary horizontal drift
						dir := rand.Intn(2)*2 - 1
						if x+dir >= 0 && x+dir < gameWidth {
							diag := (y-1)*gameWidth + (x + dir)
							if g.grid[diag] == Empty {
								g.move(idx, diag)
							}
						}
					}
				} else {
					// Dissipate at the ceiling
					g.grid[idx] = Empty
				}
				// Probabilistic decay for smoke and steam
				if (cell == Smoke || cell == Steam) && rand.Float32() < 0.015 {
					g.grid[idx] = Empty
				}
				continue
			}

			// --- PLANT LIFE-CYCLE ---
			if cell == Plant {
				if rand.Float32() < 0.04 {
					g.handlePlantGrowth(x, y)
				}
				continue
			}

			below := (y+1)*gameWidth + x

			// --- GRANULAR PHYSICS (Sand, TNT) ---
			if (cell == Sand || cell == TNT) && y < gameHeight-1 {
				if g.grid[below] == Empty {
					g.move(idx, below)
				} else if g.grid[below] == Water {
					g.swap(idx, below)
				} else {
					dir := rand.Intn(2)*2 - 1
					if x+dir >= 0 && x+dir < gameWidth {
						diag := (y+1)*gameWidth + (x + dir)
						if g.grid[diag] == Empty {
							g.move(idx, diag)
						}
					}
				}
			}

			// --- FLUID PHYSICS (Water, Acid, Lava) ---
			if cell == Water || cell == Acid || cell == Lava {
				if y < gameHeight-1 && g.grid[below] == Empty {
					g.move(idx, below)
				} else {
					dir := rand.Intn(2)*2 - 1
					if x+dir >= 0 && x+dir < gameWidth {
						side := y*gameWidth + (x + dir)
						if g.grid[side] == Empty {
							g.move(idx, side)
						}
					}
				}
			}

			// --- COMBUSTION LOGIC ---
			if cell == Fire {
				if rand.Float32() < 0.12 {
					g.grid[idx] = Smoke
				}
				g.spreadFire(x, y)
			}

			// --- ELECTRICAL DISCHARGE ---
			if cell == Volt {
				g.grid[idx] = Empty
				tx, ty := x+(rand.Intn(3)-1), y+1
				if tx >= 0 && tx < gameWidth && ty < gameHeight {
					tIdx := ty*gameWidth + tx
					if g.grid[tIdx] == Empty {
						g.grid[tIdx] = Volt
					} else if g.grid[tIdx] == Sand {
						g.grid[tIdx] = Glass
					}
				}
			}
		}
	}
}

func (g *Game) handlePlantGrowth(x, y int) {
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			tx, ty := x+dx, y+dy
			if tx >= 0 && tx < gameWidth && ty >= 0 && ty < gameHeight {
				tIdx := ty*gameWidth + tx
				if g.grid[tIdx] == Water {
					g.grid[tIdx] = Empty
					gx, gy := x+(rand.Intn(3)-1), y+(rand.Intn(2)-1)
					if gx >= 0 && gx < gameWidth && gy >= 0 && gy < gameHeight {
						gIdx := gy*gameWidth + gx
						if g.grid[gIdx] == Empty {
							g.grid[gIdx] = Plant
						}
					}
					return
				}
			}
		}
	}
}

func (g *Game) spreadFire(x, y int) {
	tx, ty := x+(rand.Intn(3)-1), y+(rand.Intn(3)-1)
	if tx >= 0 && tx < gameWidth && ty >= 0 && ty < gameHeight {
		tIdx := ty*gameWidth + tx
		target := g.grid[tIdx]
		if target == Wood || target == Plant || target == TNT || target == Gas {
			if target == TNT {
				g.explode(tx, ty, 12)
			} else {
				g.grid[tIdx] = Fire
			}
		}
	}
}

func (g *Game) explode(cx, cy, r int) {
	for y := -r; y <= r; y++ {
		for x := -r; x <= r; x++ {
			if x*x+y*y <= r*r {
				tx, ty := cx+x, cy+y
				if tx >= 0 && tx < gameWidth && ty >= 0 && ty < gameHeight {
					if g.grid[ty*gameWidth+tx] != Wall {
						g.grid[ty*gameWidth+tx] = Fire
					}
				}
			}
		}
	}
}

func (g *Game) spawn(cx, cy, p int) {
	r := g.brushSize
	for y := -r; y <= r; y++ {
		for x := -r; x <= r; x++ {
			if x*x+y*y <= r*r && rand.Float32() > 0.15 {
				tx, ty := cx+x, cy+y
				if tx >= 0 && tx < gameWidth && ty >= 0 && ty < gameHeight {
					idx := ty*gameWidth + tx
					if p == Wall || g.grid[idx] != Wall {
						g.grid[idx] = uint8(p)
					}
				}
			}
		}
	}
}

func (g *Game) move(f, t int) { g.grid[t], g.grid[f] = g.grid[f], Empty }
func (g *Game) swap(a, b int) { g.grid[a], g.grid[b] = g.grid[b], g.grid[a] }

func (g *Game) Draw(screen *ebiten.Image) {
	// Screen Buffer Rendering
	for i := range g.pixels {
		g.pixels[i] = 0
	}
	for y := 0; y < gameHeight; y++ {
		for x := 0; x < gameWidth; x++ {
			idx := y*gameWidth + x
			c := getMatColor(int(g.grid[idx]))
			off := (y*screenWidth + x) * 4
			g.pixels[off], g.pixels[off+1], g.pixels[off+2], g.pixels[off+3] = c.R, c.G, c.B, c.A
		}
	}
	screen.WritePixels(g.pixels)

	// User Interface Rendering
	ebitenutil.DrawRect(screen, float64(gameWidth), 0, float64(toolbarWidth), float64(screenHeight), color.RGBA{30, 30, 40, 255})
	mx, my := ebiten.CursorPosition()
	hovered := ""
	for i, m := range materials {
		bx, by := float64(gameWidth+8+(i%2)*25), float64(8+(i/2)*25)
		if g.selected == m {
			ebitenutil.DrawRect(screen, bx-2, by-2, 24, 24, color.White)
		}
		ebitenutil.DrawRect(screen, bx, by, 20, 20, getMatColor(m))
		if mx >= int(bx) && mx < int(bx+20) && my >= int(by) && my < int(by+20) {
			hovered = matNames[m]
		}
	}
	if hovered != "" {
		text.Draw(screen, hovered, basicfont.Face7x13, mx-10, my-10, color.White)
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Brush: %d", g.brushSize), gameWidth+5, screenHeight-20)
}

func (g *Game) Layout(w, h int) (int, int) { return screenWidth, screenHeight }

func main() {
	ebiten.SetWindowSize(screenWidth*renderScale, screenHeight*renderScale)
	ebiten.SetWindowTitle("Sim: Smoke & Physics Optimization")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}

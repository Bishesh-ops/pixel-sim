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

// --- CONFIG ---
const (
	screenWidth  = 320
	screenHeight = 240
	toolbarWidth = 60
	gameWidth    = screenWidth - toolbarWidth
	gameHeight   = screenHeight
	renderScale  = 3
)

// --- MATERIALS & THERMODYNAMICS ---
const (
	Empty = iota
	Metal
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
	HotMetal
	VoltTrail
)

var materials = []int{
	Sand, Water, Plant, Wood, Fire, Smoke, Metal, Stone,
	Lava, Acid, TNT, Gas, Volt, Glass, Steam, Empty,
}

var matNames = map[int]string{
	Empty:    "Eraser",
	Metal:    "Metal",
	Sand:     "Sand",
	Water:    "Water",
	Plant:    "Plant",
	Fire:     "Fire",
	Smoke:    "Smoke",
	Stone:    "Stone",
	Lava:     "Lava",
	TNT:      "TNT",
	Gas:      "Gas",
	Acid:     "Acid",
	Steam:    "Steam",
	Glass:    "Glass",
	Volt:     "Lightning",
	Wood:     "Wood",
	HotMetal: "Hot Metal",
}

func getMatColor(m int) color.RGBA {
	switch m {
	case Empty:
		return color.RGBA{10, 10, 15, 255}
	case Sand:
		return color.RGBA{235, 200, 75, 255}
	case Metal:
		return color.RGBA{130, 135, 145, 255}
	case HotMetal:
		return color.RGBA{255, 90, 20, 255}
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
		return color.RGBA{180, 40, 40, 255}
	case Gas:
		return color.RGBA{180, 120, 200, 100}
	case Acid:
		return color.RGBA{150, 255, 0, 255}
	case Steam:
		return color.RGBA{200, 200, 255, 100}
	case Glass:
		return color.RGBA{180, 220, 255, 100}
	case Volt, VoltTrail:
		return color.RGBA{200, 255, 255, 255}
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
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.grid = [gameWidth * gameHeight]uint8{}
	}

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

	for i := 0; i < 2; i++ {
		g.sim()
	}
	g.frameCount++
	return nil
}

// --- PHYSICS ENGINE ---
func (g *Game) sim() {
	// PASS 1: BOTTOM-TO-TOP (Gravity & Liquids)
	for y := gameHeight - 1; y >= 0; y-- {
		startX, endX, stepX := 0, gameWidth, 1
		if g.frameCount%2 == 0 {
			startX, endX, stepX = gameWidth-1, -1, -1
		}

		for x := startX; x != endX; x += stepX {
			idx := y*gameWidth + x
			cell := g.grid[idx]

			if cell == Empty || cell == Metal || cell == Stone || cell == Glass || cell == Wood || cell == HotMetal {
				continue
			}
			below := (y+1)*gameWidth + x

			if (cell == Sand || cell == TNT) && y < gameHeight-1 {
				if g.grid[below] == Empty {
					g.move(idx, below)
				} else if g.grid[below] == Water || g.grid[below] == Acid {
					g.swap(idx, below)
				} else {
					dir := rand.Intn(2)*2 - 1
					belowDir := (y+1)*gameWidth + x + dir
					belowOpp := (y+1)*gameWidth + x - dir
					if x+dir >= 0 && x+dir < gameWidth && g.grid[belowDir] == Empty {
						g.move(idx, belowDir)
					} else if x-dir >= 0 && x-dir < gameWidth && g.grid[belowOpp] == Empty {
						g.move(idx, belowOpp)
					}
				}
				continue
			}

			if cell == Water || cell == Acid || cell == Lava {
				if (cell == Lava && rand.Float32() < 0.6) || (cell == Acid && rand.Float32() < 0.3) {
					continue
				}
				if y < gameHeight-1 && g.grid[below] == Empty {
					g.move(idx, below)
					continue
				}
				dir := rand.Intn(2)*2 - 1
				sideDir := y*gameWidth + x + dir
				sideOpp := y*gameWidth + x - dir
				if x+dir >= 0 && x+dir < gameWidth && g.grid[sideDir] == Empty {
					g.move(idx, sideDir)
				} else if x-dir >= 0 && x-dir < gameWidth && g.grid[sideOpp] == Empty {
					g.move(idx, sideOpp)
				}
				continue
			}
		}
	}

	// PASS 2: TOP-TO-BOTTOM (Gases, Reactions, Volt)
	for y := 0; y < gameHeight; y++ {
		startX, endX, stepX := 0, gameWidth, 1
		if g.frameCount%2 != 0 {
			startX, endX, stepX = gameWidth-1, -1, -1
		}

		for x := startX; x != endX; x += stepX {
			idx := y*gameWidth + x
			cell := g.grid[idx]

			if cell == Smoke || cell == Steam || cell == Gas {
				if y > 0 {
					above := (y-1)*gameWidth + x
					if g.grid[above] == Empty {
						g.move(idx, above)
					} else if rand.Float32() < 0.5 {
						dir := rand.Intn(2)*2 - 1
						side := y*gameWidth + x + dir
						if x+dir >= 0 && x+dir < gameWidth && g.grid[side] == Empty {
							g.move(idx, side)
						}
					}
				} else {
					g.grid[idx] = Empty
				}
				if (cell == Smoke || cell == Steam) && rand.Float32() < 0.015 {
					g.grid[idx] = Empty
				}
				continue
			}

			// 1. THE TRAIL (Fades out quickly, does not clone itself)
			if cell == VoltTrail {
				if rand.Float32() < 0.6 {
					g.grid[idx] = Empty
				}
				continue
			}

			// 2. THE BOLT HEAD (Instantly raycasts downwards)
			if cell == Volt {
				g.grid[idx] = VoltTrail

				cx, cy := x, y
				for i := 0; i < 50; i++ { // Bolt length of 50 pixels
					tx, ty := cx+(rand.Intn(5)-2), cy+1 // Jagged downward path
					if tx >= 0 && tx < gameWidth && ty >= 0 && ty < gameHeight {
						tIdx := ty*gameWidth + tx
						target := g.grid[tIdx]

						// Leave a trail through empty space and gases
						if target == Empty || target == Smoke || target == Steam || target == VoltTrail {
							g.grid[tIdx] = VoltTrail
							cx, cy = tx, ty
						} else if target == Metal || target == HotMetal {
							// Hit a conductor! Shower sparks.
							if rand.Float32() < 0.6 {
								for j := 0; j < 4; j++ {
									sx, sy := tx+(rand.Intn(5)-2), ty+(rand.Intn(5)-2)
									if sx >= 0 && sx < gameWidth && sy >= 0 && sy < gameHeight && g.grid[sy*gameWidth+sx] == Empty {
										g.grid[sy*gameWidth+sx] = VoltTrail
									}
								}
							}
							break
						} else if target == Water {
							g.grid[tIdx] = Steam
							break
						} else if target == Wood || target == Plant {
							g.grid[tIdx] = Fire
							break
						} else if target == Gas || target == TNT {
							g.explode(tx, ty, 8)
							break
						} else if target == Sand {
							g.grid[tIdx] = Glass
							break
						} else {
							break // Hit ground/stone
						}
					} else {
						break // Off screen
					}
				}
				continue
			}

			if cell == Lava {
				g.updateLava(x, y)
				continue
			}
			if cell == Acid {
				g.updateAcid(x, y)
				continue
			}

			if cell == HotMetal {
				if rand.Float32() < 0.005 {
					g.grid[idx] = Metal
				}
				tx, ty := x+(rand.Intn(3)-1), y+(rand.Intn(3)-1)
				if tx >= 0 && tx < gameWidth && ty >= 0 && ty < gameHeight {
					tIdx := ty*gameWidth + tx
					target := g.grid[tIdx]
					if target == Water {
						g.grid[tIdx] = Steam
					} else if target == Wood || target == Plant {
						g.grid[tIdx] = Fire
					} else if target == TNT {
						g.explode(tx, ty, 5)
					} else if target == Gas {
						g.explode(tx, ty, 8)
					}
				}
				continue
			}

			if cell == Fire {
				if y > 0 && rand.Float32() < 0.3 {
					above := (y-1)*gameWidth + x + (rand.Intn(3) - 1)
					if above >= 0 && above < gameWidth*gameHeight && g.grid[above] == Empty {
						g.swap(idx, above)
					}
				}
				if rand.Float32() < 0.15 {
					g.grid[idx] = Smoke
				}
				g.updateFire(x, y)
				continue
			}

			// UPGRADED PLANT
			if cell == Plant {
				if rand.Float32() < 0.08 {
					g.updatePlant(x, y)
				}
				continue
			}
		}
	}
}

// --- CHEMICAL REACTION HELPERS ---

func (g *Game) updateLava(x, y int) {
	tx, ty := x+(rand.Intn(3)-1), y+(rand.Intn(3)-1)
	if tx >= 0 && tx < gameWidth && ty >= 0 && ty < gameHeight {
		tIdx := ty*gameWidth + tx
		target := g.grid[tIdx]

		if target == Water {
			g.grid[tIdx] = Steam
			g.grid[y*gameWidth+x] = Stone
		} else if target == Wood || target == Plant {
			g.grid[tIdx] = Fire
		} else if target == TNT || target == Gas {
			g.explode(tx, ty, 10)
		} else if target == Metal {
			if rand.Float32() < 0.08 {
				g.grid[tIdx] = HotMetal
			}
		} else if target == HotMetal {
			if rand.Float32() < 0.03 {
				g.grid[tIdx] = Lava
			}
		} else if target == Sand {
			if rand.Float32() < 0.02 {
				g.grid[tIdx] = Glass
			}
		}
	}
}

func (g *Game) updateFire(x, y int) {
	tx, ty := x+(rand.Intn(3)-1), y+(rand.Intn(3)-1)
	if tx >= 0 && tx < gameWidth && ty >= 0 && ty < gameHeight {
		tIdx := ty*gameWidth + tx
		target := g.grid[tIdx]

		if target == Wood || target == Plant {
			g.grid[tIdx] = Fire
		} else if target == TNT || target == Gas {
			g.explode(tx, ty, 8)
		} else if target == Metal {
			if rand.Float32() < 0.04 {
				g.grid[tIdx] = HotMetal
			}
		} else if target == HotMetal {
			if rand.Float32() < 0.01 {
				g.grid[tIdx] = Lava
			}
		} else if target == Water {
			g.grid[tIdx] = Steam
			g.grid[y*gameWidth+x] = Empty
		} else if target == Sand {
			if rand.Float32() < 0.005 {
				g.grid[tIdx] = Glass
			}
		}
	}
}

func (g *Game) updateAcid(x, y int) {
	tx, ty := x+(rand.Intn(3)-1), y+1
	if tx >= 0 && tx < gameWidth && ty >= 0 && ty < gameHeight {
		tIdx := ty*gameWidth + tx
		target := g.grid[tIdx]
		if target != Empty && target != Acid && target != Glass && target != Stone {
			if rand.Float32() < 0.15 {
				g.grid[tIdx] = Smoke
				g.grid[y*gameWidth+x] = Empty
			}
		}
	}
}

// UPGRADED EXPLOSION LOGIC
func (g *Game) explode(cx, cy, r int) {
	for y := -r; y <= r; y++ {
		for x := -r; x <= r; x++ {
			if x*x+y*y <= r*r {
				tx, ty := cx+x, cy+y
				if tx >= 0 && tx < gameWidth && ty >= 0 && ty < gameHeight {
					tIdx := ty*gameWidth + tx
					target := g.grid[tIdx]

					if target == Metal || target == Stone {
						if rand.Float32() < 0.5 {
							g.grid[tIdx] = HotMetal
						}
					} else if target == Sand {
						if rand.Float32() < 0.6 {
							g.grid[tIdx] = Glass
						}
					} else if target == Water {
						g.grid[tIdx] = Steam
					} else if target == Glass {
						if rand.Float32() < 0.4 {
							g.grid[tIdx] = Empty
						}
					} else if target != Empty {
						g.grid[tIdx] = Fire
					} else if target == Empty && rand.Float32() < 0.2 {
						g.grid[tIdx] = Fire
					}
				}
			}
		}
	}
}

// UPGRADED PLANT LOGIC
func (g *Game) updatePlant(x, y int) {
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			tx, ty := x+dx, y+dy
			if tx >= 0 && tx < gameWidth && ty >= 0 && ty < gameHeight {
				tIdx := ty*gameWidth + tx
				target := g.grid[tIdx]

				// Photosynthesis
				if target == Water || target == Smoke || target == Gas {
					g.grid[tIdx] = Plant

					// Growth biases upwards (-2)
					gx, gy := x+(rand.Intn(3)-1), y+(rand.Intn(3)-2)
					if gx >= 0 && gx < gameWidth && gy >= 0 && gy < gameHeight {
						gIdx := gy*gameWidth + gx
						if g.grid[gIdx] == Empty {
							g.grid[gIdx] = Plant
						}
					}
					return
				} else if target == Acid {
					if rand.Float32() < 0.3 {
						g.grid[y*gameWidth+x] = Smoke
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
					tIdx := ty*gameWidth + tx
					if p == Metal || p == Stone || g.grid[tIdx] != Metal {
						g.grid[tIdx] = uint8(p)
					}
				}
			}
		}
	}
}

func (g *Game) move(f, t int) { g.grid[t], g.grid[f] = g.grid[f], Empty }
func (g *Game) swap(a, b int) { g.grid[a], g.grid[b] = g.grid[b], g.grid[a] }

// --- RENDERER ---
func clamp(v, min, max int) uint8 {
	if v < min {
		return uint8(min)
	}
	if v > max {
		return uint8(max)
	}
	return uint8(v)
}

func (g *Game) Draw(screen *ebiten.Image) {
	for i := range g.pixels {
		g.pixels[i] = 0
	}
	for y := 0; y < gameHeight; y++ {
		for x := 0; x < gameWidth; x++ {
			idx := y*gameWidth + x
			mat := int(g.grid[idx])
			c := getMatColor(mat)

			if mat != Empty && mat != Metal {
				noise := (x*19349663 ^ y*83492791) % 15
				if mat == Water || mat == Lava || mat == Gas || mat == Steam || mat == HotMetal {
					noise = (x*19349663 ^ y*83492791 ^ g.frameCount*10) % 20
				}
				c.R = clamp(int(c.R)+noise-7, 0, 255)
				c.G = clamp(int(c.G)+noise-7, 0, 255)
				c.B = clamp(int(c.B)+noise-7, 0, 255)
			}

			off := (y*screenWidth + x) * 4
			g.pixels[off] = c.R
			g.pixels[off+1] = c.G
			g.pixels[off+2] = c.B
			g.pixels[off+3] = c.A
		}
	}
	screen.WritePixels(g.pixels)

	ebitenutil.DrawRect(screen, float64(gameWidth), 0, float64(toolbarWidth), float64(screenHeight), color.RGBA{20, 20, 25, 255})
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
	ebiten.SetWindowTitle("Thermodynamics Engine")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}

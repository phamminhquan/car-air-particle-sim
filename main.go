package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// Screen constants
const (
	screenWidth = 800
	screenHeight = 600
)

// Particle system constants
const (
	maxParticles = 1000
	baseWindSpeed = 5.0
)

// Particle data structure
type Particle struct {
	X, Y float32
	Vx, Vy float32
	Alpha float32
	Active bool
}

// Game structure
type Game struct {
	particles []Particle
	silhouette []BezierCurve
	rng *rand.Rand
}

// Function to create name game
func NewGame() *Game {
	// New game
	g := &Game {
		particles: make([]Particle, maxParticles),
		rng: rand.New(rand.NewSource(time.Now().UnixNano())), // seed rng with time
	}
	// Define car silhouette
	g.silhouette = []BezierCurve {
		// Front bumper
		{P0: Vec2{200, 420}, P1: Vec2{210, 390}, P2: Vec2{230, 380}},
		// Hood
		{P0: Vec2{230, 380}, P1: Vec2{265, 355}, P2: Vec2{320, 350}},
		// Windshielf slant
		{P0: Vec2{320, 350}, P1: Vec2{360, 310}, P2: Vec2{400, 300}},
		// Roofline
		{P0: Vec2{400, 300}, P1: Vec2{460, 300}, P2: Vec2{520, 305}},
		// Rear window
		{P0: Vec2{520, 290}, P1: Vec2{570, 330}, P2: Vec2{600, 370}},
		// Rear bumper
		{P0: Vec2{600, 370}, P1: Vec2{620, 370}, P2: Vec2{620, 420}},
	}
	return g
}

// Game method: spawn new particles
func (g *Game) spawnParticle() {
	for i := range g.particles {
		if !g.particles[i].Active {
			g.particles[i] = Particle {
				X: 0, // spawn on left
				Y: g.rng.Float32() * screenHeight, // at random height
				Vx: baseWindSpeed + g.rng.Float32() * 2, // variation of speed
				Vy: (g.rng.Float32() - 0.5) * 0.5,
				Alpha: 1.0,
				Active: true,
			}
			break
		}
	}
}

// Update game
func (g *Game) Update() error {
	// Continuously spawn new particles
	for i := 0; i < 5; i++ {
		g.spawnParticle()
	}
	// Update each particle position and handle collision
	for i := range g.particles {
		if !g.particles[i].Active {
			continue // skip this particle if it's not Active
		}
		// Update position
		p := &g.particles[i]
		//oldX, oldY := p.X, p.Y
		//oldY := p.Y
		p.X += p.Vx
		p.Y += p.Vy
		// Handle collision
		for _, curve := range g.silhouette {
			// Fast horizontal bounding box check before fine sampling
			minX := float32(math.Min(float64(curve.P0.X), float64(curve.P2.X)))
			maxX := float32(math.Max(float64(curve.P0.X), float64(curve.P2.X)))
			if p.X < minX || p.X > maxX {
				continue // skip if particle is outside of bounding box
			}
			// Find closest segment parameter t by stepping through curve X space
			// Because curves are mapped flat across the car profile, we can approximate
			// t via X linear interpolation
			t := (p.X - curve.P0.X) / (curve.P2.X - curve.P0.X)
			if t >= 0 && t <= 1 {
				curvePt := curve.Eval(t)
				// Collision detection
				// Check if particle crossed under roof
				if p.Y >= curvePt.Y && p.Y <= 420 {
					// Snap particle precisely to the curve surface
					p.Y = curvePt.Y - 1.0
					// Fetch tangent slope at interception window
					tan := curve.Tan(t)
					// Calculate incoming velocity magnitude
					speed := float32(math.Sqrt(float64(p.Vx * p.Vx + p.Vy * p.Vy)))
					// Redirect velocity along the surface vector
					p.Vx = tan.X * speed
					p.Vy = tan.Y * speed
					break
				}
			}
		}
		// Apply atmospheric drag to guide displaced air back into horizontal lane
		if p.X > 620 { // behind car tail (wake area)
			p.Vy += (0 - p.Vy) * 0.03 // slower stabilization mimics wake vortex separation
		} else {
			p.Vy += (0 - p.Vy) * 0.07 // active laminar correction over the hood
		}
		// Despawn conditions (particle fly outside screen)
		if p.X > screenWidth || p.Y < 0 || p.Y > screenHeight {
			p.Active = false
		}
	}
	return nil
}

// Draw function
func (g *Game) Draw(screen *ebiten.Image) {
	// Draw the static car shape
	var path vector.Path
	path.MoveTo(g.silhouette[0].P0.X, g.silhouette[0].P0.Y)
	for _, curve := range g.silhouette {
		path.QuadTo(curve.P1.X, curve.P1.Y, curve.P2.X, curve.P2.Y)
	}
	// Complete bottom floor enclosure box for a clean structural look
	path.LineTo(620, 420)
	path.LineTo(200, 420)
	path.Close()
	// Draw car
	drawOp := &vector.DrawPathOptions{}
	drawOp.ColorScale.ScaleWithColor(color.NRGBA{45, 48, 54, 255})
	vector.FillPath(screen, &path, nil, drawOp)
	// Render particles (if Active)
	for _, p := range g.particles {
		if !p.Active {
			continue // skip if particle not Active
		}
		// Create dynamic trail or point line
		//vector.DrawFilledCircle(screen, p.X, p.Y, 1.5,
		//	color.NRGBA{173, 216, 230, 200}, false)
		vector.StrokeLine(screen, p.X-p.Vx*0.8, p.Y-p.Vy*0.8, p.X, p.Y, 1.2,
			color.NRGBA{135, 206, 250, 180}, false)
	}
	// Display Tick per second
	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f\nFPS: %0.2f",
		ebiten.ActualTPS(), ebiten.ActualFPS()))
}

// Game layout
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	fmt.Println("Hello World!")
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Particle Simulation")
	if err := ebiten.RunGame(NewGame()); err != nil {
		panic(err)
	}
}

package main

import (
	"math"
)

// 2D vector type
type Vec2 struct {
	X, Y float32
}

// BezierCurve defines a quadratic bezier segment
// P0 = start, P1 = control, P2 = end
type BezierCurve struct {
	P0, P1, P2 Vec2
}

// Method evaluate returns the point at parameter t [0, 1]
func (b BezierCurve) Eval(t float32) Vec2 {
	u := 1 - t
	tt := t * t
	uu := u * u
	return Vec2 {
		X: uu*b.P0.X + 2*u*t*b.P1.X + tt*b.P2.X,
		Y: uu*b.P0.Y + 2*u*t*b.P1.Y + tt*b.P2.Y,
	}
}

// Method tangent returns the normalized direction vector at parameter t
func (b BezierCurve) Tan(t float32) Vec2 {
	tx := 2*(1-t)*(b.P1.X-b.P0.X) + 2*t*(b.P2.X-b.P1.X)
	ty := 2*(1-t)*(b.P1.Y-b.P0.Y) + 2*t*(b.P2.Y-b.P1.Y)
	mag := float32(math.Sqrt(float64(tx*tx + ty*ty)))
	if mag == 0 {
		return Vec2{1, 0}
	}
	return Vec2{tx/mag, ty/mag}
}

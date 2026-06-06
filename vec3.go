package main

import (
	"fmt"
	"io"
	"math"
	"math/rand/v2"
)

// Vec3 is a 3-component vector used for points, directions, and colours.
type Vec3 struct {
	X, Y, Z float64
}

// Type aliases for readability — no extra type-safety overhead.
type Point3 = Vec3
type Color = Vec3

// NewVec3 constructs a Vec3.
func NewVec3(x, y, z float64) Vec3 { return Vec3{x, y, z} }

// --- Arithmetic ---

func (v Vec3) Add(u Vec3) Vec3   { return Vec3{v.X + u.X, v.Y + u.Y, v.Z + u.Z} }
func (v Vec3) Sub(u Vec3) Vec3   { return Vec3{v.X - u.X, v.Y - u.Y, v.Z - u.Z} }
func (v Vec3) Mul(u Vec3) Vec3   { return Vec3{v.X * u.X, v.Y * u.Y, v.Z * u.Z} }
func (v Vec3) Scale(t float64) Vec3 { return Vec3{v.X * t, v.Y * t, v.Z * t} }
func (v Vec3) Div(t float64) Vec3   { return v.Scale(1 / t) }
func (v Vec3) Neg() Vec3            { return Vec3{-v.X, -v.Y, -v.Z} }

// --- Length / dot / cross ---

func (v Vec3) LengthSquared() float64 { return v.X*v.X + v.Y*v.Y + v.Z*v.Z }
func (v Vec3) Length() float64        { return math.Sqrt(v.LengthSquared()) }

// Dot returns the scalar dot product of two vectors.
func Dot(a, b Vec3) float64 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}

// Cross returns the cross product (vector perpendicular to both inputs).
func Cross(a, b Vec3) Vec3 {
	return Vec3{
		a.Y*b.Z - a.Z*b.Y,
		a.Z*b.X - a.X*b.Z,
		a.X*b.Y - a.Y*b.X,
	}
}

// Unit returns a unit vector (length 1) in the same direction.
func (v Vec3) Unit() Vec3 { return v.Div(v.Length()) }

// NearZero returns true if the vector is very close to the zero vector.
func (v Vec3) NearZero() bool {
	const s = 1e-8
	return math.Abs(v.X) < s && math.Abs(v.Y) < s && math.Abs(v.Z) < s
}

// --- Random vectors ---

// RandomVec3 returns a vector with each component in [0, 1).
func RandomVec3() Vec3 {
	return Vec3{rand.Float64(), rand.Float64(), rand.Float64()}
}

// RandomVec3Range returns a vector with each component in [min, max).
func RandomVec3Range(min, max float64) Vec3 {
	r := max - min
	return Vec3{
		min + r*rand.Float64(),
		min + r*rand.Float64(),
		min + r*rand.Float64(),
	}
}

// RandomUnitVector returns a unit vector pointing in a random direction.
// Uses the rejection method: sample the unit ball, then normalise.
func RandomUnitVector() Vec3 {
	for {
		p := RandomVec3Range(-1, 1)
		lensq := p.LengthSquared()
		if 1e-160 < lensq && lensq <= 1 {
			return p.Div(math.Sqrt(lensq))
		}
	}
}

// RandomInUnitDisk returns a random point inside the unit disk (z = 0).
// Used for defocus blur (depth-of-field).
func RandomInUnitDisk() Vec3 {
	for {
		p := Vec3{rand.Float64()*2 - 1, rand.Float64()*2 - 1, 0}
		if p.LengthSquared() < 1 {
			return p
		}
	}
}

// --- Colour output ---

// linearToGamma applies a simple gamma-2 correction (sqrt).
func linearToGamma(c float64) float64 {
	if c > 0 {
		return math.Sqrt(c)
	}
	return 0
}

// WriteColor writes a pixel colour to w in PPM byte format.
// Applies gamma correction and clamps to [0, 255].
func (c Color) WriteColor(w io.Writer) {
	r := linearToGamma(c.X)
	g := linearToGamma(c.Y)
	b := linearToGamma(c.Z)
	intensity := Interval{0.000, 0.999}
	fmt.Fprintf(w, "%d %d %d\n",
		int(256*intensity.Clamp(r)),
		int(256*intensity.Clamp(g)),
		int(256*intensity.Clamp(b)),
	)
}

// Reflect computes the reflection of v about normal n.
func Reflect(v, n Vec3) Vec3 {
	return v.Sub(n.Scale(2 * Dot(v, n)))
}

// Refract computes the refraction of unit vector uv crossing a surface
// with the given normal, where etaiOverEtat = η_in / η_out.
func Refract(uv, n Vec3, etaiOverEtat float64) Vec3 {
	cosTheta := math.Min(Dot(uv.Neg(), n), 1.0)
	rOutPerp := uv.Add(n.Scale(cosTheta)).Scale(etaiOverEtat)
	rOutParallel := n.Scale(-math.Sqrt(math.Abs(1.0 - rOutPerp.LengthSquared())))
	return rOutPerp.Add(rOutParallel)
}

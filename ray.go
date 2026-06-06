package main

// Ray represents a half-line: P(t) = Origin + t * Direction.
// Plug in t = 0 to get the origin; larger t means further along the ray.
type Ray struct {
	Origin    Point3
	Direction Vec3
}

// At returns the point along the ray at parameter t.
func (r Ray) At(t float64) Point3 {
	return r.Origin.Add(r.Direction.Scale(t))
}

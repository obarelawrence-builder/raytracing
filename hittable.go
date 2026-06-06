package main

import "math"

// HitRecord captures everything we need to know about a ray-surface intersection.
type HitRecord struct {
	P         Point3   // the hit point in world space
	Normal    Vec3     // unit surface normal at the hit point
	T         float64  // ray parameter: P = ray.At(T)
	FrontFace bool     // true if the ray hit the outside of the surface
	Mat       Material // the material at the hit point
}

// SetFaceNormal stores a normal that always points against the incoming ray.
// outwardNormal must already be a unit vector.
func (rec *HitRecord) SetFaceNormal(r Ray, outwardNormal Vec3) {
	rec.FrontFace = Dot(r.Direction, outwardNormal) < 0
	if rec.FrontFace {
		rec.Normal = outwardNormal
	} else {
		rec.Normal = outwardNormal.Neg()
	}
}

// Hittable is the interface implemented by every object in the scene.
type Hittable interface {
	// Hit tests whether the ray intersects the object within the interval rayT.
	// If it does, it fills rec with details about the closest intersection.
	Hit(r Ray, rayT Interval, rec *HitRecord) bool
}

// HittableList is a collection of Hittables. It itself satisfies Hittable by
// delegating to each child and tracking the closest hit.
type HittableList struct {
	Objects []Hittable
}

// Add appends an object to the list.
func (hl *HittableList) Add(o Hittable) {
	hl.Objects = append(hl.Objects, o)
}

// Hit iterates over all objects and returns the closest intersection within rayT.
func (hl HittableList) Hit(r Ray, rayT Interval, rec *HitRecord) bool {
	var tempRec HitRecord
	hitAnything := false
	closest := rayT.Max

	for _, obj := range hl.Objects {
		if obj.Hit(r, Interval{Min: rayT.Min, Max: closest}, &tempRec) {
			hitAnything = true
			closest = tempRec.T
			*rec = tempRec
		}
	}
	return hitAnything
}

// --- Sphere ---

// Sphere is a solid sphere with a centre, radius, and material.
type Sphere struct {
	Center Point3
	Radius float64
	Mat    Material
}

// Hit tests whether ray r intersects this sphere within rayT.
func (s Sphere) Hit(r Ray, rayT Interval, rec *HitRecord) bool {
	oc := s.Center.Sub(r.Origin)
	a := r.Direction.LengthSquared()
	h := Dot(r.Direction, oc) // half-b coefficient (simplified quadratic)
	c := oc.LengthSquared() - s.Radius*s.Radius

	discriminant := h*h - a*c
	if discriminant < 0 {
		return false
	}
	sqrtd := math.Sqrt(discriminant)

	// Find the nearest root that lies within the acceptable interval.
	root := (h - sqrtd) / a
	if !rayT.Surrounds(root) {
		root = (h + sqrtd) / a
		if !rayT.Surrounds(root) {
			return false
		}
	}

	rec.T = root
	rec.P = r.At(root)
	outwardNormal := rec.P.Sub(s.Center).Div(s.Radius)
	rec.SetFaceNormal(r, outwardNormal)
	rec.Mat = s.Mat
	return true
}

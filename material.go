package main

import "math/rand"

// Material decides how a ray interacts with a surface.
// Scatter returns:
//   - attenuation: the colour tint the surface applies to the bounced ray.
//   - scattered:   the new ray to trace.
//   - ok:          false if the ray was fully absorbed (return black).
type Material interface {
	Scatter(rIn Ray, rec *HitRecord) (attenuation Color, scattered Ray, ok bool)
}

// --- Lambertian (diffuse / matte) ---

// Lambertian scatters light in random directions biased toward the surface
// normal (true Lambertian distribution). Perfect for chalk, plaster, fabric.
type Lambertian struct {
	Albedo Color
}

func (l Lambertian) Scatter(rIn Ray, rec *HitRecord) (Color, Ray, bool) {
	dir := rec.Normal.Add(RandomUnitVector())
	// Guard against the degenerate case where the random unit vector exactly
	// cancels the normal, giving a zero scatter direction.
	if dir.NearZero() {
		dir = rec.Normal
	}
	return l.Albedo, Ray{Origin: rec.P, Direction: dir}, true
}

// --- Metal (reflective) ---

// Metal reflects rays with optional fuzz (surface roughness).
// Fuzz = 0 is a perfect mirror; Fuzz approaching 1 gives a very blurry reflection.
type Metal struct {
	Albedo Color
	Fuzz   float64
}

func (m Metal) Scatter(rIn Ray, rec *HitRecord) (Color, Ray, bool) {
	reflected := Reflect(rIn.Direction.Unit(), rec.Normal)
	fuzz := m.Fuzz
	if fuzz > 1 {
		fuzz = 1
	}
	reflected = reflected.Add(RandomUnitVector().Scale(fuzz))
	scattered := Ray{Origin: rec.P, Direction: reflected}
	// Discard rays that scatter through the surface (below the tangent plane).
	ok := Dot(scattered.Direction, rec.Normal) > 0
	return m.Albedo, scattered, ok
}

// --- Dielectric (glass / water / diamond) ---

// Dielectric refracts light according to Snell's law, and reflects it at
// grazing angles using the Schlick approximation.
type Dielectric struct {
	RefractionIndex float64 // air ≈ 1.0, glass ≈ 1.5, diamond ≈ 2.4
}

// schlick approximates the Fresnel reflectance as a function of angle.
// At grazing angles (cosine → 0), nearly all light is reflected.
func schlick(cosine, refIdx float64) float64 {
	r0 := (1 - refIdx) / (1 + refIdx)
	r0 *= r0
	c := 1 - cosine
	return r0 + (1-r0)*c*c*c*c*c
}

func (d Dielectric) Scatter(rIn Ray, rec *HitRecord) (Color, Ray, bool) {
	attenuation := NewVec3(1, 1, 1) // glass doesn't absorb light

	// When entering the glass (front face), flip the ratio.
	ri := d.RefractionIndex
	if rec.FrontFace {
		ri = 1.0 / d.RefractionIndex
	}

	unitDir := rIn.Direction.Unit()
	cosTheta := Dot(unitDir.Neg(), rec.Normal)
	if cosTheta > 1.0 {
		cosTheta = 1.0
	}

	// Determine if Snell's law has a solution (total internal reflection check).
	sin2Theta := 1.0 - cosTheta*cosTheta
	if sin2Theta < 0 {
		sin2Theta = 0
	}
	cannotRefract := ri*ri*sin2Theta > 1.0

	var direction Vec3
	if cannotRefract || schlick(cosTheta, ri) > rand.Float64() {
		direction = Reflect(unitDir, rec.Normal)
	} else {
		direction = Refract(unitDir, rec.Normal, ri)
	}
	return attenuation, Ray{Origin: rec.P, Direction: direction}, true
}

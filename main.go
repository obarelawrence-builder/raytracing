package main

import (
	"math/rand"
	"os"
)

func main() {
	world := randomScene()

	// Output filename — default "output.ppm", override with first CLI arg:
	//   .\rt.exe final.ppm
	output := "output.ppm"
	if len(os.Args) > 1 {
		output = os.Args[1]
	}

	cam := Camera{
		AspectRatio:     16.0 / 9.0,
		ImageWidth:      400,   // Change to 1200 for the full cover render
		SamplesPerPixel: 50,    // Change to 500 for the full cover render
		MaxDepth:        50,
		VFov:            20,
		LookFrom:        NewVec3(13, 2, 3),
		LookAt:          NewVec3(0, 0, 0),
		VUp:             NewVec3(0, 1, 0),
		DefocusAngle:    0.6,
		FocusDist:       10.0,
	}
	cam.Render(world, output)
}

// randomScene builds the iconic Ray Tracing in One Weekend cover scene:
// three large feature spheres (glass, matte, metal) surrounded by hundreds
// of small randomly-placed, randomly-materialised spheres.
func randomScene() HittableList {
	var world HittableList

	// Vast grey ground sphere (radius 1000 makes the top look flat).
	ground := Lambertian{Albedo: NewVec3(0.5, 0.5, 0.5)}
	world.Add(Sphere{Center: NewVec3(0, -1000, 0), Radius: 1000, Mat: ground})

	// Grid of small random spheres.
	for a := -11; a < 11; a++ {
		for b := -11; b < 11; b++ {
			chooseMat := rand.Float64()
			center := NewVec3(
				float64(a)+0.9*rand.Float64(),
				0.2,
				float64(b)+0.9*rand.Float64(),
			)
			// Skip spheres that would overlap the three big feature spheres.
			if center.Sub(NewVec3(4, 0.2, 0)).Length() <= 0.9 {
				continue
			}

			var mat Material
			switch {
			case chooseMat < 0.8: // 80 % diffuse
				albedo := RandomVec3().Mul(RandomVec3())
				mat = Lambertian{Albedo: albedo}
			case chooseMat < 0.95: // 15 % metal
				albedo := RandomVec3Range(0.5, 1)
				fuzz := rand.Float64() * 0.5
				mat = Metal{Albedo: albedo, Fuzz: fuzz}
			default: // 5 % glass
				mat = Dielectric{RefractionIndex: 1.5}
			}
			world.Add(Sphere{Center: center, Radius: 0.2, Mat: mat})
		}
	}

	// Three large feature spheres.
	world.Add(Sphere{
		Center: NewVec3(0, 1, 0), Radius: 1.0,
		Mat: Dielectric{RefractionIndex: 1.5},
	})
	world.Add(Sphere{
		Center: NewVec3(-4, 1, 0), Radius: 1.0,
		Mat: Lambertian{Albedo: NewVec3(0.4, 0.2, 0.1)},
	})
	world.Add(Sphere{
		Center: NewVec3(4, 1, 0), Radius: 1.0,
		Mat: Metal{Albedo: NewVec3(0.7, 0.6, 0.5), Fuzz: 0.0},
	})

	return world
}

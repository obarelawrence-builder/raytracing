func randomScene() HittableList {
    var world HittableList
    ground := Lambertian{Albedo: NewVec3(0.5, 0.5, 0.5)}
    world.Add(Sphere{Center: NewVec3(0, -1000, 0), Radius: 1000, Mat: ground})

    for a := -11; a < 11; a++ {
        for b := -11; b < 11; b++ {
            chooseMat := rand.Float64()
            center := NewVec3(
                float64(a) + 0.9*rand.Float64(),
                0.2,
                float64(b) + 0.9*rand.Float64(),
            )
            if center.Sub(NewVec3(4, 0.2, 0)).Length() <= 0.9 {
                continue
            }
            var mat Material
            switch {
            case chooseMat < 0.8:  // diffuse
                albedo := RandomVec3().Mul(RandomVec3())
                mat = Lambertian{Albedo: albedo}
            case chooseMat < 0.95: // metal
                albedo := RandomVec3Range(0.5, 1)
                fuzz   := rand.Float64() * 0.5
                mat = Metal{Albedo: albedo, Fuzz: fuzz}
            default:                // glass
                mat = Dielectric{RefractionIndex: 1.5}
            }
            world.Add(Sphere{Center: center, Radius: 0.2, Mat: mat})
        }
    }

    world.Add(Sphere{Center: NewVec3(0, 1, 0),  Radius: 1.0, Mat: Dielectric{RefractionIndex: 1.5}})
    world.Add(Sphere{Center: NewVec3(-4, 1, 0), Radius: 1.0, Mat: Lambertian{Albedo: NewVec3(0.4, 0.2, 0.1)}})
    world.Add(Sphere{Center: NewVec3(4, 1, 0),  Radius: 1.0, Mat: Metal{Albedo: NewVec3(0.7, 0.6, 0.5), Fuzz: 0.0}})

    return world
}

func main() {
    world := randomScene()
    cam := Camera{
        AspectRatio:     16.0 / 9.0,
        ImageWidth:      1200,
        SamplesPerPixel: 500,
        MaxDepth:        50,
        VFov:            20,
        LookFrom:        NewVec3(13, 2, 3),
        LookAt:          NewVec3(0, 0, 0),
        VUp:             NewVec3(0, 1, 0),
        DefocusAngle:    0.6,
        FocusDist:       10.0,
    }
    cam.Render(world)
}

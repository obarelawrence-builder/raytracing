package main

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
)

// Camera holds all rendering configuration and drives the render loop.
// Set the exported fields, then call Render(world).
type Camera struct {
	// --- Image ---
	AspectRatio     float64 // width / height, e.g. 16.0/9.0
	ImageWidth      int     // horizontal pixel count
	SamplesPerPixel int     // anti-aliasing samples per pixel
	MaxDepth        int     // maximum ray-bounce depth

	// --- Lens ---
	VFov         float64 // vertical field of view in degrees
	LookFrom     Point3  // camera position
	LookAt       Point3  // point the camera is aimed at
	VUp          Vec3    // world "up" direction (usually 0,1,0)
	DefocusAngle float64 // cone angle of the defocus blur (degrees); 0 = pinhole
	FocusDist    float64 // distance to the plane of perfect focus

	// --- Computed by initialize() ---
	imageHeight  int
	center       Point3
	pixel00      Point3  // centre of top-left pixel
	pixelDeltaU  Vec3    // horizontal step between pixel centres
	pixelDeltaV  Vec3    // vertical step between pixel centres
	defocusDiskU Vec3    // horizontal defocus disk radius vector
	defocusDiskV Vec3    // vertical defocus disk radius vector
	pixelScale   float64 // 1 / SamplesPerPixel
}

// initialize pre-computes all derived camera geometry.
func (c *Camera) initialize() {
	// Image height (minimum 1)
	c.imageHeight = int(float64(c.ImageWidth) / c.AspectRatio)
	if c.imageHeight < 1 {
		c.imageHeight = 1
	}
	c.pixelScale = 1.0 / float64(c.SamplesPerPixel)
	c.center = c.LookFrom

	// Viewport geometry derived from vfov and focus distance
	theta := c.VFov * math.Pi / 180.0
	h := math.Tan(theta / 2.0)
	viewportHeight := 2.0 * h * c.FocusDist
	viewportWidth := viewportHeight * float64(c.ImageWidth) / float64(c.imageHeight)

	// Camera basis vectors
	w := c.LookFrom.Sub(c.LookAt).Unit() // points away from the scene
	u := Cross(c.VUp, w).Unit()          // camera-right
	v := Cross(w, u)                     // camera-up

	// Vectors along viewport edges
	viewportU := u.Scale(viewportWidth)
	viewportV := v.Neg().Scale(viewportHeight) // negative: pixel row goes down

	c.pixelDeltaU = viewportU.Div(float64(c.ImageWidth))
	c.pixelDeltaV = viewportV.Div(float64(c.imageHeight))

	// Top-left corner of the viewport, then nudge to pixel[0,0] centre
	viewportUL := c.center.
		Sub(w.Scale(c.FocusDist)).
		Sub(viewportU.Div(2)).
		Sub(viewportV.Div(2))
	c.pixel00 = viewportUL.Add(c.pixelDeltaU.Add(c.pixelDeltaV).Scale(0.5))

	// Defocus disk basis vectors
	defocusRadius := c.FocusDist * math.Tan((c.DefocusAngle/2)*math.Pi/180)
	c.defocusDiskU = u.Scale(defocusRadius)
	c.defocusDiskV = v.Scale(defocusRadius)
}

// getRay returns one sample ray for pixel (i, j) with a random sub-pixel offset
// and (if DefocusAngle > 0) a random origin on the defocus disk.
func (c *Camera) getRay(i, j int) Ray {
	offsetX := rand.Float64() - 0.5
	offsetY := rand.Float64() - 0.5
	pixelSample := c.pixel00.
		Add(c.pixelDeltaU.Scale(float64(i) + offsetX)).
		Add(c.pixelDeltaV.Scale(float64(j) + offsetY))

	var origin Point3
	if c.DefocusAngle <= 0 {
		origin = c.center // pinhole: all rays from same point
	} else {
		p := RandomInUnitDisk()
		origin = c.center.
			Add(c.defocusDiskU.Scale(p.X)).
			Add(c.defocusDiskV.Scale(p.Y))
	}
	return Ray{Origin: origin, Direction: pixelSample.Sub(origin)}
}

// rayColor traces ray r recursively up to `depth` bounces.
// Returns the accumulated colour seen along the ray.
func (c *Camera) rayColor(r Ray, depth int, world Hittable) Color {
	// Exceeded bounce limit — no more light gathered.
	if depth <= 0 {
		return NewVec3(0, 0, 0)
	}

	var rec HitRecord
	// tMin = 0.001 avoids "shadow acne": floating-point re-hits at t ≈ 0.
	if world.Hit(r, Interval{Min: 0.001, Max: math.MaxFloat64}, &rec) {
		attenuation, scattered, ok := rec.Mat.Scatter(r, &rec)
		if ok {
			return attenuation.Mul(c.rayColor(scattered, depth-1, world))
		}
		return NewVec3(0, 0, 0) // absorbed
	}

	// Sky gradient: white at the bottom, blue at the top.
	unitDir := r.Direction.Unit()
	a := 0.5 * (unitDir.Y + 1.0)
	return NewVec3(1, 1, 1).Scale(1.0 - a).Add(NewVec3(0.5, 0.7, 1.0).Scale(a))
}

// Render runs the full render and writes a PPM image directly to outFile.
// Writing to a file directly (instead of stdout) avoids PowerShell's UTF-16
// redirection bug that produces corrupt / 1×1 pixel PPM files.
// Uses one goroutine per CPU core for parallelism.
func (c *Camera) Render(world Hittable, outFile string) {
	c.initialize()

	// Open the output file for writing.
	f, err := os.Create(outFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot create %s: %v\n", outFile, err)
		os.Exit(1)
	}
	defer f.Close()
	w := bufio.NewWriterSize(f, 1<<20) // 1 MB write buffer for speed

	// Allocate flat pixel buffer (row-major).
	pixels := make([]Color, c.ImageWidth*c.imageHeight)

	workers := runtime.NumCPU()
	rowsPerWorker := (c.imageHeight + workers - 1) / workers

	var wg sync.WaitGroup
	var scanlinesDone atomic.Int64

	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			startJ := workerID * rowsPerWorker
			endJ := startJ + rowsPerWorker
			if endJ > c.imageHeight {
				endJ = c.imageHeight
			}
			for j := startJ; j < endJ; j++ {
				for i := 0; i < c.ImageWidth; i++ {
					pixel := NewVec3(0, 0, 0)
					for s := 0; s < c.SamplesPerPixel; s++ {
						r := c.getRay(i, j)
						pixel = pixel.Add(c.rayColor(r, c.MaxDepth, world))
					}
					pixels[j*c.ImageWidth+i] = pixel.Scale(c.pixelScale)
				}
				scanlinesDone.Add(1)
				fmt.Fprintf(os.Stderr, "\rScanlines remaining: %d   ",
					c.imageHeight-int(scanlinesDone.Load()))
			}
		}(worker)
	}

	wg.Wait()
	fmt.Fprintln(os.Stderr, "\nWriting image...")

	// Write PPM header then all pixels directly into the file.
	fmt.Fprintf(w, "P3\n%d %d\n255\n", c.ImageWidth, c.imageHeight)
	for _, px := range pixels {
		px.WriteColor(w)
	}
	w.Flush()
	fmt.Fprintf(os.Stderr, "Done. Output: %s\n", outFile)
}

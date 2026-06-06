# Ray Tracing in One Weekend - Go Implementation

This is a complete Go implementation of Peter Shirley's famous "Ray Tracing in One Weekend" tutorial.
It implements a fully functional path tracer from scratch with no external graphics libraries.

## Features

- Custom `Vec3` math library
- Rays, Spheres, and Hittables
- Materials: 
  - Lambertian (diffuse)
  - Metal (reflective)
  - Dielectric (glass/refractive)
- Positionable camera with Defocus Blur (Depth of Field)
- Antialiasing (multisampling)
- **Multi-core parallel rendering**: Automatically uses all CPU cores for significantly faster renders.
- Direct-to-file PPM output (bypasses Windows terminal encoding bugs).
- Uses Go 1.22+ `math/rand/v2` for guaranteed non-deterministic random scenes.

## Building and Running

Ensure you have Go installed.

1. Build the executable:
   ```bash
   go build -o rt.exe .
   ```
2. Run the renderer. By default, it outputs a quick preview (400x225, 50 samples per pixel) to `output.ppm`:
   ```bash
   ./rt.exe
   ```
3. Or specify a custom output filename:
   ```bash
   ./rt.exe my_render.ppm
   ```

To render the high-quality final cover image (1200x675, 500 samples per pixel), edit `main.go` and update `ImageWidth` to `1200` and `SamplesPerPixel` to `500`, then recompile and run.

## Viewing PPM files
The output is a standard `.ppm` (Portable Pixmap) file. You can view this using tools like [IrfanView](https://www.irfanview.com/), [GIMP](https://www.gimp.org/), or drag and drop it into an online PPM viewer like [ppmviewer.com](https://ppmviewer.com).

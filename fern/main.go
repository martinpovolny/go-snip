package main

import (
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// Canvas dimensions. The fern's mathematical coordinate space spans roughly
// x ∈ [-3, 3] and y ∈ [0, 10], so xscale=w/6 maps that x-range to the full
// width, and yscale=h/11 maps the y-range to the full height.
// hw is the horizontal midpoint used to centre x=0 on screen.
const (
	w      = 1200
	h      = 1500
	hw     = w / 2  // x=0 in fern space maps to the horizontal centre
	xscale = w / 6  // pixels per unit in x
	yscale = h / 11 // pixels per unit in y
)

// plotpoint maps a fern-space coordinate (x, y) to a screen pixel and
// brightens its green channel.
//
// Instead of simply painting every hit pixel the same colour, we accumulate
// brightness: each additional visit moves the green value 1/8 of the way
// closer to 255, then adds 33. Starting from 0:
//   visit 1 → 33, visit 2 → 62, visit 3 → 87, …  asymptotically → 255
//
// This encodes local density: frequently-visited pixels (the thick central
// stem) end up near white-green, while lightly-visited tips stay dark green.
// Once a pixel reaches 255 it is saturated and further visits are skipped.
func plotpoint(img *image.RGBA, x, y float64) {
	// Translate from fern space to pixel space.
	// x=0 maps to hw (centre); y=0 maps to the bottom row (h - 0 = h, but
	// clipped), and y increases upward in fern space, downward in pixel space.
	xscr := hw + int(x*xscale)
	yscr := h - int(y*yscale)
	if xscr < 0 || xscr >= w || yscr < 0 || yscr >= h {
		return
	}
	cgreen := int(img.RGBAAt(xscr, yscr).G)
	if cgreen < 255 {
		// g' = 7/8·g + 33  — integer arithmetic matches the original BASIC source
		cgreen = 7*cgreen/8 + 33
		if cgreen > 255 {
			cgreen = 255
		}
		img.SetRGBA(xscr, yscr, color.RGBA{0, uint8(cgreen), 0, 255})
	}
}

func main() {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Fill canvas with opaque black so the green accumulation starts from zero.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetRGBA(x, y, color.RGBA{0, 0, 0, 255})
		}
	}

	// The Barnsley fern is an Iterated Function System (IFS): starting from
	// an arbitrary point, we repeatedly apply one of four affine transforms
	// chosen at random according to fixed probabilities. The sequence of
	// points converges to the fern's attractor regardless of the starting
	// point, so (0, 0) is fine.
	x, y := 0.0, 0.0

	// w*h iterations give enough points to densely fill the 1200×1500 canvas.
	for t := 0; t < w*h; t++ {
		r := rand.Float64() // uniform in [0, 1)
		var xn float64      // next x is computed first so we don't clobber x before y uses it

		switch {
		case r < 0.01: // 1% — stem: collapses everything toward the base of the fern
			xn = 0.0
			y = 0.16 * y

		case r < 0.86: // 85% — main frond: slight rotation + upward translation; produces the bulk of the leaf
			xn = 0.85*x + 0.04*y
			y = -0.04*x + 0.85*y + 1.6

		case r < 0.93: // 7% — right-hand leaflet: rotates ~49° and scales down
			xn = 0.2*x - 0.26*y
			y = 0.23*x + 0.22*y + 1.6

		default: // 7% — left-hand leaflet: reflects and rotates ~120°, scales down
			xn = -0.15*x + 0.28*y
			y = 0.26*x + 0.24*y + 0.44
		}

		x = xn
		plotpoint(img, x, y)
	}

	// Label the image using the x/image basic bitmap font.
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.RGBA{0, 100, 0, 255}), // dark green text
		Face: basicfont.Face7x13,
		Dot:  fixed.P(80, 50), // top-left origin of the text baseline
	}
	d.DrawString("Barnsley Fern")

	f, err := os.Create("barnsley_fern.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
}

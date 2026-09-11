// Package appicon exposes the icons used across the whole application:
//
//   - Draw: the "regular" full-color icon —glass ring, monitor, and an
//     "ON" toggle with a glow— used in the Windows/Linux tray, the .ico,
//     the macOS Dock .icns, and the README's assets/icon.png. It's drawn
//     as an SVG in assets/icon.svg and rasterized once, at high
//     resolution, into icon_master.png (see tools/render-icon-master); Draw
//     only rescales that embedded PNG to the requested size, with no
//     external dependencies at either build or run time.
//   - DrawTemplate: a monochrome silhouette (black on transparent, with the
//     screen and the toggle "cut out" as a hole) meant for the macOS menu
//     bar icon via systray.SetTemplateIcon, which the system automatically
//     recolors to white/black depending on the menu bar theme.
package appicon

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"image/png"
	"math"
	"sync"
)

//go:embed icon_master.png
var masterPNG []byte

var (
	masterOnce sync.Once
	master     *image.NRGBA
)

func loadMaster() *image.NRGBA {
	masterOnce.Do(func() {
		img, err := png.Decode(bytes.NewReader(masterPNG))
		if err != nil {
			panic("appicon: could not decode icon_master.png: " + err.Error())
		}
		master = toNRGBA(img)
	})
	return master
}

func toNRGBA(img image.Image) *image.NRGBA {
	if n, ok := img.(*image.NRGBA); ok {
		return n
	}
	b := img.Bounds()
	out := image.NewNRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			out.Set(x, y, img.At(x, y))
		}
	}
	return out
}

// Draw returns the full-color icon on a square canvas of `size` pixels,
// rescaled with area averaging (correct under premultiplied alpha) from
// the embedded master.
func Draw(size int) *image.NRGBA {
	m := loadMaster()
	if size == m.Bounds().Dx() {
		return m
	}
	return resizeNRGBA(m, size)
}

// resizeNRGBA rescales `src` to a square canvas of `size` pixels by
// averaging, for each output pixel, the area it corresponds to in the
// source image (with fractional weight at the edges) — a quality
// downscale with no dependency on any external library. Color is averaged
// in premultiplied space so transparent edges don't get muddied.
func resizeNRGBA(src *image.NRGBA, size int) *image.NRGBA {
	sb := src.Bounds()
	sw, sh := sb.Dx(), sb.Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, size, size))

	scaleX := float64(sw) / float64(size)
	scaleY := float64(sh) / float64(size)

	for y := 0; y < size; y++ {
		srcY0 := float64(y) * scaleY
		srcY1 := srcY0 + scaleY
		y0, y1 := int(math.Floor(srcY0)), int(math.Ceil(srcY1))

		for x := 0; x < size; x++ {
			srcX0 := float64(x) * scaleX
			srcX1 := srcX0 + scaleX
			x0, x1 := int(math.Floor(srcX0)), int(math.Ceil(srcX1))

			var rs, gs, bs, as, wsum float64
			for sy := y0; sy < y1; sy++ {
				wy := overlap(float64(sy), float64(sy+1), srcY0, srcY1)
				if wy <= 0 {
					continue
				}
				for sx := x0; sx < x1; sx++ {
					wx := overlap(float64(sx), float64(sx+1), srcX0, srcX1)
					if wx <= 0 {
						continue
					}
					w := wx * wy
					c := src.NRGBAAt(sb.Min.X+sx, sb.Min.Y+sy)
					a := float64(c.A)
					rs += float64(c.R) * a * w
					gs += float64(c.G) * a * w
					bs += float64(c.B) * a * w
					as += a * w
					wsum += w
				}
			}

			var r, g, b, a uint8
			if as > 0 {
				r = uint8(rs / as)
				g = uint8(gs / as)
				b = uint8(bs / as)
			}
			if wsum > 0 {
				a = uint8(as / wsum)
			}
			dst.SetNRGBA(x, y, color.NRGBA{R: r, G: g, B: b, A: a})
		}
	}
	return dst
}

func overlap(a0, a1, b0, b1 float64) float64 {
	lo := math.Max(a0, b0)
	hi := math.Min(a1, b1)
	if hi <= lo {
		return 0
	}
	return hi - lo
}

// ---- "Template" silhouette for the macOS menu bar ----

// DrawTemplate generates, with the same supersampled antialiasing as the
// design above, a monochrome silhouette of the same glyph (monitor +
// toggle): solid black where the bezel/stand is, and transparent where
// the screen and the toggle knob would be. macOS recolors it on its own
// depending on the menu bar theme.
func DrawTemplate(size int) *image.NRGBA {
	const superSample = 4
	hi := size * superSample
	big := image.NewNRGBA(image.Rect(0, 0, hi, hi))
	paintTemplate(big, float64(hi))
	return downsample(big, size, superSample)
}

func downsample(src *image.NRGBA, size, factor int) *image.NRGBA {
	out := image.NewNRGBA(image.Rect(0, 0, size, size))
	n := factor * factor
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			var rs, gs, bs, as int
			for dy := 0; dy < factor; dy++ {
				for dx := 0; dx < factor; dx++ {
					c := src.NRGBAAt(x*factor+dx, y*factor+dy)
					a := int(c.A)
					rs += int(c.R) * a
					gs += int(c.G) * a
					bs += int(c.B) * a
					as += a
				}
			}
			var r, g, b uint8
			if as > 0 {
				r = uint8(rs / as)
				g = uint8(gs / as)
				b = uint8(bs / as)
			}
			out.SetNRGBA(x, y, color.NRGBA{R: r, G: g, B: b, A: uint8(as / n)})
		}
	}
	return out
}

var (
	colBlack       = color.NRGBA{A: 0xff}
	colTransparent = color.NRGBA{}
)

type rect struct{ x0, y0, x1, y1 float64 }

func (r rect) contains(x, y float64) bool {
	return x >= r.x0 && x <= r.x1 && y >= r.y0 && y <= r.y1
}

// roundedContains approximates a rectangle with rounded corners of radius r.
func (r rect) roundedContains(x, y, radius float64) bool {
	if x < r.x0 || x > r.x1 || y < r.y0 || y > r.y1 {
		return false
	}
	cx, cy := x, y
	switch {
	case cx < r.x0+radius && cy < r.y0+radius:
		return dist(cx, cy, r.x0+radius, r.y0+radius) <= radius
	case cx > r.x1-radius && cy < r.y0+radius:
		return dist(cx, cy, r.x1-radius, r.y0+radius) <= radius
	case cx < r.x0+radius && cy > r.y1-radius:
		return dist(cx, cy, r.x0+radius, r.y1-radius) <= radius
	case cx > r.x1-radius && cy > r.y1-radius:
		return dist(cx, cy, r.x1-radius, r.y1-radius) <= radius
	default:
		return true
	}
}

func dist(x0, y0, x1, y1 float64) float64 {
	dx, dy := x0-x1, y0-y1
	return math.Sqrt(dx*dx + dy*dy)
}

// paintTemplate draws the monitor with a simplified "ON" toggle in solid
// black on transparent: the screen and the toggle knob are cut out
// (transparent) so the glyph reads clearly at the 18-22pt of a menu bar.
func paintTemplate(img *image.NRGBA, s float64) {
	monitor := rect{x0: 0.08 * s, y0: 0.14 * s, x1: 0.92 * s, y1: 0.72 * s}
	screenInset := 0.05 * s
	screen := rect{x0: monitor.x0 + screenInset, y0: monitor.y0 + screenInset, x1: monitor.x1 - screenInset, y1: monitor.y1 - screenInset}
	radius := 0.08 * s

	standTop := rect{x0: s*0.47 - 0.015*s, y0: monitor.y1, x1: s*0.47 + 0.085*s, y1: monitor.y1 + 0.06*s}
	standBase := rect{x0: s * 0.30, y0: standTop.y1, x1: s * 0.70, y1: standTop.y1 + 0.045*s}

	screenW := screen.x1 - screen.x0
	screenH := screen.y1 - screen.y0
	pillH := screenH * 0.42
	pillW := screenW * 0.62
	pillCx := screen.x0 + screenW/2
	pillCy := screen.y0 + screenH/2
	pill := rect{x0: pillCx - pillW/2, y0: pillCy - pillH/2, x1: pillCx + pillW/2, y1: pillCy + pillH/2}
	pillRadius := pillH / 2
	knobR := pillH * 0.36
	knobCx := pill.x1 - pillRadius
	knobCy := pillCy

	size := int(s)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			px, py := float64(x)+0.5, float64(y)+0.5

			switch {
			case standBase.contains(px, py), standTop.contains(px, py), monitor.roundedContains(px, py, radius):
				img.SetNRGBA(x, y, colBlack)
			}
			if screen.roundedContains(px, py, radius*0.6) {
				img.SetNRGBA(x, y, colTransparent)
				if pill.roundedContains(px, py, pillRadius) {
					img.SetNRGBA(x, y, colBlack)
				}
				if dist(px, py, knobCx, knobCy) <= knobR {
					img.SetNRGBA(x, y, colTransparent)
				}
			}
		}
	}
}

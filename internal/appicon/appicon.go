// Package appicon expone los íconos que usa toda la aplicación:
//
//   - Draw: el ícono "regular" a todo color —anillo de cristal, monitor y
//     toggle "ON" con brillo— usado en la bandeja de Windows/Linux, el .ico,
//     el .icns del Dock de macOS y assets/icon.png del README. Se dibuja
//     como SVG en assets/icon.svg y se rasteriza una vez, en alta
//     resolución, a icon_master.png (ver tools/render-icon-master); Draw
//     solo reescala ese PNG embebido al tamaño pedido, sin dependencias
//     externas ni en tiempo de build ni de ejecución.
//   - DrawTemplate: una silueta monocroma (negro sobre transparente, con la
//     pantalla y el toggle "recortados" como hueco) pensada para el ícono de
//     la barra de menú de macOS vía systray.SetTemplateIcon, que el sistema
//     recolorea automáticamente a blanco/negro según el tema de la barra.
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
			panic("appicon: no se pudo decodificar icon_master.png: " + err.Error())
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

// Draw devuelve el ícono a todo color en un canvas cuadrado de `size`
// píxeles, reescalado con promediado de área (correcto en alfa
// premultiplicado) a partir del master embebido.
func Draw(size int) *image.NRGBA {
	m := loadMaster()
	if size == m.Bounds().Dx() {
		return m
	}
	return resizeNRGBA(m, size)
}

// resizeNRGBA reescala `src` a un canvas cuadrado de `size` píxeles
// promediando, por cada píxel de salida, el área que le corresponde en la
// imagen de origen (con peso fraccional en los bordes) — un downscale de
// calidad sin depender de ninguna librería externa. El color se promedia en
// espacio premultiplicado para no ensuciar los bordes transparentes.
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

// ---- Silueta "template" para la barra de menú de macOS ----

// DrawTemplate genera, con el mismo antialiasing por sobremuestreo que el
// diseño anterior, una silueta monocroma del mismo glifo (monitor + toggle):
// negro sólido donde hay bisel/base, y transparente donde estaría la
// pantalla y la perilla del toggle. macOS la recolorea sola según el tema
// de la barra de menú.
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

// roundedContains aproxima un rectángulo con esquinas redondeadas de radio r.
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

// paintTemplate dibuja el monitor con un toggle "ON" simplificado en negro
// sólido sobre transparente: la pantalla y la perilla del toggle quedan
// recortadas (transparentes) para que el glifo se lea con claridad a los
// 18-22pt de una barra de menú.
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

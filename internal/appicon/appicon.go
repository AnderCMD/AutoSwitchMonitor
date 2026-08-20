// Package appicon dibuja, en memoria y sin assets externos, el ícono que
// usa toda la aplicación: el ícono de la bandeja del sistema en tiempo de
// ejecución y (vía tools/gen-icon) los archivos assets/icon.png /
// assets/icon.ico que se embeben en el ejecutable de Windows.
//
// Mantener un solo dibujo compartido evita que el ícono de la bandeja y el
// ícono del .exe/.app se vean distintos entre sí.
package appicon

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"math"
)

var (
	colBezel  = color.NRGBA{R: 0x1e, G: 0x29, B: 0x3b, A: 0xff} // slate-800
	colScreen = color.NRGBA{R: 0x3b, G: 0x82, B: 0xf6, A: 0xff} // blue-500
	colBadge  = color.NRGBA{R: 0xf5, G: 0x9e, B: 0x0b, A: 0xff} // amber-500
	colArrow  = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
)

// Draw genera el ícono en un canvas cuadrado de `size` píxeles: un monitor
// con su base, y una insignia circular con dos flechas curvas ("swap") que
// representan el cambio de entrada.
func Draw(size int) *image.NRGBA {
	s := float64(size)
	img := image.NewNRGBA(image.Rect(0, 0, size, size))

	monitor := rect{x0: 0.16 * s, y0: 0.10 * s, x1: 0.84 * s, y1: 0.62 * s}
	screenInset := 0.035 * s
	screen := rect{x0: monitor.x0 + screenInset, y0: monitor.y0 + screenInset, x1: monitor.x1 - screenInset, y1: monitor.y1 - screenInset}
	radius := 0.045 * s

	standTop := rect{x0: s*0.47, y0: monitor.y1, x1: s*0.53, y1: monitor.y1 + 0.09*s}
	standBase := rect{x0: s * 0.36, y0: standTop.y1, x1: s * 0.64, y1: standTop.y1 + 0.045 * s}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			px, py := float64(x)+0.5, float64(y)+0.5

			switch {
			case standBase.contains(px, py):
				img.SetNRGBA(x, y, colBezel)
			case standTop.contains(px, py):
				img.SetNRGBA(x, y, colBezel)
			case monitor.roundedContains(px, py, radius):
				img.SetNRGBA(x, y, colBezel)
			}
			if screen.roundedContains(px, py, radius*0.6) {
				img.SetNRGBA(x, y, colScreen)
			}
		}
	}

	badgeCx, badgeCy, badgeR := s*0.755, s*0.755, s*0.225
	drawBadge(img, badgeCx, badgeCy, badgeR)

	return img
}

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

// drawBadge pinta un círculo de acento con dos flechas curvas opuestas
// (símbolo de "swap"/ciclo entre dos entradas).
func drawBadge(img *image.NRGBA, cx, cy, r float64) {
	bounds := img.Bounds()
	ringOuter := r * 0.72
	ringInner := r * 0.52
	headR := r * 0.16

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			px, py := float64(x)+0.5, float64(y)+0.5
			d := dist(px, py, cx, cy)
			if d <= r {
				img.SetNRGBA(x, y, colBadge)
			}
		}
	}

	// Dos arcos opuestos de ~150°, cada uno con una punta de flecha.
	drawArc(img, cx, cy, ringInner, ringOuter, -20, 160)
	drawArc(img, cx, cy, ringInner, ringOuter, 160, 340)

	arrowHead(img, cx, cy, (ringInner+ringOuter)/2, 160, headR)
	arrowHead(img, cx, cy, (ringInner+ringOuter)/2, 340, headR)
}

func drawArc(img *image.NRGBA, cx, cy, rInner, rOuter, fromDeg, toDeg float64) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			px, py := float64(x)+0.5, float64(y)+0.5
			d := dist(px, py, cx, cy)
			if d < rInner || d > rOuter {
				continue
			}
			angle := math.Atan2(py-cy, px-cx) * 180 / math.Pi
			if angle < 0 {
				angle += 360
			}
			if angle >= fromDeg && angle <= toDeg {
				img.SetNRGBA(x, y, colArrow)
			}
		}
	}
}

func arrowHead(img *image.NRGBA, cx, cy, ringR, atDeg, size float64) {
	rad := atDeg * math.Pi / 180
	tipX := cx + ringR*math.Cos(rad)
	tipY := cy + ringR*math.Sin(rad)

	// Triángulo apuntando en la dirección tangencial del arco.
	tangent := rad + math.Pi/2
	baseX1 := tipX - size*math.Cos(rad) + size*0.6*math.Cos(tangent)
	baseY1 := tipY - size*math.Sin(rad) + size*0.6*math.Sin(tangent)
	baseX2 := tipX - size*math.Cos(rad) - size*0.6*math.Cos(tangent)
	baseY2 := tipY - size*math.Sin(rad) - size*0.6*math.Sin(tangent)
	tip2X := tipX + size*math.Cos(rad)
	tip2Y := tipY + size*math.Sin(rad)

	minX, maxX := minOf3(tip2X, baseX1, baseX2), maxOf3(tip2X, baseX1, baseX2)
	minY, maxY := minOf3(tip2Y, baseY1, baseY2), maxOf3(tip2Y, baseY1, baseY2)

	for y := int(minY) - 1; y <= int(maxY)+1; y++ {
		for x := int(minX) - 1; x <= int(maxX)+1; x++ {
			if pointInTriangle(float64(x)+0.5, float64(y)+0.5, tip2X, tip2Y, baseX1, baseY1, baseX2, baseY2) {
				img.SetNRGBA(x, y, colArrow)
			}
		}
	}
}

func pointInTriangle(px, py, x1, y1, x2, y2, x3, y3 float64) bool {
	d1 := sign(px, py, x1, y1, x2, y2)
	d2 := sign(px, py, x2, y2, x3, y3)
	d3 := sign(px, py, x3, y3, x1, y1)
	hasNeg := d1 < 0 || d2 < 0 || d3 < 0
	hasPos := d1 > 0 || d2 > 0 || d3 > 0
	return !(hasNeg && hasPos)
}

func sign(px, py, x1, y1, x2, y2 float64) float64 {
	return (px-x2)*(y1-y2) - (x1-x2)*(py-y2)
}

func minOf3(a, b, c float64) float64 { return math.Min(a, math.Min(b, c)) }
func maxOf3(a, b, c float64) float64 { return math.Max(a, math.Max(b, c)) }

// EncodeICO empaqueta una sola imagen NRGBA como un .ico de 32bpp válido.
// Usado tanto por el ícono de bandeja en Windows como por tools/gen-icon
// para generar assets/icon.ico.
func EncodeICO(img *image.NRGBA) []byte {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()

	rowSize := w * 4
	xorSize := rowSize * h
	andRowSize := ((w + 31) / 32) * 4
	andSize := andRowSize * h

	var xor bytes.Buffer
	for y := h - 1; y >= 0; y-- { // filas de abajo hacia arriba
		for x := 0; x < w; x++ {
			c := img.NRGBAAt(x, y)
			xor.WriteByte(c.B)
			xor.WriteByte(c.G)
			xor.WriteByte(c.R)
			xor.WriteByte(c.A)
		}
	}
	and := make([]byte, andSize) // sin recorte: se usa el canal alfa

	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, uint16(0)) // reserved
	binary.Write(&buf, binary.LittleEndian, uint16(1)) // type = icon
	binary.Write(&buf, binary.LittleEndian, uint16(1)) // count

	dim := byte(w)
	if w >= 256 {
		dim = 0
	}
	bytesInRes := uint32(40 + xorSize + andSize)
	imageOffset := uint32(6 + 16)

	buf.WriteByte(dim)
	buf.WriteByte(dim)
	buf.WriteByte(0)
	buf.WriteByte(0)
	binary.Write(&buf, binary.LittleEndian, uint16(1))
	binary.Write(&buf, binary.LittleEndian, uint16(32))
	binary.Write(&buf, binary.LittleEndian, bytesInRes)
	binary.Write(&buf, binary.LittleEndian, imageOffset)

	binary.Write(&buf, binary.LittleEndian, uint32(40))
	binary.Write(&buf, binary.LittleEndian, int32(w))
	binary.Write(&buf, binary.LittleEndian, int32(h*2))
	binary.Write(&buf, binary.LittleEndian, uint16(1))
	binary.Write(&buf, binary.LittleEndian, uint16(32))
	binary.Write(&buf, binary.LittleEndian, uint32(0))
	binary.Write(&buf, binary.LittleEndian, uint32(xorSize))
	binary.Write(&buf, binary.LittleEndian, int32(0))
	binary.Write(&buf, binary.LittleEndian, int32(0))
	binary.Write(&buf, binary.LittleEndian, uint32(0))
	binary.Write(&buf, binary.LittleEndian, uint32(0))

	buf.Write(xor.Bytes())
	buf.Write(and)

	return buf.Bytes()
}

// EncodeMultiICO empaqueta varias resoluciones de la misma imagen en un
// solo .ico (lo que Windows espera para que el ícono se vea nítido en el
// Explorador, la barra de tareas y los accesos directos).
func EncodeMultiICO(images []*image.NRGBA) []byte {
	type entry struct {
		data []byte
		w    int
	}
	entries := make([]entry, 0, len(images))
	for _, img := range images {
		entries = append(entries, entry{data: rawICOImage(img), w: img.Bounds().Dx()})
	}

	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, uint16(0))
	binary.Write(&buf, binary.LittleEndian, uint16(1))
	binary.Write(&buf, binary.LittleEndian, uint16(len(entries)))

	offset := uint32(6 + 16*len(entries))
	for _, e := range entries {
		dim := byte(e.w)
		if e.w >= 256 {
			dim = 0
		}
		buf.WriteByte(dim)
		buf.WriteByte(dim)
		buf.WriteByte(0)
		buf.WriteByte(0)
		binary.Write(&buf, binary.LittleEndian, uint16(1))
		binary.Write(&buf, binary.LittleEndian, uint16(32))
		binary.Write(&buf, binary.LittleEndian, uint32(len(e.data)))
		binary.Write(&buf, binary.LittleEndian, offset)
		offset += uint32(len(e.data))
	}
	for _, e := range entries {
		buf.Write(e.data)
	}
	return buf.Bytes()
}

// rawICOImage devuelve el bloque BITMAPINFOHEADER+XOR+AND de una imagen,
// sin el ICONDIR/ICONDIRENTRY (para usarse dentro de un .ico multi-tamaño).
func rawICOImage(img *image.NRGBA) []byte {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	rowSize := w * 4
	xorSize := rowSize * h
	andRowSize := ((w + 31) / 32) * 4
	andSize := andRowSize * h

	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, uint32(40))
	binary.Write(&buf, binary.LittleEndian, int32(w))
	binary.Write(&buf, binary.LittleEndian, int32(h*2))
	binary.Write(&buf, binary.LittleEndian, uint16(1))
	binary.Write(&buf, binary.LittleEndian, uint16(32))
	binary.Write(&buf, binary.LittleEndian, uint32(0))
	binary.Write(&buf, binary.LittleEndian, uint32(xorSize))
	binary.Write(&buf, binary.LittleEndian, int32(0))
	binary.Write(&buf, binary.LittleEndian, int32(0))
	binary.Write(&buf, binary.LittleEndian, uint32(0))
	binary.Write(&buf, binary.LittleEndian, uint32(0))

	for y := h - 1; y >= 0; y-- {
		for x := 0; x < w; x++ {
			c := img.NRGBAAt(x, y)
			buf.WriteByte(c.B)
			buf.WriteByte(c.G)
			buf.WriteByte(c.R)
			buf.WriteByte(c.A)
		}
	}
	buf.Write(make([]byte, andSize))

	return buf.Bytes()
}

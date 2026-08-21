package appicon

import (
	"bytes"
	"encoding/binary"
	"image"
)

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

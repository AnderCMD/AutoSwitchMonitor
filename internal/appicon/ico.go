package appicon

import (
	"bytes"
	"encoding/binary"
	"image"
)

// EncodeICO packs a single NRGBA image as a valid 32bpp .ico.
// Used both by the Windows tray icon and by tools/gen-icon to generate
// assets/icon.ico.
func EncodeICO(img *image.NRGBA) []byte {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()

	rowSize := w * 4
	xorSize := rowSize * h
	andRowSize := ((w + 31) / 32) * 4
	andSize := andRowSize * h

	var xor bytes.Buffer
	for y := h - 1; y >= 0; y-- { // rows bottom to top
		for x := 0; x < w; x++ {
			c := img.NRGBAAt(x, y)
			xor.WriteByte(c.B)
			xor.WriteByte(c.G)
			xor.WriteByte(c.R)
			xor.WriteByte(c.A)
		}
	}
	and := make([]byte, andSize) // no masking: the alpha channel is used instead

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

// EncodeMultiICO packs several resolutions of the same image into a single
// .ico (what Windows expects so the icon looks sharp in Explorer, the
// taskbar, and shortcuts).
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

// rawICOImage returns the BITMAPINFOHEADER+XOR+AND block of an image,
// without the ICONDIR/ICONDIRENTRY (for use inside a multi-size .ico).
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

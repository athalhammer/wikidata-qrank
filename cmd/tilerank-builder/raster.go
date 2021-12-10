// SPDX-License-Identifier: MIT

package main

import (
	"encoding/binary"
	"fmt"

	"github.com/lanrat/extsort"
)

type Raster struct {
	tile        TileKey
	parent      *Raster
	viewsPerKm2 float32
	pixels      [256 * 256]float32
}

func (r *Raster) Paint(tile TileKey, viewsPerKm2 float32) {
	rZoom, rX, rY := r.tile.ZoomXY()

	// If the to-be-painted tile is smaller than 1 pixel, we scale it
	// to one pixel and reduce the number of views accordingly.
	// We only do this at deep zoom levels, where the area per pixel
	// is nearly uniform despite the distortion of the web mercator
	// projection.
	if zoom := tile.Zoom(); zoom > rZoom+8 {
		viewsPerKm2 /= float32(int32(1 << (2 * (zoom - (rZoom + 8)))))
		tile = tile.ToZoom(rZoom + 8)
	}

	zoom, x, y := tile.ZoomXY()
	deltaZoom := zoom - rZoom
	left := (x - rX<<deltaZoom) << (8 - deltaZoom)
	top := (y - rY<<deltaZoom) << (8 - deltaZoom)
	width := uint32(1 << (8 - deltaZoom))
	// Because our tiles are squares, the height is the same as the width.
	for y := top; y < top+width; y++ {
		for x := left; x < left+width; x++ {
			r.pixels[y<<8+x] += viewsPerKm2
		}
	}
}

func NewRaster(tile TileKey, parent *Raster) *Raster {
	zoom := tile.Zoom()

	// Check that NewRaster() is called for the right parent. This check
	// should never fail, no matter what the input data is. If it does fail,
	// something must be wrong with our logic to construct parent rasters.
	if parent != nil {
		if zoom != parent.tile.Zoom()+1 {
			panic(fmt.Sprintf("NewRaster(%s) with parent.tile=%s", tile, parent.tile))
		}
	} else if zoom != 0 {
		panic(fmt.Sprintf("NewRaster(%s) with parent=<nil>", tile))
	}

	return &Raster{tile: tile, parent: parent}
}

type RasterWriter struct {
	palette map[uint32]uint16 // color -> index
}

func NewRasterWriter() *RasterWriter {
	return &RasterWriter{palette: make(map[uint32]uint16, 65536)}
}

func (w *RasterWriter) Write(r *Raster) {
}

// WriteUniform produces a raster whose pixels all have the same color.
// In a typical output, about 55% of all rasters are uniformly coloreds,
// so we treat them specially as an optimization.
func (w *RasterWriter) WriteUniform(tile TileKey, color uint32) error {
	var t cogTile
	zoom, x, y := tile.ZoomXY()
	t.zoom = zoom
	t.y = x
	t.y = y
	colorIndex, exists := w.palette[color]
	if !exists {
		numColors := len(w.palette)
		if numColors >= 0xffff {
			// If this ever triggers, a fallback would be to read back the
			// already emitted tiles, convert them to non-indexed form,
			// write them out again, and then continue writing. This would
			// be complex to implement, and from the data we’ve seen
			// it’s not necessary because only used about 20K colors
			// are sufficient for the entire world.
			panic("palette full; need to implement fallback")
		}
		colorIndex = uint16(numColors)
		w.palette[color] = colorIndex
	}
	t.uniformColorIndex = colorIndex
	//fmt.Printf("TODO: Send %v to sorting channel\n", t)
	return nil
}

func (w *RasterWriter) Close() error {
	fmt.Printf("len(palette)=%d\n", len(w.palette))
	return nil
}

// cogTile represents a raster tile that will be written into
// a Cloud-Optimized GeoTIFF file. The file format requires
// a specific arrangement of the data, which is different from
// the order in which we’re painting our raster tiles.
type cogTile struct {
	zoom              uint8
	x, y              uint32
	uniformColorIndex uint16
	byteCount         uint32
	offset            uint64
}

// ToBytes serializes a cogTile into a byte array.
func (c cogTile) ToBytes() []byte {
	var buf [1 + 4*binary.MaxVarintLen32 + binary.MaxVarintLen64]byte
	buf[0] = c.zoom
	pos := 1
	pos += binary.PutUvarint(buf[pos:], uint64(c.x))
	pos += binary.PutUvarint(buf[pos:], uint64(c.y))
	pos += binary.PutUvarint(buf[pos:], uint64(c.uniformColorIndex))
	pos += binary.PutUvarint(buf[pos:], uint64(c.byteCount))
	pos += binary.PutUvarint(buf[pos:], c.offset)
	return buf[0:pos]
}

// Function cogTileFromBytes de-serializes a cogTile from a byte slice.
// The result is returned as an extsort.SortType because that is
// needed by the library for external sorting.
func cogTileFromBytes(b []byte) extsort.SortType {
	zoom, pos := b[0], 1
	x, len := binary.Uvarint(b[1:])
	pos += len
	y, len := binary.Uvarint(b[pos:])
	pos += len
	uniformColorIndex, len := binary.Uvarint(b[pos:])
	pos += len
	byteCount, len := binary.Uvarint(b[pos:])
	pos += len
	offset, len := binary.Uvarint(b[pos:])
	pos += len
	return cogTile{
		zoom:              zoom,
		x:                 uint32(x),
		y:                 uint32(y),
		uniformColorIndex: uint16(uniformColorIndex),
		byteCount:         uint32(byteCount),
		offset:            offset,
	}
}

// cogTileLess returns true if the TIFF tag for raster a should come
// before b in the Cloud-Optimized GeoTIFF file format.
func cogTilekLess(a, b extsort.SortType) bool {
	aa := a.(cogTile)
	bb := b.(cogTile)
	if aa.zoom != bb.zoom {
		return aa.zoom > bb.zoom
	} else if aa.y != bb.y {
		return aa.y < bb.y
	} else {
		return aa.x < bb.x
	}
}

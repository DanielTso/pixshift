package transform

import (
	"image"
	"image/draw"
)

// AutoRotate applies the rotation/flip indicated by the EXIF orientation tag.
// Orientation values follow the EXIF spec (1-8). If orientation is 0 or 1 the
// image is returned unchanged.
func AutoRotate(img image.Image, exifOrientation int) image.Image {
	switch exifOrientation {
	case 2:
		return flipH(img)
	case 3:
		return rotate180(img)
	case 4:
		return flipV(img)
	case 5:
		return flipH(rotate90CW(img))
	case 6:
		return rotate90CW(img)
	case 7:
		return flipH(rotate90CCW(img))
	case 8:
		return rotate90CCW(img)
	default:
		return img
	}
}

// toRGBA converts any image.Image to *image.RGBA for direct pixel manipulation.
func toRGBA(img image.Image) *image.RGBA {
	if rgba, ok := img.(*image.RGBA); ok {
		return rgba
	}
	b := img.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), img, b.Min, draw.Src)
	return dst
}

// rotate90CW rotates the image 90 degrees clockwise using direct Pix slice ops.
func rotate90CW(img image.Image) image.Image {
	src := toRGBA(img)
	w := src.Bounds().Dx()
	h := src.Bounds().Dy()
	dst := image.NewRGBA(image.Rect(0, 0, h, w))

	srcPix := src.Pix
	dstPix := dst.Pix
	srcStride := src.Stride
	dstStride := dst.Stride

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			si := y*srcStride + x*4
			// dst(h-1-y, x) = src(x, y)
			dx := h - 1 - y
			dy := x
			di := dy*dstStride + dx*4
			dstPix[di+0] = srcPix[si+0]
			dstPix[di+1] = srcPix[si+1]
			dstPix[di+2] = srcPix[si+2]
			dstPix[di+3] = srcPix[si+3]
		}
	}
	return dst
}

// rotate180 rotates the image 180 degrees using direct Pix slice ops.
func rotate180(img image.Image) image.Image {
	src := toRGBA(img)
	w := src.Bounds().Dx()
	h := src.Bounds().Dy()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))

	srcPix := src.Pix
	dstPix := dst.Pix
	srcStride := src.Stride
	dstStride := dst.Stride

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			si := y*srcStride + x*4
			dx := w - 1 - x
			dy := h - 1 - y
			di := dy*dstStride + dx*4
			dstPix[di+0] = srcPix[si+0]
			dstPix[di+1] = srcPix[si+1]
			dstPix[di+2] = srcPix[si+2]
			dstPix[di+3] = srcPix[si+3]
		}
	}
	return dst
}

// rotate90CCW rotates the image 90 degrees counter-clockwise using direct Pix slice ops.
func rotate90CCW(img image.Image) image.Image {
	src := toRGBA(img)
	w := src.Bounds().Dx()
	h := src.Bounds().Dy()
	dst := image.NewRGBA(image.Rect(0, 0, h, w))

	srcPix := src.Pix
	dstPix := dst.Pix
	srcStride := src.Stride
	dstStride := dst.Stride

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			si := y*srcStride + x*4
			// dst(y, w-1-x) = src(x, y)
			dx := y
			dy := w - 1 - x
			di := dy*dstStride + dx*4
			dstPix[di+0] = srcPix[si+0]
			dstPix[di+1] = srcPix[si+1]
			dstPix[di+2] = srcPix[si+2]
			dstPix[di+3] = srcPix[si+3]
		}
	}
	return dst
}

// flipH flips the image horizontally using direct Pix slice row swaps.
func flipH(img image.Image) image.Image {
	src := toRGBA(img)
	w := src.Bounds().Dx()
	h := src.Bounds().Dy()

	// Copy first.
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	copy(dst.Pix, src.Pix)

	pix := dst.Pix
	stride := dst.Stride

	for y := 0; y < h; y++ {
		rowOff := y * stride
		for x := 0; x < w/2; x++ {
			li := rowOff + x*4
			ri := rowOff + (w-1-x)*4
			// Swap 4 bytes.
			pix[li+0], pix[ri+0] = pix[ri+0], pix[li+0]
			pix[li+1], pix[ri+1] = pix[ri+1], pix[li+1]
			pix[li+2], pix[ri+2] = pix[ri+2], pix[li+2]
			pix[li+3], pix[ri+3] = pix[ri+3], pix[li+3]
		}
	}
	return dst
}

// flipV flips the image vertically using direct Pix slice row swaps.
func flipV(img image.Image) image.Image {
	src := toRGBA(img)
	w := src.Bounds().Dx()
	h := src.Bounds().Dy()

	// Copy first.
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	copy(dst.Pix, src.Pix)

	pix := dst.Pix
	stride := dst.Stride
	rowBuf := make([]byte, w*4)

	for y := 0; y < h/2; y++ {
		topOff := y * stride
		botOff := (h - 1 - y) * stride
		// Swap entire rows.
		copy(rowBuf, pix[topOff:topOff+w*4])
		copy(pix[topOff:topOff+w*4], pix[botOff:botOff+w*4])
		copy(pix[botOff:botOff+w*4], rowBuf)
	}
	return dst
}

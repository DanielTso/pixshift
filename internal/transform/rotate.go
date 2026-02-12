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

// rotate90CW rotates the image 90 degrees clockwise.
func rotate90CW(img image.Image) image.Image {
	b := img.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dy(), b.Dx()))
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			dst.Set(b.Max.Y-1-y, x, img.At(x, y))
		}
	}
	return dst
}

// rotate180 rotates the image 180 degrees.
func rotate180(img image.Image) image.Image {
	b := img.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			dst.Set(b.Max.X-1-x, b.Max.Y-1-y, img.At(x, y))
		}
	}
	return dst
}

// rotate90CCW rotates the image 90 degrees counter-clockwise.
func rotate90CCW(img image.Image) image.Image {
	b := img.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dy(), b.Dx()))
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			dst.Set(y, b.Max.X-1-x, img.At(x, y))
		}
	}
	return dst
}

// flipH flips the image horizontally.
func flipH(img image.Image) image.Image {
	b := img.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), img, b.Min, draw.Src)
	w := b.Dx()
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < w/2; x++ {
			l := dst.At(x, y)
			r := dst.At(w-1-x, y)
			dst.Set(x, y, r)
			dst.Set(w-1-x, y, l)
		}
	}
	return dst
}

// flipV flips the image vertically.
func flipV(img image.Image) image.Image {
	b := img.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), img, b.Min, draw.Src)
	h := b.Dy()
	for y := 0; y < h/2; y++ {
		for x := 0; x < b.Dx(); x++ {
			t := dst.At(x, y)
			bv := dst.At(x, h-1-y)
			dst.Set(x, y, bv)
			dst.Set(x, h-1-y, t)
		}
	}
	return dst
}

package transform

import (
	"image"
	"image/draw"
	"math"
)

const smartCropBlockSize = 16

// SmartCrop crops the image to the target dimensions, selecting the region
// with the highest visual entropy (most detail/interest). This uses color
// variance analysis rather than face detection, making it suitable for
// landscapes, products, and general photography.
func SmartCrop(img image.Image, targetW, targetH int) image.Image {
	b := img.Bounds()
	srcW := b.Dx()
	srcH := b.Dy()

	// If target >= source in both dimensions, return source unchanged.
	if targetW >= srcW && targetH >= srcH {
		return img
	}

	// Clamp target to source dimensions.
	if targetW > srcW {
		targetW = srcW
	}
	if targetH > srcH {
		targetH = srcH
	}

	// For very small images (smaller than block size), fall back to center crop.
	if srcW <= smartCropBlockSize || srcH <= smartCropBlockSize {
		return centerCrop(img, targetW, targetH)
	}

	// If target matches source after clamping, return unchanged.
	if targetW == srcW && targetH == srcH {
		return img
	}

	// Build importance map using color variance in blocks.
	blocksX := srcW / smartCropBlockSize
	blocksY := srcH / smartCropBlockSize
	importance := buildImportanceMap(img, blocksX, blocksY)

	// Sliding window search: find the position with highest total importance.
	bestX, bestY := findBestWindow(importance, blocksX, blocksY, srcW, srcH, targetW, targetH)

	// Extract the chosen region.
	rect := image.Rect(
		b.Min.X+bestX, b.Min.Y+bestY,
		b.Min.X+bestX+targetW, b.Min.Y+bestY+targetH,
	)

	type subImager interface {
		SubImage(r image.Rectangle) image.Image
	}
	if si, ok := img.(subImager); ok {
		return si.SubImage(rect)
	}

	dst := image.NewRGBA(image.Rect(0, 0, targetW, targetH))
	draw.Draw(dst, dst.Bounds(), img, rect.Min, draw.Src)
	return dst
}

// buildImportanceMap computes the color variance for each block in the image.
// Higher variance means more visual detail/interest.
func buildImportanceMap(img image.Image, blocksX, blocksY int) []float64 {
	b := img.Bounds()
	importance := make([]float64, blocksX*blocksY)

	for by := 0; by < blocksY; by++ {
		for bx := 0; bx < blocksX; bx++ {
			x0 := b.Min.X + bx*smartCropBlockSize
			y0 := b.Min.Y + by*smartCropBlockSize
			importance[by*blocksX+bx] = blockVariance(img, x0, y0, smartCropBlockSize)
		}
	}

	return importance
}

// blockVariance computes the standard deviation of RGB values across all
// pixels in a block, used as a measure of visual interest.
func blockVariance(img image.Image, x0, y0, size int) float64 {
	var sumR, sumG, sumB float64
	var sumR2, sumG2, sumB2 float64
	n := float64(size * size)

	for y := y0; y < y0+size; y++ {
		for x := x0; x < x0+size; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			rf := float64(r >> 8)
			gf := float64(g >> 8)
			bf := float64(b >> 8)
			sumR += rf
			sumG += gf
			sumB += bf
			sumR2 += rf * rf
			sumG2 += gf * gf
			sumB2 += bf * bf
		}
	}

	// Variance = E[X^2] - (E[X])^2, then take sqrt for stddev.
	varR := sumR2/n - (sumR/n)*(sumR/n)
	varG := sumG2/n - (sumG/n)*(sumG/n)
	varB := sumB2/n - (sumB/n)*(sumB/n)

	return math.Sqrt(varR + varG + varB)
}

// findBestWindow searches for the window position with the highest sum of
// importance values. Uses block-sized steps for efficiency.
func findBestWindow(importance []float64, blocksX, blocksY, srcW, srcH, targetW, targetH int) (int, int) {
	// Window size in blocks (how many blocks fit in the target dimensions).
	winBlocksX := targetW / smartCropBlockSize
	winBlocksY := targetH / smartCropBlockSize
	if winBlocksX > blocksX {
		winBlocksX = blocksX
	}
	if winBlocksY > blocksY {
		winBlocksY = blocksY
	}
	if winBlocksX < 1 {
		winBlocksX = 1
	}
	if winBlocksY < 1 {
		winBlocksY = 1
	}

	// Maximum pixel offset for the window (based on actual source size).
	maxPixelX := srcW - targetW
	maxPixelY := srcH - targetH
	if maxPixelX < 0 {
		maxPixelX = 0
	}
	if maxPixelY < 0 {
		maxPixelY = 0
	}

	// Maximum block offset to keep the window within the importance map.
	maxBlockX := blocksX - winBlocksX
	maxBlockY := blocksY - winBlocksY

	bestScore := -1.0
	bestX, bestY := 0, 0

	// Step through possible positions at block granularity.
	step := smartCropBlockSize
	for pixY := 0; pixY <= maxPixelY; pixY += step {
		blockOffY := pixY / smartCropBlockSize
		if blockOffY > maxBlockY {
			blockOffY = maxBlockY
		}
		for pixX := 0; pixX <= maxPixelX; pixX += step {
			blockOffX := pixX / smartCropBlockSize
			if blockOffX > maxBlockX {
				blockOffX = maxBlockX
			}

			score := windowScore(importance, blocksX, blockOffX, blockOffY, winBlocksX, winBlocksY)
			if score > bestScore {
				bestScore = score
				bestX = pixX
				bestY = pixY
			}
		}
	}

	// Also check the exact maximum offset (may not be block-aligned).
	if maxPixelX%step != 0 || maxPixelY%step != 0 {
		blockOffX := maxPixelX / smartCropBlockSize
		if blockOffX > maxBlockX {
			blockOffX = maxBlockX
		}
		blockOffY := maxPixelY / smartCropBlockSize
		if blockOffY > maxBlockY {
			blockOffY = maxBlockY
		}
		score := windowScore(importance, blocksX, blockOffX, blockOffY, winBlocksX, winBlocksY)
		if score > bestScore {
			bestX = maxPixelX
			bestY = maxPixelY
		}
	}

	return bestX, bestY
}

// windowScore sums importance values for blocks within the given window.
func windowScore(importance []float64, blocksX, offX, offY, winW, winH int) float64 {
	var sum float64
	for y := offY; y < offY+winH; y++ {
		for x := offX; x < offX+winW; x++ {
			sum += importance[y*blocksX+x]
		}
	}
	return sum
}

// centerCrop extracts a centered region of targetW x targetH from img.
func centerCrop(img image.Image, targetW, targetH int) image.Image {
	b := img.Bounds()
	x := b.Min.X + (b.Dx()-targetW)/2
	y := b.Min.Y + (b.Dy()-targetH)/2
	rect := image.Rect(x, y, x+targetW, y+targetH)

	type subImager interface {
		SubImage(r image.Rectangle) image.Image
	}
	if si, ok := img.(subImager); ok {
		return si.SubImage(rect)
	}

	dst := image.NewRGBA(image.Rect(0, 0, targetW, targetH))
	draw.Draw(dst, dst.Bounds(), img, rect.Min, draw.Src)
	return dst
}

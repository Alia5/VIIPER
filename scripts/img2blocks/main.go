// img2blocks converts an image to colored block-character ASCII art with dithering
package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"
)

// Block characters from darkest to lightest
var blockChars = []rune{'█', '▓', '▒', '░', ' '}

func main() {
	width := flag.Int("w", 40, "Output width in characters")
	height := flag.Int("h", 0, "Output height (0 = auto based on aspect ratio)")
	invert := flag.Bool("invert", false, "Invert brightness")
	output := flag.String("o", "", "Output file (default: stdout)")
	svgScale := flag.Float64("svg-scale", 4.0, "Scale factor for SVG rendering (higher = better quality)")
	dither := flag.Float64("dither", 1.0, "Dithering strength (0=none, 1=normal, 2+=aggressive)")
	alpha := flag.Int("alpha", 128, "Alpha threshold (0-255, pixels below this become spaces)")
	blocks := flag.String("blocks", "█▓▒░ ", "Block characters from darkest to lightest")
	braille := flag.Bool("braille", false, "Use braille characters (2x4 dots per char, ignores -blocks)")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: img2blocks [options] <image-file>")
		fmt.Fprintln(os.Stderr, "Supports: PNG, JPG, GIF, BMP, WebP, SVG")
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	inputFile := flag.Arg(0)
	var img image.Image
	var err error

	// Check if it's an SVG
	ext := strings.ToLower(filepath.Ext(inputFile))
	if ext == ".svg" {
		img, err = loadSVG(inputFile, *svgScale)
	} else {
		img, err = loadRasterImage(inputFile)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading image: %v\n", err)
		os.Exit(1)
	}

	opts := convertOptions{
		width:       *width,
		height:      *height,
		invert:      *invert,
		dither:      *dither,
		alphaThresh: uint8(*alpha),
		blockChars:  []rune(*blocks),
		braille:     *braille,
	}

	var result string
	if *braille {
		result = convertToBraille(img, opts)
	} else {
		result = convertToBlocks(img, opts)
	}

	if *output != "" {
		err = os.WriteFile(*output, []byte(result), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Print(result)
	}
}

func loadRasterImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	return img, err
}

func loadSVG(path string, scale float64) (image.Image, error) {
	icon, err := oksvg.ReadIcon(path, oksvg.IgnoreErrorMode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SVG: %w", err)
	}

	// Get the SVG dimensions and scale up for better quality
	w := int(icon.ViewBox.W * scale)
	h := int(icon.ViewBox.H * scale)

	icon.SetTarget(0, 0, float64(w), float64(h))

	// Create RGBA image with transparent background
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Fill with transparent (already default, but be explicit)
	for i := range img.Pix {
		img.Pix[i] = 0
	}

	// Rasterize the SVG
	scanner := rasterx.NewScannerGV(w, h, img, img.Bounds())
	raster := rasterx.NewDasher(w, h, scanner)
	icon.Draw(raster, 1.0)

	return img, nil
}

type convertOptions struct {
	width       int
	height      int
	invert      bool
	dither      float64
	alphaThresh uint8
	blockChars  []rune
	braille     bool
}

func convertToBlocks(img image.Image, opts convertOptions) string {
	bounds := img.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	targetWidth := opts.width
	targetHeight := opts.height

	// Terminal characters are roughly 2:1 aspect ratio (taller than wide)
	charAspect := 0.5

	if targetHeight == 0 {
		targetHeight = int(float64(targetWidth) * float64(srcHeight) / float64(srcWidth) * charAspect)
	}

	if targetHeight < 1 {
		targetHeight = 1
	}

	// Create error diffusion buffer for Floyd-Steinberg dithering
	// We'll dither the luminance channel
	errBuf := make([][]float64, targetHeight+1)
	for i := range errBuf {
		errBuf[i] = make([]float64, targetWidth+2) // +2 for boundary
	}

	var out strings.Builder
	var lastR, lastG, lastB uint8 = 0, 0, 0
	colorSet := false

	for y := 0; y < targetHeight; y++ {
		for x := 0; x < targetWidth; x++ {
			// Sample the source image (average the region)
			r, g, b, a := sampleRegion(img, bounds,
				x*srcWidth/targetWidth,
				y*srcHeight/targetHeight,
				(x+1)*srcWidth/targetWidth,
				(y+1)*srcHeight/targetHeight,
			)

// Calculate luminance (perceived brightness)
			lum := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)

			// Handle transparency - alpha first, then dark pixels
			if a < opts.alphaThresh || lum < float64(opts.alphaThresh) {
				if colorSet {
					out.WriteString("\x1b[0m")
					colorSet = false
				}
				out.WriteRune(' ')
				continue
			}

			// Add error from dithering (scaled by dither strength)
			lum += errBuf[y][x+1] * opts.dither

			// Clamp
			if lum < 0 {
				lum = 0
			}
			if lum > 255 {
				lum = 255
			}

			// Quantize to block character
			numChars := len(opts.blockChars)
			charIdx := int(lum/255.0*float64(numChars-1) + 0.5)
			if charIdx >= numChars {
				charIdx = numChars - 1
			}
			if opts.invert {
				charIdx = numChars - 1 - charIdx
			}

			// Calculate quantization error
			quantizedLum := float64(numChars-1-charIdx) / float64(numChars-1) * 255
			if opts.invert {
				quantizedLum = float64(charIdx) / float64(numChars-1) * 255
			}
			quantErr := lum - quantizedLum

			// Floyd-Steinberg error diffusion
			errBuf[y][x+2] += quantErr * 7 / 16
			if y+1 < targetHeight {
				errBuf[y+1][x] += quantErr * 3 / 16
				errBuf[y+1][x+1] += quantErr * 5 / 16
				errBuf[y+1][x+2] += quantErr * 1 / 16
			}

			// Output with color (only change color if different)
			char := opts.blockChars[charIdx]

			// For space character, no need to set color
			if char == ' ' {
				if colorSet {
					out.WriteString("\x1b[0m")
					colorSet = false
				}
				out.WriteRune(' ')
			} else {
				// Set foreground color if changed
				if !colorSet || r != lastR || g != lastG || b != lastB {
					out.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b))
					lastR, lastG, lastB = r, g, b
					colorSet = true
				}
				out.WriteRune(char)
			}
		}

		// Reset color at end of line and add newline
		if colorSet {
			out.WriteString("\x1b[0m")
			colorSet = false
		}
		out.WriteRune('\n')
	}

	return out.String()
}

func sampleRegion(img image.Image, bounds image.Rectangle, x1, y1, x2, y2 int) (r, g, b, a uint8) {
	if x2 <= x1 {
		x2 = x1 + 1
	}
	if y2 <= y1 {
		y2 = y1 + 1
	}

	var sumR, sumG, sumB, sumA uint64
	count := uint64(0)

	for py := y1; py < y2; py++ {
		for px := x1; px < x2; px++ {
			c := img.At(bounds.Min.X+px, bounds.Min.Y+py)
			cr, cg, cb, ca := c.RGBA()
			// RGBA returns 16-bit values, convert to 8-bit
			sumR += uint64(cr >> 8)
			sumG += uint64(cg >> 8)
			sumB += uint64(cb >> 8)
			sumA += uint64(ca >> 8)
			count++
		}
	}

	if count == 0 {
		return 0, 0, 0, 0
	}

	return uint8(sumR / count), uint8(sumG / count), uint8(sumB / count), uint8(sumA / count)
}

// convertToBraille converts image to braille characters (2x4 dots per char)
func convertToBraille(img image.Image, opts convertOptions) string {
	bounds := img.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	targetWidth := opts.width
	targetHeight := opts.height

	// Braille is 2x4 dots, so we need pixel dimensions
	// Each character represents 2 pixels wide and 4 pixels tall
	// Terminal chars are ~2:1 aspect, braille dots compensate
	if targetHeight == 0 {
		// Calculate based on aspect ratio
		// Each braille char is 2 dots wide, 4 dots tall
		// With terminal aspect ~2:1, effective ratio is about 1:1
		targetHeight = int(float64(targetWidth) * float64(srcHeight) / float64(srcWidth) * 0.5)
	}
	if targetHeight < 1 {
		targetHeight = 1
	}

	// Pixel dimensions (2x width, 4x height for braille dots)
	pixelWidth := targetWidth * 2
	pixelHeight := targetHeight * 4

	// Sample image to pixel grid
	type pixel struct {
		r, g, b, a uint8
	}
	pixels := make([][]pixel, pixelHeight)
	for py := 0; py < pixelHeight; py++ {
		pixels[py] = make([]pixel, pixelWidth)
		for px := 0; px < pixelWidth; px++ {
			r, g, b, a := sampleRegion(img, bounds,
				px*srcWidth/pixelWidth,
				py*srcHeight/pixelHeight,
				(px+1)*srcWidth/pixelWidth,
				(py+1)*srcHeight/pixelHeight,
			)
			pixels[py][px] = pixel{r, g, b, a}
		}
	}

	// Braille dot positions (Unicode braille pattern):
	// Dot 1 (0x01) Dot 4 (0x08)
	// Dot 2 (0x02) Dot 5 (0x10)
	// Dot 3 (0x04) Dot 6 (0x20)
	// Dot 7 (0x40) Dot 8 (0x80)
	dotBits := [4][2]rune{
		{0x01, 0x08},
		{0x02, 0x10},
		{0x04, 0x20},
		{0x40, 0x80},
	}

	var out strings.Builder
	var lastR, lastG, lastB uint8 = 0, 0, 0
	colorSet := false

	for cy := 0; cy < targetHeight; cy++ {
		for cx := 0; cx < targetWidth; cx++ {
			var brailleChar rune = 0x2800 // Base braille character
			var totalR, totalG, totalB uint32
			var colorCount uint32

			// Check each of the 8 dots in the 2x4 braille cell
			for dy := 0; dy < 4; dy++ {
				for dx := 0; dx < 2; dx++ {
					py := cy*4 + dy
					px := cx*2 + dx

					if py >= pixelHeight || px >= pixelWidth {
						continue
					}

					p := pixels[py][px]

					// Skip transparent pixels (alpha first, then dark pixels)
					if p.a < opts.alphaThresh {
						continue
					}
					lum := 0.299*float64(p.r) + 0.587*float64(p.g) + 0.114*float64(p.b)
					if lum < float64(opts.alphaThresh) {
						continue
					}

					// Non-transparent pixel = dot on
					brailleChar |= dotBits[dy][dx]

					// Accumulate color
					totalR += uint32(p.r)
					totalG += uint32(p.g)
					totalB += uint32(p.b)
					colorCount++
				}
			}

			// If no visible pixels in cell, output space
			if colorCount == 0 {
				if colorSet {
					out.WriteString("\x1b[0m")
					colorSet = false
				}
				out.WriteRune(' ')
				continue
			}

			// Average color for this cell
			avgR := uint8(totalR / colorCount)
			avgG := uint8(totalG / colorCount)
			avgB := uint8(totalB / colorCount)

			// Set color if changed
			if !colorSet || avgR != lastR || avgG != lastG || avgB != lastB {
				out.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm", avgR, avgG, avgB))
				lastR, lastG, lastB = avgR, avgG, avgB
				colorSet = true
			}

			out.WriteRune(brailleChar)
		}

		// Reset color at end of line
		if colorSet {
			out.WriteString("\x1b[0m")
			colorSet = false
		}
		out.WriteRune('\n')
	}

	return out.String()
}

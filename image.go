// utility/image.go
package Utility

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"strings"

	"github.com/chai2010/webp"
	"github.com/nfnt/resize"
	"github.com/polds/imgbase64"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

// SvgToPng converts an SVG file into a PNG at the given dimensions.
func SvgToPng(input, output string, w, h int) error {
	in, err := os.Open(input)
	if err != nil {
		return err
	}
	defer in.Close()

	icon, _ := oksvg.ReadIconStream(in)
	icon.SetTarget(0, 0, float64(w), float64(h))
	rgba := image.NewRGBA(image.Rect(0, 0, w, h))
	icon.Draw(rasterx.NewDasher(w, h, rasterx.NewScannerGV(w, h, rgba, rgba.Bounds())), 1)

	out, err := os.Create(output)
	if err != nil {
		return err
	}
	defer out.Close()

	return png.Encode(out, rgba)
}

// CreateThumbnail resizes an image and returns its base64 representation.
func CreateThumbnail(path string, thumbnailMaxHeight int, thumbnailMaxWidth int) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	file.Seek(0, 0)
	var originalImg image.Image

	if strings.HasSuffix(strings.ToLower(file.Name()), ".png") {
		originalImg, err = png.Decode(file)
	} else if strings.HasSuffix(strings.ToLower(file.Name()), ".jpg") ||
		strings.HasSuffix(strings.ToLower(file.Name()), ".jpeg") {
		originalImg, err = jpeg.Decode(file)
	} else if strings.HasSuffix(strings.ToLower(file.Name()), ".gif") {
		originalImg, err = gif.Decode(file)
	} else if strings.HasSuffix(strings.ToLower(file.Name()), ".webp") {
		originalImg, err = webp.Decode(file)
	} else {
		return "", errors.New("unsupported image format: " + file.Name())
	}
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	var img image.Image
	if thumbnailMaxHeight == -1 && thumbnailMaxWidth == -1 {
		img = originalImg
	} else {
		hRatio := thumbnailMaxHeight / originalImg.Bounds().Size().Y
		wRatio := thumbnailMaxWidth / originalImg.Bounds().Size().X

		var h, w int
		if hRatio*originalImg.Bounds().Size().Y < thumbnailMaxWidth {
			h = thumbnailMaxHeight
			w = hRatio * originalImg.Bounds().Size().Y
		} else {
			h = wRatio * thumbnailMaxHeight
			w = thumbnailMaxWidth
		}

		// don’t upscale
		if hRatio > 1 {
			h = originalImg.Bounds().Size().Y
		}
		if wRatio > 1 {
			w = originalImg.Bounds().Size().X
		}

		img = resize.Resize(uint(h), uint(w), originalImg, resize.Lanczos3)
	}

	var buf bytes.Buffer
	if strings.HasSuffix(strings.ToLower(file.Name()), ".png") {
		err = png.Encode(&buf, img)
	} else {
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: jpeg.DefaultQuality})
	}
	if err != nil {
		return "", fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	return imgbase64.FromBuffer(buf), nil
}


package services

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	_ "golang.org/x/image/webp"

	"avatar-ready-converter/processor/models"
	"github.com/disintegration/imaging"
)

type ImageProcessor struct{}

func NewImageProcessor() *ImageProcessor {
	return &ImageProcessor{}
}

// Process takes an image and processing options, returns processed image bytes
func (ip *ImageProcessor) Process(imageData []byte, opts models.ProcessOptions) ([]byte, error) {
	opts.SetDefaults()

	// Decode the image
	img, err := imaging.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	
	// Process: crop to square and resize
	processedImg := ip.cropToSquare(img)
	processedImg = imaging.Resize(processedImg, opts.Size, opts.Size, imaging.Lanczos)

	// Encode to requested format
	var buf bytes.Buffer
	switch opts.Format {
	case "png":
		err = png.Encode(&buf, processedImg)
	case "jpg", "jpeg":
		err = jpeg.Encode(&buf, processedImg, &jpeg.Options{Quality: 90})
		default:
			return nil, fmt.Errorf("unsupported format: %s", opts.Format)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	return buf.Bytes(), nil
}

// cropToSquare crops image to center square
func (ip *ImageProcessor) cropToSquare(img image.Image) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Already square
	if width == height {
		return img
	}

	// Crop to center square
	size := width
	if height < width {
		size = height
	}

	return imaging.CropCenter(img, size, size)
}

package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"

	"github.com/nfnt/resize"
)

// The file naming pattern is S_X_Y_<rest of the crap we don't care about>.png
var (
	satellitePngNamePattern = regexp.MustCompile(`.*S_(?P<x>\d{3})_(?P<y>\d{3})_lco\.png`)
	// This is for 512x512 images. Divide by 512/size
	imageOverlap = 30
	maxSize      = 512
)

// Stitches input images together. Missing images are assumed to be transparent
func StitchImages(paths []string, subImageSize image.Point) (stitchedImage *image.RGBA, err error) {
	sort.Strings(paths)

	var maxX, maxY int
	for _, imagePath := range paths {
		point := pointFromFileName(path.Base(imagePath))
		if point.X > maxX {
			maxX = point.X + 1
		}
		if point.Y > maxY {
			maxY = point.Y + 1
		}
	}

	segmentSize := imageOverlap / (maxSize / subImageSize.X)

	width := (maxX * subImageSize.X) - (maxX * segmentSize)
	height := (maxY * subImageSize.Y) - (maxY * segmentSize)

	fmt.Println("Image size should be", width, height)

	// Create the in-memory RGBA image
	rgbaImage := image.NewRGBA(image.Rect(0, 0, width, height))

	for i, imagePath := range paths {
		fmt.Printf("Stitching %d/%d\n", i+1, len(paths))
		if !satellitePngNamePattern.Match([]byte(imagePath)) {
			return nil, fmt.Errorf("Invalid filename: %s", imagePath)
		}

		point := pointFromFileName(path.Base(imagePath))
		point.X *= subImageSize.X - segmentSize
		point.Y *= subImageSize.Y - segmentSize

		var gridPart image.Image

		// Load the PNG image
		file, err := os.Open(imagePath)
		if err != nil {
			return nil, fmt.Errorf("While converting %s: %s", imagePath, err)
		}

		gridPart, err = png.Decode(file)
		if err != nil {
			return nil, fmt.Errorf("While converting %s: %s", imagePath, err)
		}

		// Repeat this image
		if gridPart.Bounds().Size().X < subImageSize.X {
			// Note that even though the second argument is the bounds,
			// the effective rectangle is smaller due to clipping.
			gridPart = resize.Resize(uint(subImageSize.X), uint(subImageSize.Y), gridPart, resize.Lanczos3)
		}

		// This will turn out being something like 0,0, 512, 512 which is grid index 0,0 (top-left)
		// or something like, 512,0, 1024, 512 which is grid index 1, 0
		destRect := image.Rectangle{point, point.Add(subImageSize)}
		draw.Draw(rgbaImage, destRect, gridPart, gridPart.Bounds().Min, draw.Src)

		file.Close()
	}

	return rgbaImage, nil
}

// Converts a file name of form S_<X>_<Y>_lco.png to an `image.Point`
func pointFromFileName(name string) image.Point {
	matches := satellitePngNamePattern.FindStringSubmatch(name)

	p := image.Point{}

	for i, name := range satellitePngNamePattern.SubexpNames() {
		// Ignore the whole regexp match and unnamed groups
		if i == 0 || name == "" {
			continue
		}

		intval, _ := strconv.Atoi(matches[i])
		if name == "x" {
			p.X = intval
		} else if name == "y" {
			p.Y = intval
		}
	}

	return p
}

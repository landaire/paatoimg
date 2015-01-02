package main

import (
	"errors"
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
)

// The file naming pattern is S_X_Y_<rest of the crap we don't care about>.png
var (
	satelliteNamePattern     = regexp.MustCompile(`.*S_(?P<x>\d{3})_(?P<y>\d{3})_lco\.png`)
	ErrPathsNotSquare        = errors.New("Number of image paths is not a perfect square")
	ErrInvalidGridSquareSize = errors.New("Invalid grid subimage size")
)

// Stitches input images together. The number of paths must be a perfect square
// (e.g. 4, 9, 16, 25, 64) since the resulting image is expected to be a grid
func StitchImages(paths []string, subImageSize image.Point) (stitchedImage *image.RGBA, err error) {
	// Check that the number of paths is a perfect square
	if !isPerfectSquare(float64(len(paths))) {
		return nil, ErrPathsNotSquare
	}

	width := int(math.Floor(math.Sqrt(float64(len(paths))))) * subImageSize.X
	height := width

	// Create the in-memory RGBA image
	rgbaImage := image.NewRGBA(image.Rect(0, 0, width, height))

	// Sort the paths
	sort.Strings(paths)

	for _, imagePath := range paths {
		if !satelliteNamePattern.Match([]byte(imagePath)) {
			return nil, fmt.Errorf("Invalid filename: %s", imagePath)
		}

		point := pointFromFileName(path.Base(imagePath))
		point.X *= subImageSize.X
		point.Y *= subImageSize.Y

		// Load the PNG image
		file, err := os.Open(imagePath)
		if err != nil {
			return nil, err
		}

		gridPart, err := png.Decode(file)
		if err != nil {
			return nil, err
		}

		if gridPart.Bounds().Size() != subImageSize {
			return nil, ErrInvalidGridSquareSize
		}

		// This will turn out being something like 0,0, 512, 512 which is grid index 0,0 (top-left)
		// or something like, 512,0, 1024, 512 which is grid index 1, 0
		destRect := image.Rectangle{point, point.Add(subImageSize)}

		//		fmt.Println("Drawing image starting at (%d, %d) to (%d, %d) from source")
		draw.Draw(rgbaImage, destRect, gridPart, gridPart.Bounds().Min, draw.Src)

		file.Close()
	}

	return rgbaImage, nil
}

func isPerfectSquare(num float64) bool {
	sqrt := math.Floor(math.Sqrt(num))

	return sqrt*sqrt == num
}

func pointFromFileName(name string) image.Point {
	matches := satelliteNamePattern.FindStringSubmatch(name)

	p := image.Point{}

	for i, name := range satelliteNamePattern.SubexpNames() {
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

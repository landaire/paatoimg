package main

import (
	"bytes"
	"crypto/sha1"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"path/filepath"

	"github.com/landaire/osutil"
)

const testImagesDir = "./imagetests"

func TestImageStitching(t *testing.T) {
	if exists, err := osutil.Exists(testImagesDir); !exists || err != nil {
		t.Fatal("Image testing dir does not exist")
	}

	pngdir := filepath.Join(testImagesDir, "png")
	files, err := ioutil.ReadDir(pngdir)

	if err != nil {
		t.Fatal(err)
	}

	paths := []string{}
	for _, file := range files {
		if satellitePngNamePattern.Match([]byte(file.Name())) {
			abs, err := filepath.Abs(filepath.Join(pngdir, file.Name()))
			if err != nil {
				t.Fatal(err)
			}
			paths = append(paths, abs)
		}
	}

	image, err := StitchImages(paths, image.Point{512, 512})
	if err != nil {
		t.Fatal(err)
	}

	// Change this to true to update the out.png
	if false {
		outFile, _ := os.Create(filepath.Join(testImagesDir, "out.png"))
		err = png.Encode(outFile, image)
		outFile.Close()
	}

	// Open the good output file
	file, err := os.Open(filepath.Join(testImagesDir, "out.png"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	// Ensure the expected and actual output images are the same
	actualSha := sha1.New()
	err = png.Encode(actualSha, image)
	if err != nil {
		t.Fatal(err)
	}

	expectedSha := sha1.New()
	io.Copy(expectedSha, file)

	if bytes.Compare(expectedSha.Sum(nil), actualSha.Sum(nil)) > 0 {
		t.Fatal("Expected SHA1 != actual SHA1")
	}
}

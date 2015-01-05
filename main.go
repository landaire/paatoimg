// Command paatoimg is used for extracting Arma 3 satellite PAA files from PBO files, converting the PAA files to PNG
// at the specified resolution, then stitching all of the images together.
//
// Notes
//
// If a large size argument is supplied (> 256), you will more than likely end up using more than 5 GB of memory. My Windows
// machine only has 8 GB and I get a memory exception when trying to render images with --size=512
package main

import (
	"fmt"
	"image"
	"image/png"
	"os"

	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/landaire/osutil"
)

const (
	version = "0.1.0"
)

func main() {
	app := cli.NewApp()
	app.Name = "paatoimg"
	app.Usage = "Extract satellite PAA files from PBO archives, and convert them to a giant stitched maps"
	app.Author = "Lander Brandt"
	app.Email = "@landaire"
	app.Version = version

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "pbo",
			Usage: "Base PBO file to read (others will be detected automatically)",
		},
		cli.StringFlag{
			Name:  "outdir, od",
			Usage: "Output directory to dump PNG files",
		},
		cli.StringFlag{
			Name:  "outfile, of",
			Usage: "Output output file to write stitched PNG file",
		},
		cli.BoolFlag{
			Name:  "no",
			Usage: "Disable overwriting existing PNG files (defaults to false)",
		},
		cli.IntFlag{
			Name:  "size, s",
			Usage: "Size of the output image squares",
			Value: 128,
		},
	}

	app.Action = Stitch

	err := app.Run(os.Args)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
	}
}

// Calls `DumpPaaFiles` on the input PBO file, calls `ConvertPaaToPng` if the image in --outdir does not exist and
// --no was supplied, otherwise it always calls `ConvertPaaToPng`. Finally, `StitchImages` is called on PNG files
func Stitch(c *cli.Context) {
	required := []string{"pbo", "outdir", "outfile"}
	for _, flag := range required {
		if !c.GlobalIsSet(flag) {
			c.App.Command("help").Run(c)
			return
		}
	}

	paaFiles, err := DumpPaaFiles(c.String("pbo"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error occurred dumping PAA files:", err)
		os.Exit(1)
	}

	// Create the output dir if it doesn't exist
	if exist, _ := osutil.Exists(c.String("outdir")); !exist {
		err := os.MkdirAll(c.String("outdir"), os.ModePerm)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error occurred creating output dir:", err)
			os.Exit(1)
		}
	}

	pngs := []string{}
	for _, file := range paaFiles {
		paaFileName := filepath.Base(file)
		pngFileName := paaFileName[:len(paaFileName)-len(filepath.Ext(paaFileName))] + ".png"
		pngFileName = filepath.Join(c.String("outdir"), pngFileName)

		if exists, _ := osutil.Exists(pngFileName); (c.Bool("no") && !exists) || !c.Bool("no") {
			output, err := ConvertPaaToPng(file, pngFileName, c.Int("size"))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error occurred converting PAA to PNG: %s\n%s\n", output, err)
				os.Exit(1)
			}
		}

		pngs = append(pngs, pngFileName)
	}

	stitchedImage, err := StitchImages(pngs, image.Point{c.Int("size"), c.Int("size")})

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error occurred creating stitched image:", err)
		os.Exit(1)
	}

	outFile, err := os.Create(c.String("outfile"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error occurred creating output file:", err)
		os.Exit(1)
	}
	defer outFile.Close()

	fmt.Println("Flushing image to disk...")
	err = png.Encode(outFile, stitchedImage)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error occurred while writing PNG:", err)
		os.Exit(1)
	}
}

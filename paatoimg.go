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
	}

	app.Action = Stitch

	err := app.Run(os.Args)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	}
}

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
		fmt.Fprintf(os.Stderr, "Error occurred dumping PAA files: %s", err)
		os.Exit(1)
	}

	// Create the output dir if it doesn't exist
	if exist, _ := osutil.Exists(c.String("outdir")); !exist {
		err := os.MkdirAll(c.String("outdir"), os.ModePerm)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error occurred creating output dir: %s", err)
			os.Exit(1)
		}
	}

	pngs := []string{}
	for _, file := range paaFiles {
		paaFileName := filepath.Base(file)
		pngFileName := paaFileName[:len(paaFileName)-len(filepath.Ext(paaFileName))] + ".png"

		output, err := ConvertPaaToPng(file, filepath.Join(c.String("outdir"), pngFileName))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error occurred converting PAA to PNG: %s\n%s\n", output, err)
			os.Exit(1)
		}

		pngs = append(pngs, pngFileName)
	}

	stitchedImage, err := StitchImages(pngs, image.Point{512, 512})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error occurred creating stitched image: %s", err)
		os.Exit(1)
	}

	outFile, err := os.Create(c.String("outfile"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error occurred creating output file: %s", err)
		os.Exit(1)
	}
	defer outFile.Close()

	err = png.Encode(outFile, stitchedImage)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error occurred while writing PNG: %s", err)
		os.Exit(1)
	}
}

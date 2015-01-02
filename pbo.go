package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"io"

	"github.com/landaire/pbo"
)

const (
	pal2pace = `C:\Program Files (x86)\Bohemia Interactive\Tools\TexView 2\Pal2PacE.exe`
)

var (
	satellitePaaNamePattern = regexp.MustCompile(`.*S_(?P<x>\d{3})_(?P<y>\d{3})_lco\.paa`)
)

func DumpPaaFiles(rootpbo string) (paaFiles []string, err error) {
	// Get all PBO files that will be loaded.
	pbodir := filepath.Dir(rootpbo)

	pboFilename := filepath.Base(rootpbo)
	pboFilenameNoExt := pboFilename[:len(pboFilename)-len(filepath.Ext(pboFilename))]

	layersPattern := regexp.MustCompile(fmt.Sprintf("%s_\\d{2}_\\d{2}.pbo$", pboFilenameNoExt))

	files, err := ioutil.ReadDir(pbodir)

	if err != nil {
		return nil, err
	}

	pboFiles := []string{}
	for _, file := range files {
		if layersPattern.Match([]byte(file.Name())) {
			pboFiles = append(pboFiles, filepath.Join(pbodir, file.Name()))
		}
	}

	// Parse all of the PBO files
	parsedPbos := make([]*pbo.Pbo, len(pboFiles))

	for i, file := range pboFiles {
		pbo, err := pbo.NewPbo(file)
		if err != nil {
			return nil, err
		}

		parsedPbos[i] = pbo
	}

	// Write out the satellite PAA files from each PBO to the tempdir
	tempdir, err := ioutil.TempDir("", "paatoimg")
	if err != nil {
		return nil, err
	}

	fmt.Println(tempdir)

	paaFiles = []string{}
	for _, pbo := range parsedPbos {
		for _, entry := range pbo.Entries {
			if satellitePaaNamePattern.Match([]byte(entry.Name)) {
				paaPath := filepath.Join(tempdir, entry.Name)
				outFile, err := os.Create(paaPath)

				if err != nil {
					return nil, err
				}

				// Copy the PAA data to the temp file
				io.Copy(outFile, entry)

				outFile.Close()

				paaFiles = append(paaFiles, paaPath)
			}
		}
	}

	return paaFiles, nil
}

func ConvertPaaToPng(source, dest string) (errOutput string, err error) {
	command := exec.Command(pal2pace, "-size=512", source, dest)
	err = command.Run()
	output, _ := command.Output()
	errOutput = string(output)

	return
}

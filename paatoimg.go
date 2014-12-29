package main

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/landaire/paa"
)

const (
	path = "/Users/lander/Documents/arma/a3/map_altis/data/layers/00_00/S_002_017_lco.paa"
)

func main() {
	file, _ := os.Open(path)
	defer file.Close()

	taggs, err := paa.ReadPaa(file)

	if err != nil {
		fmt.Println(err)
		return
	}

	for _, tagg := range taggs {
		fmt.Printf("%#v\n", tagg)

		if tagg.Name() == paa.OFFS {
			data := tagg.Data()
			for i, offset := range data {
				if data[i] == 0x0 {
					break
				}

				var length int64

				if i == len(data)-1 || data[i+1] == 0x0 {
					stat, _ := file.Stat()
					length = stat.Size() - int64(offset)
				} else {
					length = int64(data[i+1] - data[i])
				}

				var width, height uint16

				file.Seek(int64(offset), os.SEEK_SET)

				binary.Read(file, binary.LittleEndian, &width)
				binary.Read(file, binary.LittleEndian, &height)

				fmt.Println("Width:", width)
				fmt.Println("Height:", height)
				fmt.Println("Length:", length)
			}
		}
	}
}

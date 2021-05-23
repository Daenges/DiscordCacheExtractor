package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
)

var wg = sync.WaitGroup{}

func main() {

	src, dst, pathSeperator, name, chunkSize, keepUnknownFileTypes, discordMode := readArgs()
	var chunkList []os.FileInfo
	var calculatedFiles = 0

	if len(src) < 1 || len(dst) < 1 {
		printHelp()
		fmt.Printf("src: %v", src)
		fmt.Printf("dst: %v", dst)
		log.Fatal("Invalid paths.")
	}

	files, err := ioutil.ReadDir(src)

	if err != nil {
		fmt.Println("There are no files in this directory.")
		printHelp()
		fmt.Println("PROGRAM STOPPED")
		os.Exit(2)
	}

	sourceBasePath := src + pathSeperator

	fmt.Println("Copying files..")

	for num, f := range files {

		if !strings.Contains(f.Name(), "data") &&
			(!strings.Contains(f.Name(), "index") || !discordMode) &&
			!strings.Contains(f.Name(), ".") {

			chunkList = append(chunkList, f)
		}

		if len(chunkList)%chunkSize == 0 && num != 0 {
			// Start copy process
			wg.Add(1)
			go fileArrayCopy(chunkList, dst, sourceBasePath, name, calculatedFiles, keepUnknownFileTypes)

			// Clear the slice
			chunkList = []os.FileInfo{}
			// Add new name range
			calculatedFiles += chunkSize
			fmt.Printf("%.2f%%\n", (float64(calculatedFiles)/float64(len(files)))*100)
		}
	}

	// Start last process if the number of files was not dividable by the chunksize
	if len(chunkList) > 0 {
		wg.Add(1)
		go fileArrayCopy(chunkList, dst, sourceBasePath, name, calculatedFiles, keepUnknownFileTypes)
	}

	wg.Wait()
	fmt.Println("Done!")
}

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
)

func readArgs() (src string, dst string, pathSeperator string, name string, chunkSize int, keepUnknownFileTypes bool, discordMode bool) {

	if len(os.Args[1:]) < 1 {
		printHelp()
		log.Fatal("Missing arguments!")
	}

	argsWithoutProg := os.Args[1:] // Argument Input

	// STANDARD PARAMETER
	dst, _ = os.Getwd()
	chunkSize = 10
	keepUnknownFileTypes = false
	pathSeperator = "\\"
	name = "Picture_"
	discordMode = true

	if runtime.GOOS != "windows" {
		pathSeperator = "/"
	}

	if len(argsWithoutProg) == 1 { // Accept execution with one parameter without defining it with -src
		src = argsWithoutProg[0]

	} else {
		for i := 0; i < len(argsWithoutProg); i++ {

			if argsWithoutProg[i] == "-src" {
				src = getContentInNextQuotes(argsWithoutProg, i+1)
				if src[:1] != "\\" {
					src += "\\"
				}
				continue
			}

			if argsWithoutProg[i] == "-dst" {
				dst = getContentInNextQuotes(argsWithoutProg, i+1)
				if dst[:1] != "\\" {
					dst += "\\"
				}
				continue
			}

			if argsWithoutProg[i] == "-cs" {
				tmp, err := strconv.Atoi(argsWithoutProg[i+1])
				if err != nil || tmp < 1 {
					chunkSize = 5
					log.Printf("Error: %v \n Set chunkSize to %v.", err, chunkSize)
				} else {
					chunkSize = tmp
				}
				continue
			}

			if argsWithoutProg[i] == "-tc" {
				threadCount, err := strconv.Atoi(argsWithoutProg[i+1])

				if err != nil || threadCount < 1 {
					log.Printf("Error: %v \n threadCount was not set.", err)
				} else {
					runtime.GOMAXPROCS(threadCount)
				}
				continue
			}

			if argsWithoutProg[i] == "-s" {
				pathSeperator = argsWithoutProg[i+1]
			}

			if argsWithoutProg[i] == "-k" {
				keepUnknownFileTypes = true
				continue
			}

			if argsWithoutProg[i] == "-n" {
				name = argsWithoutProg[i+1]
			}

			if argsWithoutProg[i] == "-!d" {
				discordMode = false
			}
		}
	}

	if src == "" || dst == "" { // Error detection
		log.Fatalf("Necessary parameters are missing. Src: %v Dst: %v", src, dst)
	}
	if dst[len(dst)-1:] != pathSeperator { // Make sure the destination is a Folder.
		dst += pathSeperator
	}

	return
}

func printHelp() {
	fmt.Println("HELP PAGE")
	fmt.Println("-src [Path] - enter Discord path *necessary")
	fmt.Println("-dst [Path] - enter path to save")
	fmt.Println("-cs [Num] - enter how big the chunk for each thread should be")
	fmt.Println("-tc [Num] - How many threads should run at once")
	fmt.Println("-k - keep files with unknown filetype")
	fmt.Println("-s [/ or \\\\] - seperator used by your file system")
}

func getContentInNextQuotes(argSlice []string, startPoint int) (content string) {
	startRead := false

	for i := startPoint; i < len(argSlice); i++ {
		if argSlice[i][0] == '"' {
			if !startRead {
				startRead = true
			}
		}
		if startRead {
			content += argSlice[i]
			fmt.Printf("Added %v to content", argSlice[i])
		}
		if argSlice[i][:1] == "\"" {
			content = strings.Replace(content, "\"", "", -1)
			return
		}
	}
	fmt.Println("No path in quotes found!")
	return
}

func fileArrayCopy(files []os.FileInfo, dst string, orig string, name string, startName int, keepUnknownFileTypes bool) {

	defer wg.Done()

	for _, f := range files {
		if file, err := os.Open(orig + f.Name()); err != nil {
			log.Printf("Can not open file: %v \n Error: %v", f, err)

		} else {
			if fileTypeRaw, err := getFileContentType(file); err != nil {
				log.Printf("Unable to get filetype of: %v \n Error: %v", f, err)
			} else {

				fileTypeRaw = strings.ToLower(fileTypeRaw) // image/png
				fileType := strings.Split(fileTypeRaw, "/")[1]

				if fileType != "octet-stream" || keepUnknownFileTypes { // Skip Files with unknown type
					if err = copyFiles(orig+f.Name(), dst+name+strconv.Itoa(startName)+"."+fileType); err != nil {
						log.Printf("Unable to copy: %v \n Error: %v", f, err)
					} else {
						startName++
					}
				}
			}
			file.Close()
		}
	}
}

func copyFiles(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

func getFileContentType(out *os.File) (string, error) {

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err := out.Read(buffer)
	if err != nil {
		return "", err
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)

	return contentType, nil
}

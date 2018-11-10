package main

import (
	"flag"
	"github.com/chai2010/tiff"
	"github.com/pkg/errors"
	"golang.org/x/image/bmp"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func main() {
	target := flag.String("target", "./", "the directory to traverse, or file to convert")
	fromSuffix := flag.String("from", "tiff", "the suffix to search for")
	toSuffix := flag.String("to", "jpg", "the suffix to write to")
	quitOnError := flag.Bool("quitOnError", true, "whether to stop if an error happens")
	deleteOriginal := flag.Bool("deleteOriginal", false, "whether the original files are deleted")
	flag.Parse()

	if err := traverse(*target, *fromSuffix, *toSuffix, *quitOnError, *deleteOriginal); err != nil {
		log.Fatal(err)
	}
}

func traverse(target, from, to string, quitOnError bool, deleteOriginal bool) error {
	stat, err := os.Stat(target)
	if err != nil {
		if quitOnError {
			return err
		} else {
			log.Printf("A non-fatal error occurred: " + err.Error())
			return nil
		}
	}

	if stat.IsDir() {

		// traverse contents, calling this function on all sub-dirs
		fileInfos, err := ioutil.ReadDir(target)
		if err != nil {
			if quitOnError {
				return err
			} else {
				log.Printf("A non-fatal error occurred: " + err.Error())
				return nil
			}
		}

		for _, f := range fileInfos {
			if f.IsDir() {
				traverse(target + string(os.PathSeparator) + f.Name(), from, to, quitOnError, deleteOriginal)
			} else {
				if err := convertFileIfMatch(target + string(os.PathSeparator) + f.Name(), from, to, quitOnError, deleteOriginal); err != nil {
					if quitOnError {
						return err
					} else {
						log.Printf("A non-fatal error occurred: " + err.Error())
					}
				}
			}
		}

		return nil
	}

	// stat is of a file, convert it if needed
	return convertFileIfMatch(target, from, to, quitOnError, deleteOriginal)
}

func convertFileIfMatch(name string, from string, to string, quitOnError bool, deleteOriginal bool) error {
	if strings.HasSuffix(strings.ToUpper(name), strings.ToUpper(from)) {
		indexOfLast := strings.LastIndex(strings.ToUpper(name), strings.ToUpper(from))

		if indexOfLast != -1 {
			newName := name[0:indexOfLast] + to

			file, err := os.Open(name)
			if err != nil {
				if quitOnError{
					return err
				} else {
					log.Printf("while opening output file: " + err.Error())
				}
			}
			defer file.Close()
			file.Seek(0, 0)

			imageData, _, err := image.Decode(file)
			if err != nil {
				return err
			}
			file.Close()

			outFile, err := openOrCreate(newName)
			if quitOnError && err != nil {
				return err
			} else if err != nil {
				log.Printf("while creating file: " + err.Error())
			}
			defer outFile.Close()

			if strings.ToUpper(to) == "JPG" {
				err := jpeg.Encode(outFile, imageData, nil)
				if quitOnError && err != nil {
					return err
				} else if err != nil {
					log.Printf("while saving jpeg: " + err.Error())
				}
			} else if strings.ToUpper(to) == "PNG" {
				err := png.Encode(outFile, imageData)
				if quitOnError && err != nil {
					return err
				} else if err != nil {
					log.Printf("while saving png: " + err.Error())
				}
			} else if strings.ToUpper(to) == "GIF" {
				err := gif.Encode(outFile, imageData, nil)
				if quitOnError && err != nil {
					return err
				} else if err != nil {
					log.Printf("while saving gif: " + err.Error())
				}
			} else if strings.ToUpper(to) == "BMP" {
				err := bmp.Encode(outFile, imageData)
				if quitOnError && err != nil {
					return err
				} else if err != nil {
					log.Printf("while saving gif: " + err.Error())
				}
			} else if strings.ToUpper(to) == "TIFF" {
				err := tiff.Encode(outFile, imageData, nil)
				if quitOnError && err != nil {
					return err
				} else if err != nil {
					log.Printf("while saving tiff: " + err.Error())
				}
			} else {
				if quitOnError && err != nil {
					return errors.New("unknown suffix: " + to)
				} else if err != nil {
					log.Printf("unknown suffix: " + to)
				}
			}

			if deleteOriginal {
				outFile.Close()
				err := os.Remove(name)

				if quitOnError && err != nil {
					return err
				} else if err != nil {
					log.Printf("while deleting original file " + name + ", error = " + err.Error())
				}
			}
		}
	}

	return nil
}

func openOrCreate(filename string) (*os.File, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return os.Create(filename)
	}
	return os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC, 066)
}
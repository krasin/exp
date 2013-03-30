// Cascade classifier accepts a cascade and an image and detects all the objects
// on the image using this cascade.
// For more details, see
// YEF Yet Even Faster Real-Time Object Detection, 2007 by Yotam Abramson and Bruno Steux
package main

import (
	"flag"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"
	"path"

	"github.com/deboshire/exp/detect/cascade_classifier/format"
)

var cascadePath = flag.String("cascade", "", "Path to the cascade classifier")
var inputPath = flag.String("input", "", "Input image")
var outputPath = flag.String("output", "", "Output image")

func MustLoadImage(path string) *image.Gray {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("Coudn't open %s: %v", path, err)
	}
	defer f.Close()
	var m image.Image
	m, _, err = image.Decode(f)
	g := image.NewGray(m.Bounds())
	draw.Draw(g, g.Bounds(), m, image.Point{0, 0}, draw.Src)
	return g
}

func MustLoadCascade(path string) (cascade *format.Cascade) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("Couldn't open: %s: %v", path, err)
	}
	defer f.Close()
	if cascade, err = format.LoadJson(f); err != nil {
		log.Fatalf("Failed to load cascade from %s: %v", path, err)
	}
	return
}

func MustSaveImage(destPath, inputPath string, img image.Image) {
	if destPath == "" {
		destPath = path.Clean(inputPath) + ".gray.png"
	}
	f, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("Could not save output image to %s: %v", destPath, err)
	}
	defer f.Close()
	if err = png.Encode(f, img); err != nil {
		log.Fatalf("Error while saving the output image to %s: %v", destPath, err)
	}
}

func Detect(img *image.Gray) []image.Point {
	return nil
}

func DetectAll(img *image.Gray, cascade *format.Cascade) (res *image.Gray) {
	res = image.NewGray(img.Bounds())
	draw.Draw(res, res.Bounds(), img, image.Point{0, 0}, draw.Src)
	return
}

func main() {
	flag.Parse()
	if *cascadePath == "" || *inputPath == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	input := MustLoadImage(*inputPath)
	cascade := MustLoadCascade(*cascadePath)
	output := DetectAll(input, cascade)

	MustSaveImage(*outputPath, *inputPath, output)
}

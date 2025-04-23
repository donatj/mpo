package main

import (
	"flag"
	"fmt"
	"image"
	"os"

	"github.com/donatj/mpo"

	_ "image/gif"
	"image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

var (
	output  = flag.String("outfile", "output.mpo", "Output filename")
	quality = flag.Int("quality", 90, "JPEG quality [0-100]")
	help    = flag.Bool("help", false, "Displays this text")
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <imagefile> [<imagefile> ...]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Convert one or more images to a Multi-Picture Object (MPO) file.\n\n")
		fmt.Fprintf(os.Stderr, "Supported image formats: JPEG, PNG, GIF, BMP, TIFF, WebP\n\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: At least one image file is required - ideally two.")
		flag.Usage()
		os.Exit(2)
	}
}

func main() {
	images := make([]image.Image, 0, flag.NArg())
	for _, arg := range flag.Args() {
		f, err := os.Open(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening %s: %v\n", arg, err)
			os.Exit(1)
		}
		defer f.Close()

		img, _, err := image.Decode(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error decoding %s: %v\n", arg, err)
			os.Exit(1)
		}

		images = append(images, img)
	}

	if len(images) == 0 {
		fmt.Fprintln(os.Stderr, "No images to encode")
		os.Exit(1)
	}

	f, err := os.Create(*output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	err = mpo.EncodeAll(f, &mpo.MPO{Image: images}, &jpeg.Options{Quality: *quality})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding MPO: %v\n", err)
		os.Exit(1)
	}
}

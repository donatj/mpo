package main

import (
	"flag"
	"fmt"
	"github.com/donatj/mpo"
	"image"
	"image/jpeg"
	"log"
	"os"
)

var (
	format = flag.String("format", "stereo", "Output format [stereo|red-cyan|cyan-red|red-green|green-red]")
	output = flag.String("outfile", "output.jpg", "Output filename")
	help   = flag.Bool("help", false, "Displays this text")
)

func init() {
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Println("usage: mpo2img <mpofile>")
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func main() {
	m, err := mpo.DecodeAll(flag.Arg(0))
	if err != nil {
		log.Fatalf("err on %v %s", err, flag.Arg(0))
	}

	var img image.Image
	switch *format {
	case "stereo":
		img = m.ConvertToStereo()
	case "red-cyan":
		img, err = m.ConvertToAnaglyph(mpo.RedCyan)
		if err != nil {
			log.Fatal(err)
		}
	case "cyan-red":
		img, err = m.ConvertToAnaglyph(mpo.CyanRed)
		if err != nil {
			log.Fatal(err)
		}
	case "red-green":
		img, err = m.ConvertToAnaglyph(mpo.RedGreen)
		if err != nil {
			log.Fatal(err)
		}
	case "green-red":
		img, err = m.ConvertToAnaglyph(mpo.GreenRed)
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal("Unknown format:", *format)
	}

	f, err := os.OpenFile(*output, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	if err = jpeg.Encode(f, img, nil); err != nil {
		log.Fatal(err)
	}
}

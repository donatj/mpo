# MPO Encoder/Decoder Library

[![Go Report Card](https://goreportcard.com/badge/donatj/mpo)](https://goreportcard.com/report/donatj/mpo)
[![GoDoc](https://godoc.org/github.com/donatj/mpo?status.svg)](https://godoc.org/github.com/donatj/mpo)
[![awesome-go](https://img.shielded.dev/s?title=listed%20on&text=awesome-go&color=blue)](https://github.com/avelino/awesome-go)

Simple Go JPEG MPO (Multi Picture Object) Decoder and Encoder - Library and CLI Tool

The library and CLI can:

- **Decode** an MPO into individual JPEG frames.
- **Encode** multiple JPEG frames into a Baseline-MP MPO.
- **Convert** an MPO to a stereoscopic (side-by-side) JPEG.
- **Create** anaglyph images (red–cyan, cyan–red, red–green, green–red).

A Web UI for converting MPO to JPEG is available at:

https://donatstudios.com/MPO-to-JPEG-Stereo

## Install CLI Tool

Binaries are available for Darwin (macOS), Linux and Windows on the release page:

https://github.com/donatj/mpo/releases

### From Source

```bash
go install github.com/donatj/mpo/cmd/mpo2img@latest
go install github.com/donatj/mpo/cmd/img2mpo@latest
```

## CLI Usage

### mpo2img

Convert an MPO file to a stereoscopic JPEG or anaglyph image.

```
$ mpo2img -help
Usage: mpo2img <mpofile>

Convert a Multi-Picture Object (MPO) file to an image.

  -format string
        Output format [stereo|red-cyan|cyan-red|red-green|green-red] (default "stereo")
  -help
        Displays this text
  -outfile string
        Output filename (default "output.jpg")
```

### img2mpo

encode multiple images into an MPO file.

```
$ img2mpo -help
Usage: img2mpo <imagefile> [<imagefile> ...]

Convert one or more images to a Multi-Picture Object (MPO) file.

Supported image formats: JPEG, PNG, GIF, BMP, TIFF, WebP

  -help
        Displays this text
  -outfile string
        Output filename (default "output.mpo")
  -quality int
        JPEG quality [0-100] (default 90)
```

## WIP

Todo:

- Optimization
- Add more control over stereo/anaglyph

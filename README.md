# MPO Decoder Library

[![Go Report Card](http://goreportcard.com/badge/donatj/mpo)](http://goreportcard.com/report/donatj/mpo)
[![GoDoc](https://godoc.org/github.com/donatj/mpo?status.svg)](https://godoc.org/github.com/donatj/mpo)

Simple Go JPEG MPO (Multi Picture Object) Decoder

## Install Sample Command Line Tool

```bash
go get github.com/donatj/mpo
go install github.com/donatj/mpo/mpo2img
```

Usage

```
mpo2img
usage: mpo2img <mpofile>
  -format string
    	Output format [stereo|red-cyan|cyan-red|red-green|green-red] (default "stereo")
  -help
    	Displays this text
  -outfile string
    	Output filename (default "output.jpg")
```

## WIP

Todo:
- Optimization
- Add Writer
- Add more control over stereo/anaglyph

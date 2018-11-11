# MPO Decoder Library

[![Go Report Card](https://goreportcard.com/badge/donatj/mpo)](https://goreportcard.com/report/donatj/mpo)
[![GoDoc](https://godoc.org/github.com/donatj/mpo?status.svg)](https://godoc.org/github.com/donatj/mpo)

Simple Go JPEG MPO (Multi Picture Object) Decoder - Library and CLI Tool

## Install CLI Tool

Binaries are availible for Darwin (macOS), Linux and Windows on the release page:

https://github.com/donatj/mpo/releases

### From Source

```bash
go get -u github.com/donatj/mpo/cmd/mpo2img
```

## CLI Usage

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

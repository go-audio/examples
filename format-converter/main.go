package main

import (
	"flag"
	"fmt"
	"os"

	"strings"

	"github.com/go-audio/aiff"
	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

var (
	flagInput  = flag.String("input", "", "The file to convert")
	flagFormat = flag.String("format", "aiff", "The format to convert to (wav or aiff)")
	flagOutput = flag.String("output", "out", "The output filename")
)

func main() {
	flag.Parse()
	if *flagInput == "" {
		fmt.Println("Provide an input file using the -input flag")
		os.Exit(1)
	}
	switch strings.ToLower(*flagFormat) {
	case "aiff", "aif":
		*flagFormat = "aiff"
	case "wave", "wav":
		*flagFormat = "wav"
	default:
		fmt.Println("Provide a valid -format flag")
		os.Exit(1)
	}
	f, err := os.Open(*flagInput)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	type decoder interface {
		FullPCMBuffer() (*audio.IntBuffer, error)
	}
	var dec decoder

	type encoder interface {
		Write(b *audio.IntBuffer) error
		Close() error
	}
	var enc encoder

	var buf *audio.IntBuffer
	var bitDepth int

	wd := wav.NewDecoder(f)
	if wd.IsValidFile() {
		bitDepth = int(wd.BitDepth)
		dec = wd
	} else {
		f.Seek(0, 0)
		aiffd := aiff.NewDecoder(f)
		if aiffd.IsValidFile() {
			bitDepth = int(wd.BitDepth)
		}
		dec = aiffd
	}

	// TODO: switch to encode/decode in chunks
	buf, err = dec.FullPCMBuffer()
	if err != nil {
		panic(err)
	}

	outputFilename := fmt.Sprintf("%s.%s", *flagOutput, *flagFormat)
	of, err := os.Create(outputFilename)
	if err != nil {
		panic(err)
	}

	fmt.Println("File format", buf.Format.SampleRate, bitDepth, buf.Format.NumChannels)

	if *flagFormat == "aiff" {
		enc = aiff.NewEncoder(of, buf.Format.SampleRate, bitDepth, buf.Format.NumChannels)
	} else {
		enc = wav.NewEncoder(of,
			buf.Format.SampleRate,
			bitDepth,
			buf.Format.NumChannels,
			1)
	}

	if err := enc.Write(buf); err != nil {
		panic(err)
	}
	enc.Close()
	of.Close()
	fmt.Println(outputFilename, "written to disk")
}

package main

import (
	"flag"
	"fmt"
	"os"

	"strings"

	"github.com/go-audio/aiff"
	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/mattetti/audio/decoder"
)

var (
	flagInput  = flag.String("input", "", "The file to convert")
	flagFormat = flag.String("format", "wav", "The format to convert to (wav or aiff)")
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

	var dec decoder.Decoder

	type encoder interface {
		Write(b *audio.IntBuffer) error
		Close() error
	}
	var enc encoder

	var buf *audio.IntBuffer

	wd := wav.NewDecoder(f)
	if wd.IsValidFile() {
		dec = wd
	} else {
		f.Seek(0, 0)
		aiffd := aiff.NewDecoder(f)
		if !aiffd.IsValidFile() {
			fmt.Println("input file isn't a valid wav or aiff file")
			os.Exit(1)
		}
		dec = aiffd
	}

	if !dec.WasPCMAccessed() {
		err := dec.FwdToPCM()
		if err != nil {
			panic(err)
		}
	}

	format := dec.Format()

	var of *os.File
	outputFilename := fmt.Sprintf("%s.%s", *flagOutput, *flagFormat)
	switch *flagFormat {
	case "wav", ".wav", "wave":
		of, err = os.Create(outputFilename)
		if err != nil {
			panic(err)
		}
		defer of.Close()
		enc = wav.NewEncoder(of, format.SampleRate, int(dec.SampleBitDepth()), format.NumChannels, 1)
	case "aif", ".aif", "aiff", ".aiff":
		of, err = os.Create(outputFilename)
		if err != nil {
			panic(err)
		}
		defer of.Close()
		enc = aiff.NewEncoder(of, format.SampleRate, int(dec.SampleBitDepth()), format.NumChannels)
	default:
		fmt.Printf("output format %s not supported\n", *flagFormat)
		os.Exit(1)
	}

	buf = &audio.IntBuffer{Format: format, Data: make([]int, 4096)}

	var t int64
	var n int
	for err == nil && t <= dec.PCMLen() {
		n, err = dec.PCMBuffer(buf)
		if err != nil {
			fmt.Println("failed to read the input file -", err)
			os.Exit(1)
		}
		t += int64(n)
		if n == 0 {
			break
		}
		if n != len(buf.Data) {
			buf.Data = buf.Data[:n]
		}
		if err = enc.Write(buf); err != nil {
			fmt.Println("failed to write to the output file -", err)
			os.Exit(1)
		}
		if n != len(buf.Data) {
			break
		}
	}

	if err = enc.Close(); err != nil {
		fmt.Println("failed to close the encoder stream")
		os.Exit(1)
	}
	of.Close()
	fmt.Println(outputFilename, "written to disk")
}

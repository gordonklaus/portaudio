package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/signal"

	"github.com/bobertlo/go-mpg123/mpg123"
	"github.com/gordonklaus/portaudio"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("missing required argument:  input file name")
		return
	}
	fmt.Println("Playing.  Press Ctrl-C to stop.")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	// create mpg123 decoder instance
	decoder, err := mpg123.NewDecoder("")
	chk(err)

	fileName := os.Args[1]
	chk(decoder.Open(fileName))
	defer decoder.Close()

	// get audio format information
	rate, channels, _ := decoder.GetFormat()

	// make sure output format does not change
	decoder.FormatNone()
	decoder.Format(rate, channels, mpg123.ENC_SIGNED_16)

	portaudio.Initialize()
	defer portaudio.Terminate()
	out := make([]int16, 8192)
	stream, err := portaudio.OpenDefaultStream(0, channels, float64(rate), len(out), &out)
	chk(err)
	defer stream.Close()

	chk(stream.Start())
	defer stream.Stop()
	for {
		audio := make([]byte, 2*len(out))
		_, err = decoder.Read(audio)
		if err == mpg123.EOF {
			break
		}
		chk(err)

		chk(binary.Read(bytes.NewBuffer(audio), binary.LittleEndian, out))
		chk(stream.Write())
		select {
		case <-sig:
			return
		default:
		}
	}
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}

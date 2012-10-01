package main

import (
	"code.google.com/p/portaudio-go/portaudio"
	"math/rand"
	"time"
)

func main() {
	chk := func(err error) { if err != nil { panic(err) } }
	stream, err := portaudio.OpenDefaultStream(0, 1, 44100, 128, noiseGenerator{})
	chk(err)
	defer stream.Close()
	chk(stream.Start())
	time.Sleep(1e9)
	chk(stream.Stop())
}

type noiseGenerator struct{}
func (noiseGenerator) ProcessAudio(_, out []uint8) {
	for i := range out {
		out[i] = uint8(rand.Uint32())
	}
}

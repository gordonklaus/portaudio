package main

import (
	"code.google.com/p/portaudio-go/portaudio"
	"time"
	"math/rand"
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
func (noiseGenerator) ProcessAudio(_, out [][]float32) {
	for i := range out[0] {
		out[0][i] = .1 * (2 * rand.Float32() - 1)
	}
}

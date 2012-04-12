package main

import (
	"code.google.com/p/portaudio-go/portaudio"
	"time"
	"math/rand"
)

func main() {
	chk := func(err error) { if err != nil { panic(err) } }
	stream, err := portaudio.OpenDefaultStream(1, 1, 44100, 128, new(NoiseGenerator))
	chk(err)
	defer stream.Close()
	chk(stream.Start())
	time.Sleep(1e9)
	chk(stream.Stop())
}

type NoiseGenerator int
func (*NoiseGenerator) ProcessAudio(inputBuffer, outputBuffer []float32) {
	for i := range outputBuffer {
		outputBuffer[i] = .1*(2*rand.Float32() - 1)
	}
}

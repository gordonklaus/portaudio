package main

import "portaudio-go.googlecode.com/hg/portaudio"
import (
	"time"
	"rand"
)

func main() {
	stream, err := portaudio.OpenDefaultStream(1, 1, 44100, 128, callback)
	if err != nil { panic(err.Text) }
	defer stream.Close()
	err = stream.Start()
	if err != nil { panic(err.Text) }
	time.Sleep(1e9)
	err = stream.Stop()
	if err != nil { panic(err.Text) }
}

func callback(inputBuffer, outputBuffer []float32) {
	for i := range outputBuffer {
		outputBuffer[i] = .1*(2*rand.Float32() - 1)
	}
}

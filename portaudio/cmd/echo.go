package main

import "portaudio-go.googlecode.com/hg/portaudio"
import (
	"time"
)

func main() {
	stream, err := portaudio.OpenDefaultStream(1, 1, 44100, 4096, callback)
	if err != nil { panic(err.Text) }
	defer stream.Close()
	err = stream.Start()
	if err != nil { panic(err.Text) }
	time.Sleep(4e9)
	err = stream.Stop()
	if err != nil { panic(err.Text) }
}

var buffer []float32 = make([]float32, 4096)
func callback(inputBuffer, outputBuffer []float32) {
	for i := range outputBuffer {
		outputBuffer[i] = .7*buffer[i]
	}
	buffer = inputBuffer
}

package main

import (
	"code.google.com/p/portaudio-go/portaudio"
	"time"
)

func main() {
	chk := func(err error) { if err != nil { panic(err) } }
	bufferSize := 4096
	stream, err := portaudio.OpenDefaultStream(1, 1, 44100, bufferSize, &Echoer{make([]float32, bufferSize)})
	chk(err)
	defer stream.Close()
	chk(stream.Start())
	time.Sleep(4e9)
	chk(stream.Stop())
}

type Echoer struct {
	buffer []float32
}

func (e *Echoer) ProcessAudio(inputBuffer, outputBuffer []float32) {
	for i := range outputBuffer {
		outputBuffer[i] = .7*e.buffer[i]
	}
	copy(e.buffer, inputBuffer)
}

package main

import (
	"code.google.com/p/portaudio-go/portaudio"
	"time"
)

func main() {
	stream, err := portaudio.OpenDefaultStream(1, 1, 44100, 4096, NewEchoer())
	if err != nil { panic(err.Text) }
	defer stream.Close()
	err = stream.Start()
	if err != nil { panic(err.Text) }
	time.Sleep(4e9)
	err = stream.Stop()
	if err != nil { panic(err.Text) }
}

type Echoer struct {
	buffer []float32
}

func NewEchoer() *Echoer {
	return &Echoer{make([]float32, 4096)}
}

func (e *Echoer) ProcessAudio(inputBuffer, outputBuffer []float32) {
	for i := range outputBuffer {
		outputBuffer[i] = .7*e.buffer[i]
	}
	copy(e.buffer, inputBuffer)
}

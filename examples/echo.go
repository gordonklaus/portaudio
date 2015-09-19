package main

import (
	"github.com/gordonklaus/portaudio"
	"time"
)

func main() {
	portaudio.Initialize()
	defer portaudio.Terminate()
	e := newEcho(time.Second / 3)
	defer e.Close()
	chk(e.Start())
	time.Sleep(4 * time.Second)
	chk(e.Stop())
}

type echo struct {
	*portaudio.Stream
	buffer []float32
	i      int
}

func newEcho(delay time.Duration) *echo {
	h, err := portaudio.DefaultHostApi()
	chk(err)
	p := portaudio.LowLatencyParameters(h.DefaultInputDevice, h.DefaultOutputDevice)
	p.Input.Channels = 1
	p.Output.Channels = 1
	e := &echo{buffer: make([]float32, int(p.SampleRate*delay.Seconds()))}
	e.Stream, err = portaudio.OpenStream(p, e.processAudio)
	chk(err)
	return e
}

func (e *echo) processAudio(in, out []float32) {
	for i := range out {
		out[i] = .7 * e.buffer[e.i]
		e.buffer[e.i] = in[i]
		e.i = (e.i + 1) % len(e.buffer)
	}
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}

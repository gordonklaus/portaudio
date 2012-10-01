package main

import (
	"code.google.com/p/portaudio-go/portaudio"
	"math"
	"time"
)

const sampleRate = 44100

func main() {
	chk := func(err error) { if err != nil { panic(err) } }
	stream, err := portaudio.OpenDefaultStream(0, 2, sampleRate, 0, newStereoSine(256, 320, sampleRate))
	chk(err)
	defer stream.Close()
	chk(stream.Start())
	time.Sleep(2 * time.Second)
	chk(stream.Stop())
}

type stereoSine struct {
	stepL, phaseL float64
	stepR, phaseR float64
}
func newStereoSine(freqL, freqR, sampleRate float64) *stereoSine { return &stereoSine{freqL / sampleRate, 0, freqR / sampleRate, 0} }
func (g *stereoSine) ProcessAudio(_, out [][]float32) {
	for i := range out[0] {
		out[0][i] = float32(math.Sin(2 * math.Pi * g.phaseL))
		_, g.phaseL = math.Modf(g.phaseL + g.stepL)
		out[1][i] = float32(math.Sin(2 * math.Pi * g.phaseR))
		_, g.phaseR = math.Modf(g.phaseR + g.stepR)
	}
}

package portaudio

/*
#cgo LDFLAGS: -lportaudio
#include "pa.h"
*/
import "C"

import (
	"code.google.com/p/rog-go/exp/callback"
	"unsafe"
	"reflect"
	"errors"
)

func init() {
	C.setCallbackFunc(callback.Func)
}

// An AudioProcessor reads input buffers and writes output buffers.
// len(inputs) == numInputChannels, len(outputs) == numOutputChannels,
// and len(inputs[i]) == len(outputs[j]) == framesPerBuffer
type AudioProcessor interface {
	ProcessAudio(inputs, outputs [][]float32)
}

type Stream struct {
	numInputChannels, numOutputChannels int
	paStream unsafe.Pointer
	closed bool
	audioProcessor AudioProcessor
}

func newError(err C.PaError) error {
	return errors.New(C.GoString(C.Pa_GetErrorText(err)))
}

func OpenDefaultStream(numInputChannels, numOutputChannels int,
						sampleRate float64, framesPerBuffer int,
						audioProcessor AudioProcessor) (*Stream, error) {
	err := C.Pa_Initialize()
	if err != C.paNoError {
		return nil, newError(err)
	}
	
	stream := &Stream{numInputChannels, numOutputChannels, nil, false, audioProcessor}
	err = C.Pa_OpenDefaultStream(&stream.paStream, C.int(numInputChannels), C.int(numOutputChannels), C.paFloat32 | C.paNonInterleaved, C.double(sampleRate), C.ulong(framesPerBuffer), C.getPaStreamCallback(), unsafe.Pointer(stream))
	if err != C.paNoError {
		return nil, newError(err)
	}
	return stream, nil
}

func (s *Stream) Start() error {
	err := C.Pa_StartStream(s.paStream)
	if err != C.paNoError {
		return newError(err)
	}
	return nil
}

//export streamCallback
func streamCallback(arg unsafe.Pointer) {
	context := (*C.context)(arg)
	stream := (*Stream)(context.stream)
	frameCount := (int)(context.frameCount)
	stream.audioProcessor.ProcessAudio(channels(context.inputBuffer, stream.numInputChannels, frameCount), channels(context.outputBuffer, stream.numOutputChannels, frameCount))
}

func channels(buffers unsafe.Pointer, numChans int, frameCount int) [][]float32 {
	bufs := (*[1<<29 - 1]unsafe.Pointer)(buffers)
	c := make([][]float32, numChans)
	for i := 0; i < numChans; i++ {
		c[i] = sliceAt(bufs[i], frameCount)
	}
	return c
}

func sliceAt(buffer unsafe.Pointer, size int) []float32 {
	if buffer == nil { return nil }
	slice := (*[1<<29 - 1]float32)(buffer)[:size]
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
	sliceHeader.Cap = size
	return slice
}

func (s *Stream) Stop() error {
	err := C.Pa_StopStream(s.paStream)
	if err != C.paNoError {
		return newError(err)
	}
	return nil
}

func (s *Stream) Close() error {
	if !s.closed {
		s.closed = true
		err := C.Pa_Terminate()
		if err != C.paNoError {
			return newError(err)
		}
	}
	return nil
}

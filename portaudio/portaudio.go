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

type AudioProcessor interface {
	ProcessAudio(inputBuffer, outputBuffer []float32)
}

type Stream struct {
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
	
	stream := &Stream{audioProcessor:audioProcessor}
	err = C.Pa_OpenDefaultStream(&stream.paStream, C.int(numInputChannels), C.int(numOutputChannels), C.paFloat32, C.double(sampleRate), C.ulong(framesPerBuffer), C.getPaStreamCallback(), unsafe.Pointer(stream))
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

func sliceAt(buffer unsafe.Pointer, size int) []float32 {
	if buffer == nil { return nil }
	slice := (*[1<<29 - 1]float32)(buffer)[:size]
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
	sliceHeader.Cap = size
	return slice
}

//export streamCallback
func streamCallback(arg unsafe.Pointer) {
	context := (*C.context)(arg)
	stream := (*Stream)(context.stream)
	frameCount := (int)(context.frameCount)
	stream.audioProcessor.ProcessAudio(sliceAt(context.inputBuffer, frameCount), sliceAt(context.outputBuffer, frameCount))
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

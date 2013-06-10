package portaudio

// #cgo LDFLAGS: -lportaudio
// #include "pa.h"
import "C"

import (
	"errors"
	."fmt"
	"reflect"
	"unsafe"
)

type Int24 [3]byte

type Stream struct {
	paStream unsafe.Pointer
	closed bool
	in, out unsafe.Pointer
	callback func(pin, pout uintptr, n int)
}

func newError(err C.PaError) error {
	return errors.New(C.GoString(C.Pa_GetErrorText(err)))
}

/*
Opens the default input and/or output devices.

audioProcessor must have a method
	ProcessAudio(in, out [][]SampleType)
or
	ProcessAudio(in, out []SampleType)
where SampleType is float32, int32, Int24, int16, int8, or uint8.

In the former case, channels are non-interleaved:
len(in) == numInputChannels, len(out) == numOutputChannels,
and len(in[i]) == len(out[j]) == framesPerBuffer.

In the latter case, channels are interleaved:
len(in) == numInputChannels * framesPerBuffer and len(out) == numOutputChannels * framesPerBuffer.
*/
func OpenDefaultStream(numInputChannels, numOutputChannels int,
						sampleRate float64, framesPerBuffer int,
						audioProcessor interface{}) (*Stream, error) {
	paErr := C.Pa_Initialize()
	if paErr != C.paNoError {
		return nil, newError(paErr)
	}
	
	s := &Stream{}
	fmt, err := s.init(audioProcessor, numInputChannels, numOutputChannels)
	if err != nil {
		return nil, err
	}
	paErr = C.Pa_OpenDefaultStream(&s.paStream, C.int(numInputChannels), C.int(numOutputChannels), fmt, C.double(sampleRate), C.ulong(framesPerBuffer), C.getPaStreamCallback(), unsafe.Pointer(s))
	if paErr != C.paNoError {
		return nil, newError(paErr)
	}
	return s, nil
}

func (s *Stream) init(audioProcessor interface{}, numIn, numOut int) (fmt C.PaSampleFormat, err error) {
	switch p := audioProcessor.(type) {
	case interface{ProcessAudio(in, out [][]float32)}:
		in, out := make([][]float32, numIn), make([][]float32, numOut)
		s.in, s.out = unsafe.Pointer(&in), unsafe.Pointer(&out)
		s.callback = func(pin, pout uintptr, n int) {
			s.updateSlices(pin, pout, n)
			p.ProcessAudio(in, out)
		}
		fmt = C.paFloat32 | C.paNonInterleaved
	case interface{ProcessAudio(in, out []float32)}:
		var in, out []float32
		s.in, s.out = unsafe.Pointer(&in), unsafe.Pointer(&out)
		s.callback = func(pin, pout uintptr, n int) {
			s.updateSlicesInterleaved(pin, pout, n, numIn, numOut)
			p.ProcessAudio(in, out)
		}
		fmt = C.paFloat32
	case interface{ProcessAudio(in, out [][]int32)}:
		in, out := make([][]int32, numIn), make([][]int32, numOut)
		s.in, s.out = unsafe.Pointer(&in), unsafe.Pointer(&out)
		s.callback = func(pin, pout uintptr, n int) {
			s.updateSlices(pin, pout, n)
			p.ProcessAudio(in, out)
		}
		fmt = C.paInt32 | C.paNonInterleaved
	case interface{ProcessAudio(in, out []int32)}:
		var in, out []int32
		s.in, s.out = unsafe.Pointer(&in), unsafe.Pointer(&out)
		s.callback = func(pin, pout uintptr, n int) {
			s.updateSlicesInterleaved(pin, pout, n, numIn, numOut)
			p.ProcessAudio(in, out)
		}
		fmt = C.paInt32
	case interface{ProcessAudio(in, out [][]Int24)}:
		in, out := make([][]Int24, numIn), make([][]Int24, numOut)
		s.in, s.out = unsafe.Pointer(&in), unsafe.Pointer(&out)
		s.callback = func(pin, pout uintptr, n int) {
			s.updateSlices(pin, pout, n)
			p.ProcessAudio(in, out)
		}
		fmt = C.paInt24 | C.paNonInterleaved
	case interface{ProcessAudio(in, out []Int24)}:
		var in, out []Int24
		s.in, s.out = unsafe.Pointer(&in), unsafe.Pointer(&out)
		s.callback = func(pin, pout uintptr, n int) {
			s.updateSlicesInterleaved(pin, pout, n, numIn, numOut)
			p.ProcessAudio(in, out)
		}
		fmt = C.paInt24
	case interface{ProcessAudio(in, out [][]int16)}:
		in, out := make([][]int16, numIn), make([][]int16, numOut)
		s.in, s.out = unsafe.Pointer(&in), unsafe.Pointer(&out)
		s.callback = func(pin, pout uintptr, n int) {
			s.updateSlices(pin, pout, n)
			p.ProcessAudio(in, out)
		}
		fmt = C.paInt16 | C.paNonInterleaved
	case interface{ProcessAudio(in, out []int16)}:
		var in, out []int16
		s.in, s.out = unsafe.Pointer(&in), unsafe.Pointer(&out)
		s.callback = func(pin, pout uintptr, n int) {
			s.updateSlicesInterleaved(pin, pout, n, numIn, numOut)
			p.ProcessAudio(in, out)
		}
		fmt = C.paInt16
	case interface{ProcessAudio(in, out [][]int8)}:
		in, out := make([][]int8, numIn), make([][]int8, numOut)
		s.in, s.out = unsafe.Pointer(&in), unsafe.Pointer(&out)
		s.callback = func(pin, pout uintptr, n int) {
			s.updateSlices(pin, pout, n)
			p.ProcessAudio(in, out)
		}
		fmt = C.paInt8 | C.paNonInterleaved
	case interface{ProcessAudio(in, out []int8)}:
		var in, out []int8
		s.in, s.out = unsafe.Pointer(&in), unsafe.Pointer(&out)
		s.callback = func(pin, pout uintptr, n int) {
			s.updateSlicesInterleaved(pin, pout, n, numIn, numOut)
			p.ProcessAudio(in, out)
		}
		fmt = C.paInt8
	case interface{ProcessAudio(in, out [][]uint8)}:
		in, out := make([][]uint8, numIn), make([][]uint8, numOut)
		s.in, s.out = unsafe.Pointer(&in), unsafe.Pointer(&out)
		s.callback = func(pin, pout uintptr, n int) {
			s.updateSlices(pin, pout, n)
			p.ProcessAudio(in, out)
		}
		fmt = C.paUInt8 | C.paNonInterleaved
	case interface{ProcessAudio(in, out []uint8)}:
		var in, out []uint8
		s.in, s.out = unsafe.Pointer(&in), unsafe.Pointer(&out)
		s.callback = func(pin, pout uintptr, n int) {
			s.updateSlicesInterleaved(pin, pout, n, numIn, numOut)
			p.ProcessAudio(in, out)
		}
		fmt = C.paUInt8
	default:
		err = errors.New(Sprintf("%T lacks a supported ProcessAudio method", audioProcessor))
	}
	return
}

func (s *Stream) updateSlices(pin, pout uintptr, frames int) {
	setChannels(s.in, pin, frames)
	setChannels(s.out, pout, frames)
}

func (s *Stream) updateSlicesInterleaved(pin, pout uintptr, frames int, numIn, numOut int) {
	setSlice(s.in, pin, frames * numIn)
	setSlice(s.out, pout, frames * numOut)
}

func setChannels(slice unsafe.Pointer, p uintptr, frames int) {
	s := (*reflect.SliceHeader)(slice)
	sp := s.Data
	for i := 0; i < s.Len; i++ {
		setSlice(unsafe.Pointer(sp), *(*uintptr)(unsafe.Pointer(p)), frames)
		sp += unsafe.Sizeof(reflect.SliceHeader{})
		p += unsafe.Sizeof(uintptr(0))
	}
}

func setSlice(s unsafe.Pointer, data uintptr, n int) {
	h := (*reflect.SliceHeader)(s)
	h.Data = data
	h.Len = n
	h.Cap = n
}

func (s *Stream) Start() error {
	err := C.Pa_StartStream(s.paStream)
	if err != C.paNoError {
		return newError(err)
	}
	return nil
}

//export paStreamCallback
func paStreamCallback(inputBuffer, outputBuffer uintptr, frameCount C.ulong, timeInfo *C.PaStreamCallbackTimeInfo, statusFlags C.PaStreamCallbackFlags, userData unsafe.Pointer) C.int {
	(*Stream)(userData).callback(inputBuffer, outputBuffer, int(frameCount))
	return C.paContinue
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

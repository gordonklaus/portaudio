package portaudio

/*
#cgo LDFLAGS: -lportaudio
#include "pa.h"
*/
import "C"

import (
	"code.google.com/p/rog-go/exp/callback"
	"errors"
	."fmt"
	"reflect"
	"unsafe"
)

func init() {
	C.setCallbackFunc(callback.Func)
}

type Int24 [3]byte

type Stream struct {
	paStream unsafe.Pointer
	closed bool
	in, out unsafe.Pointer
	callback func(*C.context)
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
		s.callback = func(c *C.context) {
			s.updateSlices(c)
			p.ProcessAudio(in, out)
		}
		fmt = C.paFloat32 | C.paNonInterleaved
	case interface{ProcessAudio(in, out []float32)}:
		var in, out []float32
		s.in, s.out = unsafe.Pointer(&in), unsafe.Pointer(&out)
		s.callback = func(c *C.context) {
			s.updateSlicesInterleaved(c, numIn, numOut)
			p.ProcessAudio(in, out)
		}
		fmt = C.paFloat32
	case interface{ProcessAudio(in, out [][]int32)}:
		in, out := make([][]int32, numIn), make([][]int32, numOut)
		s.in, s.out = unsafe.Pointer(&in), unsafe.Pointer(&out)
		s.callback = func(c *C.context) {
			s.updateSlices(c)
			p.ProcessAudio(in, out)
		}
		fmt = C.paInt32 | C.paNonInterleaved
	case interface{ProcessAudio(in, out []int32)}:
		var in, out []int32
		s.in, s.out = unsafe.Pointer(&in), unsafe.Pointer(&out)
		s.callback = func(c *C.context) {
			s.updateSlicesInterleaved(c, numIn, numOut)
			p.ProcessAudio(in, out)
		}
		fmt = C.paInt32
	case interface{ProcessAudio(in, out [][]Int24)}:
		in, out := make([][]Int24, numIn), make([][]Int24, numOut)
		s.in, s.out = unsafe.Pointer(&in), unsafe.Pointer(&out)
		s.callback = func(c *C.context) {
			s.updateSlices(c)
			p.ProcessAudio(in, out)
		}
		fmt = C.paInt24 | C.paNonInterleaved
	case interface{ProcessAudio(in, out []Int24)}:
		var in, out []Int24
		s.in, s.out = unsafe.Pointer(&in), unsafe.Pointer(&out)
		s.callback = func(c *C.context) {
			s.updateSlicesInterleaved(c, numIn, numOut)
			p.ProcessAudio(in, out)
		}
		fmt = C.paInt24
	case interface{ProcessAudio(in, out [][]int16)}:
		in, out := make([][]int16, numIn), make([][]int16, numOut)
		s.in, s.out = unsafe.Pointer(&in), unsafe.Pointer(&out)
		s.callback = func(c *C.context) {
			s.updateSlices(c)
			p.ProcessAudio(in, out)
		}
		fmt = C.paInt16 | C.paNonInterleaved
	case interface{ProcessAudio(in, out []int16)}:
		var in, out []int16
		s.in, s.out = unsafe.Pointer(&in), unsafe.Pointer(&out)
		s.callback = func(c *C.context) {
			s.updateSlicesInterleaved(c, numIn, numOut)
			p.ProcessAudio(in, out)
		}
		fmt = C.paInt16
	case interface{ProcessAudio(in, out [][]int8)}:
		in, out := make([][]int8, numIn), make([][]int8, numOut)
		s.in, s.out = unsafe.Pointer(&in), unsafe.Pointer(&out)
		s.callback = func(c *C.context) {
			s.updateSlices(c)
			p.ProcessAudio(in, out)
		}
		fmt = C.paInt8 | C.paNonInterleaved
	case interface{ProcessAudio(in, out []int8)}:
		var in, out []int8
		s.in, s.out = unsafe.Pointer(&in), unsafe.Pointer(&out)
		s.callback = func(c *C.context) {
			s.updateSlicesInterleaved(c, numIn, numOut)
			p.ProcessAudio(in, out)
		}
		fmt = C.paInt8
	case interface{ProcessAudio(in, out [][]uint8)}:
		in, out := make([][]uint8, numIn), make([][]uint8, numOut)
		s.in, s.out = unsafe.Pointer(&in), unsafe.Pointer(&out)
		s.callback = func(c *C.context) {
			s.updateSlices(c)
			p.ProcessAudio(in, out)
		}
		fmt = C.paUInt8 | C.paNonInterleaved
	case interface{ProcessAudio(in, out []uint8)}:
		var in, out []uint8
		s.in, s.out = unsafe.Pointer(&in), unsafe.Pointer(&out)
		s.callback = func(c *C.context) {
			s.updateSlicesInterleaved(c, numIn, numOut)
			p.ProcessAudio(in, out)
		}
		fmt = C.paUInt8
	default:
		err = errors.New(Sprintf("%T lacks a supported ProcessAudio method", audioProcessor))
	}
	return
}

func (s *Stream) updateSlices(c *C.context) {
	setChannels(s.in, c.inputBuffer, c.frameCount)
	setChannels(s.out, c.outputBuffer, c.frameCount)
}

func (s *Stream) updateSlicesInterleaved(c *C.context, in, out int) {
	if in > 0 { setSlice(s.in, c.inputBuffer, int(c.frameCount) * in) }
	if out > 0 { setSlice(s.out, c.outputBuffer, int(c.frameCount) * out) }
}

func setChannels(slice, buffers unsafe.Pointer, frames C.ulong) {
	s := (*reflect.SliceHeader)(slice)
	sp := s.Data
	bp := uintptr(buffers)
	for i := 0; i < s.Len; i++ {
		setSlice(unsafe.Pointer(sp), *(*unsafe.Pointer)(unsafe.Pointer(bp)), int(frames))
		sp += unsafe.Sizeof(reflect.SliceHeader{})
		bp += unsafe.Sizeof(uintptr(0))
	}
}

func setSlice(s, b unsafe.Pointer, n int) {
	h := (*reflect.SliceHeader)(s)
	h.Data = uintptr(b)
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

//export streamCallback
func streamCallback(arg unsafe.Pointer) {
	c := (*C.context)(arg)
	s := (*Stream)(c.stream)
	s.callback(c)
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

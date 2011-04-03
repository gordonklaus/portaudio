package portaudio

//#include <portaudio.h>
//
//typedef struct context context;
//struct context {
//	void *stream;
//	const void *inputBuffer;
//	void *outputBuffer;
//	unsigned long frameCount;
//	int ret;
//};
//
//extern void streamCallback(void*);
//
//// callbackFunc holds the callback library function.
//// It is stored in a function pointer because C linkage
//// does not work across packages.
//static void(*callbackFunc)(void (*f)(void*), void*);
//
//void setCallbackFunc(void *cb){ callbackFunc = cb; }
//
//int paStreamCallback(const void *inputBuffer, void *outputBuffer, unsigned long frameCount
//		, const PaStreamCallbackTimeInfo *timeInfo
//		, PaStreamCallbackFlags statusFlags
//		, void *userData) {
//	context context = { userData, inputBuffer, outputBuffer, frameCount };
//	callbackFunc(streamCallback, &context);
//	return context.ret;
//}
//
//PaStreamCallback* getPaStreamCallback() { return paStreamCallback; }
import "C"

import (
	"rog-go.googlecode.com/hg/exp/callback"
	"unsafe"
	"reflect"
	"log"
)

func init() {
	C.setCallbackFunc(callback.Func)
}

type Error struct {
	Text string
}

type AudioProcessor interface {
	ProcessAudio(inputBuffer, outputBuffer []float32)
}

type Stream struct {
	paStream unsafe.Pointer
	closed bool
	audioProcessor AudioProcessor
}

func OpenDefaultStream(numInputChannels, numOutputChannels int,
						sampleRate float64, framesPerBuffer int,
						audioProcessor AudioProcessor) (*Stream, *Error) {
	error := C.Pa_Initialize()
	if error != C.paNoError {
		return nil, &Error{C.GoString(C.Pa_GetErrorText(error))}
	}
	
	stream := &Stream{}
	error = C.Pa_OpenDefaultStream(&stream.paStream, C.int(numInputChannels), C.int(numOutputChannels), C.paFloat32, C.double(sampleRate), C.ulong(framesPerBuffer), C.getPaStreamCallback(), unsafe.Pointer(stream))
	if error != C.paNoError {
		return nil, &Error{C.GoString(C.Pa_GetErrorText(error))}
	}
	stream.audioProcessor = audioProcessor
	return stream, nil
}

func (s *Stream) Start() *Error {
	error := C.Pa_StartStream(s.paStream)
	if error != C.paNoError {
		return &Error{C.GoString(C.Pa_GetErrorText(error))}
	}
	return nil
}

func sliceAt(buffer unsafe.Pointer, size int) []float32 {
	if buffer == nil { return nil }
	slice := (*[1<<30]float32)(buffer)[:size]
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
	context.ret = C.paContinue
}

func (s *Stream) Stop() *Error {
	error := C.Pa_StopStream(s.paStream)
	if error != C.paNoError {
		return &Error{C.GoString(C.Pa_GetErrorText(error))}
	}
	return nil
}

func (s *Stream) Close() *Error {
	if !s.closed {
		s.closed = true
		error := C.Pa_Terminate()
		if error != C.paNoError {
			return &Error{C.GoString(C.Pa_GetErrorText(error))}
		}
	}
	return nil
}

// not actually called for finalization, yet (but planned to do so in a future Go release)
func (s *Stream) destroy() {
	error := s.Close()
	if error != nil {
		log.Print("Stream.Close() failed in Stream.destroy()")
	}
}

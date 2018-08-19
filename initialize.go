package portaudio

/*
#cgo pkg-config: portaudio-2.0
#include <portaudio.h>
extern PaStreamCallback* paStreamCallback;
*/
import "C"

import (
	"sync/atomic"
)

// A simple counter about how many times port audio was
// initialized
var initialized = new(int32)

// Initialize initializes internal data structures and
// prepares underlying host APIs for use. With the exception
// of Version(), VersionText(), and ErrorText(), this function
// MUST be called before using any other PortAudio API functions.
//
// If Initialize() is called multiple times, each successful call
// must be matched with a corresponding call to Terminate(). Pairs of
// calls to Initialize()/Terminate() may overlap, and are not required to be fully nested.
//
// Note that if Initialize() returns an error code, Terminate() should NOT be called.
func Initialize() error {
	paErr := C.Pa_Initialize()
	if paErr != C.paNoError {
		return newError(paErr)
	}
	atomic.AddInt32(initialized, 1)
	return nil
}

// Terminate deallocates all resources allocated by PortAudio
// since it was initialized by a call to Initialize().
//
// In cases where Initialize() has been called multiple times,
// each call must be matched with a corresponding call to Pa_Terminate().
// The final matching call to Pa_Terminate() will automatically
// close any PortAudio streams that are still open..
//
// Terminate MUST be called before exiting a program which uses PortAudio.
// Failure to do so may result in serious resource leaks, such as audio devices
// not being available until the next reboot.
func Terminate() error {
	paErr := C.Pa_Terminate()
	if paErr != C.paNoError {
		return newError(paErr)
	}

	set := atomic.AddInt32(initialized, -1)
	if set <= 0 {
		atomic.StoreInt32(initialized, 0)
		cached = false
	}
	return nil
}

// Returns whether or not port audio has already been initialized
func isInitialized() bool {
	return atomic.LoadInt32(initialized) > 0
}

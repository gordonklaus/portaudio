#include "pa.h"
#include "_cgo_export.h"

// callbackFunc holds the callback library function.
// It is stored in a function pointer because C linkage
// does not work across packages.
static void(*callbackFunc)(void (*f)(void*), void*);

void setCallbackFunc(void *cb){ callbackFunc = cb; }

int paStreamCallback(const void *inputBuffer, void *outputBuffer, unsigned long frameCount
		, const PaStreamCallbackTimeInfo *timeInfo
		, PaStreamCallbackFlags statusFlags
		, void *userData) {
	context context = { userData, inputBuffer, outputBuffer, frameCount };
	callbackFunc(streamCallback, &context);
	return paContinue;
}

PaStreamCallback* getPaStreamCallback() { return paStreamCallback; }

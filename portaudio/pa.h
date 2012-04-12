#ifndef PA_H
#define PA_H


#include <portaudio.h>

typedef struct context context;
struct context {
	void *stream;
	const void *inputBuffer;
	void *outputBuffer;
	unsigned long frameCount;
};

// from pa.c
extern void setCallbackFunc(void *cb);
extern int paStreamCallback(const void *inputBuffer, void *outputBuffer, unsigned long frameCount
		, const PaStreamCallbackTimeInfo *timeInfo
		, PaStreamCallbackFlags statusFlags
		, void *userData);
extern PaStreamCallback* getPaStreamCallback();


#endif // PA_H

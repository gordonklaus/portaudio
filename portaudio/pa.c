#include "pa.h"
#include "_cgo_export.h"

PaStreamCallback* getPaStreamCallback() {
	return (PaStreamCallback*)paStreamCallback;
}

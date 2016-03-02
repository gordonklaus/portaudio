/*
For the most part, these bindings parallel the underlying PortAudio API; please refer to http://www.portaudio.com/docs.html for details.  Differences introduced by the bindings are documented here:

Instead of passing a flag to OpenStream, audio sample formats are inferred from the signature of the stream callback or, for a blocking stream, from the types of the buffers.  See the StreamCallback and Buffer types for details.

Blocking I/O:  Read and Write do not accept buffer arguments; instead they use the buffers (or pointers to buffers) provided to OpenStream.  The number of samples to read or write is determined by the size of the buffers.

The StreamParameters struct combines parameters for both the input and the output device as well as the sample rate, buffer size, and flags.
*/
package portaudio

/*
#cgo pkg-config: portaudio-2.0
#include <portaudio.h>
extern PaStreamCallback* paStreamCallback;
*/
import "C"

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"sync"
	"time"
	"unsafe"
)

func Version() int {
	return int(C.Pa_GetVersion())
}

func VersionText() string {
	return C.GoString(C.Pa_GetVersionText())
}

type Error C.PaError

func (err Error) Error() string {
	return C.GoString(C.Pa_GetErrorText(C.PaError(err)))
}

const (
	NotInitialized                        Error = C.paNotInitialized
	InvalidChannelCount                   Error = C.paInvalidChannelCount
	InvalidSampleRate                     Error = C.paInvalidSampleRate
	InvalidDevice                         Error = C.paInvalidDevice
	InvalidFlag                           Error = C.paInvalidFlag
	SampleFormatNotSupported              Error = C.paSampleFormatNotSupported
	BadIODeviceCombination                Error = C.paBadIODeviceCombination
	InsufficientMemory                    Error = C.paInsufficientMemory
	BufferTooBig                          Error = C.paBufferTooBig
	BufferTooSmall                        Error = C.paBufferTooSmall
	NullCallback                          Error = C.paNullCallback
	BadStreamPtr                          Error = C.paBadStreamPtr
	TimedOut                              Error = C.paTimedOut
	InternalError                         Error = C.paInternalError
	DeviceUnavailable                     Error = C.paDeviceUnavailable
	IncompatibleHostApiSpecificStreamInfo Error = C.paIncompatibleHostApiSpecificStreamInfo
	StreamIsStopped                       Error = C.paStreamIsStopped
	StreamIsNotStopped                    Error = C.paStreamIsNotStopped
	InputOverflowed                       Error = C.paInputOverflowed
	OutputUnderflowed                     Error = C.paOutputUnderflowed
	HostApiNotFound                       Error = C.paHostApiNotFound
	InvalidHostApi                        Error = C.paInvalidHostApi
	CanNotReadFromACallbackStream         Error = C.paCanNotReadFromACallbackStream
	CanNotWriteToACallbackStream          Error = C.paCanNotWriteToACallbackStream
	CanNotReadFromAnOutputOnlyStream      Error = C.paCanNotReadFromAnOutputOnlyStream
	CanNotWriteToAnInputOnlyStream        Error = C.paCanNotWriteToAnInputOnlyStream
	IncompatibleStreamHostApi             Error = C.paIncompatibleStreamHostApi
	BadBufferPtr                          Error = C.paBadBufferPtr
)

type UnanticipatedHostError struct {
	HostApiType HostApiType
	Code        int
	Text        string
}

func (err UnanticipatedHostError) Error() string {
	return err.Text
}

func newError(err C.PaError) error {
	switch err {
	case C.paUnanticipatedHostError:
		hostErr := C.Pa_GetLastHostErrorInfo()
		return UnanticipatedHostError{
			HostApiType(hostErr.hostApiType),
			int(hostErr.errorCode),
			C.GoString(hostErr.errorText),
		}
	case C.paNoError:
		return nil
	}
	return Error(err)
}

var initialized = 0

func Initialize() error {
	paErr := C.Pa_Initialize()
	if paErr != C.paNoError {
		return newError(paErr)
	}
	initialized++
	return nil
}

func Terminate() error {
	paErr := C.Pa_Terminate()
	if paErr != C.paNoError {
		return newError(paErr)
	}
	initialized--
	if initialized <= 0 {
		initialized = 0
		cached = false
	}
	return nil
}

type HostApiType int

func (t HostApiType) String() string {
	return hostApiStrings[t]
}

var hostApiStrings = [...]string{
	InDevelopment:   "InDevelopment",
	DirectSound:     "DirectSound",
	MME:             "MME",
	ASIO:            "ASIO",
	SoundManager:    "SoundManager",
	CoreAudio:       "CoreAudio",
	OSS:             "OSS",
	ALSA:            "ALSA",
	AL:              "AL",
	BeOS:            "BeOS",
	WDMkS:           "WDMKS",
	JACK:            "JACK",
	WASAPI:          "WASAPI",
	AudioScienceHPI: "AudioScienceHPI",
}

const (
	InDevelopment   HostApiType = C.paInDevelopment
	DirectSound     HostApiType = C.paDirectSound
	MME             HostApiType = C.paMME
	ASIO            HostApiType = C.paASIO
	SoundManager    HostApiType = C.paSoundManager
	CoreAudio       HostApiType = C.paCoreAudio
	OSS             HostApiType = C.paOSS
	ALSA            HostApiType = C.paALSA
	AL              HostApiType = C.paAL
	BeOS            HostApiType = C.paBeOS
	WDMkS           HostApiType = C.paWDMKS
	JACK            HostApiType = C.paJACK
	WASAPI          HostApiType = C.paWASAPI
	AudioScienceHPI HostApiType = C.paAudioScienceHPI
)

type HostApiInfo struct {
	Type                HostApiType
	Name                string
	DefaultInputDevice  *DeviceInfo
	DefaultOutputDevice *DeviceInfo
	Devices             []*DeviceInfo
}

type DeviceInfo struct {
	index                    C.PaDeviceIndex
	Name                     string
	MaxInputChannels         int
	MaxOutputChannels        int
	DefaultLowInputLatency   time.Duration
	DefaultLowOutputLatency  time.Duration
	DefaultHighInputLatency  time.Duration
	DefaultHighOutputLatency time.Duration
	DefaultSampleRate        float64
	HostApi                  *HostApiInfo
}

func HostApis() ([]*HostApiInfo, error) {
	hosts, _, err := hostsAndDevices()
	if err != nil {
		return nil, err
	}
	return hosts, nil
}

func HostApi(apiType HostApiType) (*HostApiInfo, error) {
	hosts, err := HostApis()
	if err != nil {
		return nil, err
	}
	i := C.Pa_HostApiTypeIdToHostApiIndex(C.PaHostApiTypeId(apiType))
	if i < 0 {
		return nil, newError(C.PaError(i))
	}
	return hosts[i], nil
}

func DefaultHostApi() (*HostApiInfo, error) {
	hosts, err := HostApis()
	if err != nil {
		return nil, err
	}
	i := C.Pa_GetDefaultHostApi()
	if i < 0 {
		return nil, newError(C.PaError(i))
	}
	return hosts[i], nil
}

func Devices() ([]*DeviceInfo, error) {
	_, devs, err := hostsAndDevices()
	if err != nil {
		return nil, err
	}
	return devs, nil
}

func DefaultInputDevice() (*DeviceInfo, error) {
	devs, err := Devices()
	if err != nil {
		return nil, err
	}
	i := C.Pa_GetDefaultInputDevice()
	if i < 0 {
		return nil, newError(C.PaError(i))
	}
	return devs[i], nil
}

func DefaultOutputDevice() (*DeviceInfo, error) {
	devs, err := Devices()
	if err != nil {
		return nil, err
	}
	i := C.Pa_GetDefaultOutputDevice()
	if i < 0 {
		return nil, newError(C.PaError(i))
	}
	return devs[i], nil
}

/* cache the HostApi/Device list to simplify the enumeration code.
Note that portaudio itself caches the lists, so these won't go stale.

However, there is talk of extending the portaudio API to allow clients
to rescan available devices without calling Pa_Terminate() followed by
Pa_Initialize() - our caching strategy will have to change if this
goes ahead. See https://www.assembla.com/spaces/portaudio/tickets/11
*/
var (
	cached   bool
	hostApis []*HostApiInfo
	devices  []*DeviceInfo
)

func hostsAndDevices() ([]*HostApiInfo, []*DeviceInfo, error) {
	if !cached {
		nhosts := C.Pa_GetHostApiCount()
		ndevs := C.Pa_GetDeviceCount()
		if nhosts < 0 {
			return nil, nil, newError(C.PaError(nhosts))
		}
		if ndevs < 0 {
			return nil, nil, newError(C.PaError(ndevs))
		}
		devices = make([]*DeviceInfo, ndevs)
		hosti := make([]C.PaHostApiIndex, ndevs)
		for i := range devices {
			i := C.PaDeviceIndex(i)
			paDev := C.Pa_GetDeviceInfo(i)
			devices[i] = &DeviceInfo{
				index:                    i,
				Name:                     C.GoString(paDev.name),
				MaxInputChannels:         int(paDev.maxInputChannels),
				MaxOutputChannels:        int(paDev.maxOutputChannels),
				DefaultLowInputLatency:   duration(paDev.defaultLowInputLatency),
				DefaultLowOutputLatency:  duration(paDev.defaultLowOutputLatency),
				DefaultHighInputLatency:  duration(paDev.defaultHighInputLatency),
				DefaultHighOutputLatency: duration(paDev.defaultHighOutputLatency),
				DefaultSampleRate:        float64(paDev.defaultSampleRate),
			}
			hosti[i] = paDev.hostApi
		}
		hostApis = make([]*HostApiInfo, nhosts)
		for i := range hostApis {
			i := C.PaHostApiIndex(i)
			paHost := C.Pa_GetHostApiInfo(i)
			devs := make([]*DeviceInfo, paHost.deviceCount)
			for j := range devs {
				devs[j] = devices[C.Pa_HostApiDeviceIndexToDeviceIndex(i, C.int(j))]
			}
			hostApis[i] = &HostApiInfo{
				Type:                HostApiType(paHost._type),
				Name:                C.GoString(paHost.name),
				DefaultInputDevice:  lookupDevice(devices, paHost.defaultInputDevice),
				DefaultOutputDevice: lookupDevice(devices, paHost.defaultOutputDevice),
				Devices:             devs,
			}
		}
		for i := range devices {
			devices[i].HostApi = hostApis[hosti[i]]
		}
		cached = true
	}
	return hostApis, devices, nil
}

func duration(paTime C.PaTime) time.Duration {
	return time.Duration(paTime * C.PaTime(time.Second))
}

func lookupDevice(d []*DeviceInfo, i C.PaDeviceIndex) *DeviceInfo {
	if i >= 0 {
		return d[i]
	}
	return nil
}

/*
StreamParameters includes all parameters required to open a stream except for the callback or buffers.
*/
type StreamParameters struct {
	Input, Output   StreamDeviceParameters
	SampleRate      float64
	FramesPerBuffer int
	Flags           StreamFlags
}

/*
StreamDeviceParameters specifies parameters for one device (either input or output) in a stream.  A nil Device indicates that no device is to be used -- i.e., for an input- or output-only stream.
*/
type StreamDeviceParameters struct {
	Device   *DeviceInfo
	Channels int
	Latency  time.Duration
}

const FramesPerBufferUnspecified = C.paFramesPerBufferUnspecified

type StreamFlags C.PaStreamFlags

const (
	NoFlag                                StreamFlags = C.paNoFlag
	ClipOff                               StreamFlags = C.paClipOff
	DitherOff                             StreamFlags = C.paDitherOff
	NeverDropInput                        StreamFlags = C.paNeverDropInput
	PrimeOutputBuffersUsingStreamCallback StreamFlags = C.paPrimeOutputBuffersUsingStreamCallback
	PlatformSpecificFlags                 StreamFlags = C.paPlatformSpecificFlags
)

/*
High latency parameters are mono in, stereo out (if supported), high latency, the smaller of the default sample rates of the two devices, andFramesPerBufferUnspecified.  One of the devices may be nil.
*/
func HighLatencyParameters(in, out *DeviceInfo) (p StreamParameters) {
	sampleRate := 0.0
	if in != nil {
		p := &p.Input
		p.Device = in
		p.Channels = 1
		if in.MaxInputChannels < 1 {
			p.Channels = in.MaxInputChannels
		}
		p.Latency = in.DefaultHighInputLatency
		sampleRate = in.DefaultSampleRate
	}
	if out != nil {
		p := &p.Output
		p.Device = out
		p.Channels = 2
		if out.MaxOutputChannels < 2 {
			p.Channels = out.MaxOutputChannels
		}
		p.Latency = out.DefaultHighOutputLatency
		if r := out.DefaultSampleRate; r < sampleRate || sampleRate == 0 {
			sampleRate = r
		}
	}
	p.SampleRate = sampleRate
	p.FramesPerBuffer = FramesPerBufferUnspecified
	return p
}

/*
Low latency parameters are mono in, stereo out (if supported), low latency, the larger of the default sample rates of the two devices, and FramesPerBufferUnspecified.  One of the devices may be nil.
*/
func LowLatencyParameters(in, out *DeviceInfo) (p StreamParameters) {
	sampleRate := 0.0
	if in != nil {
		p := &p.Input
		p.Device = in
		p.Channels = 1
		if in.MaxInputChannels < 1 {
			p.Channels = in.MaxInputChannels
		}
		p.Latency = in.DefaultLowInputLatency
		sampleRate = in.DefaultSampleRate
	}
	if out != nil {
		p := &p.Output
		p.Device = out
		p.Channels = 2
		if out.MaxOutputChannels < 2 {
			p.Channels = out.MaxOutputChannels
		}
		p.Latency = out.DefaultLowOutputLatency
		if r := out.DefaultSampleRate; r > sampleRate {
			sampleRate = r
		}
	}
	p.SampleRate = sampleRate
	p.FramesPerBuffer = FramesPerBufferUnspecified
	return p
}

/*
Returns nil if the format is supported, otherwise an error.
The args parameter has the same meaning as in OpenStream.
*/
func IsFormatSupported(p StreamParameters, args ...interface{}) error {
	s := &Stream{}
	err := s.init(p, args...)
	if err != nil {
		return err
	}
	return newError(C.Pa_IsFormatSupported(s.inParams, s.outParams, C.double(p.SampleRate)))
}

type Int24 [3]byte

type Stream struct {
	id                  uintptr
	paStream            unsafe.Pointer
	inParams, outParams *C.PaStreamParameters
	in, out             *reflect.SliceHeader
	timeInfo            StreamCallbackTimeInfo
	flags               StreamCallbackFlags
	args                []reflect.Value
	callback            reflect.Value
	closed              bool
}

/*
This type exists for documentation purposes only.

A StreamCallback is a func whose signature resembles

	func(in Buffer, out Buffer, timeInfo StreamCallbackTimeInfo, flags StreamCallbackFlags)

where the final one or two parameters may be omitted.  For an input- or output-only stream, one of the Buffer parameters may also be omitted.  The two Buffer types may be different.
*/
type StreamCallback interface{}

/*
This type exists for documentation purposes only.

A Buffer is of the form [][]SampleType or []SampleType
where SampleType is float32, int32, Int24, int16, int8, or uint8.

In the first form, channels are non-interleaved:
len(buf) == numChannels and len(buf[i]) == framesPerBuffer

In the second form, channels are interleaved:
len(buf) == numChannels * framesPerBuffer
*/
type Buffer interface{}

type StreamCallbackTimeInfo struct {
	InputBufferAdcTime, CurrentTime, OutputBufferDacTime time.Duration
}

type StreamCallbackFlags C.PaStreamCallbackFlags

const (
	InputUnderflow  StreamCallbackFlags = C.paInputUnderflow
	InputOverflow   StreamCallbackFlags = C.paInputOverflow
	OutputUnderflow StreamCallbackFlags = C.paOutputUnderflow
	OutputOverflow  StreamCallbackFlags = C.paOutputOverflow
	PrimingOutput   StreamCallbackFlags = C.paPrimingOutput
)

/*
For an input- or output-only stream, p.Output.Device or p.Input.Device must be nil, respectively.

The args may consist of either a single StreamCallback or, for a blocking stream, two Buffers or pointers to Buffers.  For an input- or output-only stream, one of the Buffer args may be omitted.
*/
func OpenStream(p StreamParameters, args ...interface{}) (*Stream, error) {
	if initialized <= 0 {
		return nil, NotInitialized
	}

	s := &Stream{}
	err := s.init(p, args...)
	if err != nil {
		return nil, err
	}

	cb := C.paStreamCallback
	if !s.callback.IsValid() {
		cb = nil
	}

	id := scm.Track(s)
	paErr := C.Pa_OpenStream(&s.paStream, s.inParams, s.outParams, C.double(p.SampleRate), C.ulong(p.FramesPerBuffer), C.PaStreamFlags(p.Flags), cb, unsafe.Pointer(id))
	if paErr != C.paNoError {
		return nil, newError(paErr)
	}
	return s, nil
}

/*
The args parameter has the same meaning as in OpenStream.
*/
func OpenDefaultStream(numInputChannels, numOutputChannels int, sampleRate float64, framesPerBuffer int, args ...interface{}) (*Stream, error) {
	if initialized <= 0 {
		return nil, NotInitialized
	}

	var inDev, outDev *DeviceInfo
	var err error
	if numInputChannels > 0 {
		inDev, err = DefaultInputDevice()
		if err != nil {
			return nil, err
		}
	}
	if numOutputChannels > 0 {
		outDev, err = DefaultOutputDevice()
		if err != nil {
			return nil, err
		}
	}
	p := HighLatencyParameters(inDev, outDev)
	p.Input.Channels = numInputChannels
	p.Output.Channels = numOutputChannels
	p.SampleRate = sampleRate
	p.FramesPerBuffer = framesPerBuffer
	return OpenStream(p, args...)
}

func (s *Stream) init(p StreamParameters, args ...interface{}) error {
	switch len(args) {
	case 0:
		return fmt.Errorf("too few args")
	case 1, 2:
		if fun := reflect.ValueOf(args[0]); fun.Kind() == reflect.Func {
			return s.initCallback(p, fun)
		}
		return s.initBuffers(p, args...)
	default:
		return fmt.Errorf("too many args")
	}
}

func (s *Stream) initCallback(p StreamParameters, fun reflect.Value) error {
	t := fun.Type()
	if t.IsVariadic() {
		return fmt.Errorf("StreamCallback must not be variadic")
	}
	nArgs := t.NumIn()
	if nArgs == 0 {
		return fmt.Errorf("too few parameters in StreamCallback")
	}
	args := make([]reflect.Value, nArgs)
	i := 0
	bothBufs := nArgs > 1 && t.In(1).Kind() == reflect.Slice
	bufArg := func(p StreamDeviceParameters) (*C.PaStreamParameters, *reflect.SliceHeader, error) {
		if p.Device != nil || bothBufs {
			if i >= nArgs {
				return nil, nil, fmt.Errorf("too few Buffer parameters in StreamCallback")
			}
			t := t.In(i)
			sampleFmt := sampleFormat(t)
			if sampleFmt == 0 {
				return nil, nil, fmt.Errorf("expected Buffer type in StreamCallback, got %v", t)
			}
			buf := reflect.New(t)
			args[i] = buf.Elem()
			i++
			if p.Device != nil {
				pap := paStreamParameters(p, sampleFmt)
				if pap.sampleFormat&C.paNonInterleaved != 0 {
					n := int(pap.channelCount)
					buf.Elem().Set(reflect.MakeSlice(t, n, n))
				}
				return pap, (*reflect.SliceHeader)(unsafe.Pointer(buf.Pointer())), nil
			}
		}
		return nil, nil, nil
	}
	var err error
	s.inParams, s.in, err = bufArg(p.Input)
	if err != nil {
		return err
	}
	s.outParams, s.out, err = bufArg(p.Output)
	if err != nil {
		return err
	}
	if i < nArgs {
		t := t.In(i)
		if t != reflect.TypeOf(StreamCallbackTimeInfo{}) {
			return fmt.Errorf("invalid StreamCallback")
		}
		args[i] = reflect.ValueOf(&s.timeInfo).Elem()
		i++
	}
	if i < nArgs {
		t := t.In(i)
		if t != reflect.TypeOf(StreamCallbackFlags(0)) {
			return fmt.Errorf("invalid StreamCallback")
		}
		args[i] = reflect.ValueOf(&s.flags).Elem()
		i++
	}
	if i < nArgs {
		return fmt.Errorf("too many parameters in StreamCallback")
	}
	if t.NumOut() > 0 {
		return fmt.Errorf("too many results in StreamCallback")
	}
	s.callback = fun
	s.args = args
	return nil
}

func (s *Stream) initBuffers(p StreamParameters, args ...interface{}) error {
	bothBufs := len(args) == 2
	bufArg := func(p StreamDeviceParameters) (*C.PaStreamParameters, *reflect.SliceHeader, error) {
		if p.Device != nil || bothBufs {
			if len(args) == 0 {
				return nil, nil, fmt.Errorf("too few Buffer args")
			}
			arg := reflect.ValueOf(args[0])
			args = args[1:]
			t := arg.Type()
			if t.Kind() == reflect.Ptr {
				t = t.Elem()
			} else {
				argPtr := reflect.New(t)
				argPtr.Elem().Set(arg)
				arg = argPtr
			}
			sampleFmt := sampleFormat(t)
			if sampleFmt == 0 {
				return nil, nil, fmt.Errorf("invalid Buffer type %v", t)
			}
			if arg.IsNil() {
				return nil, nil, fmt.Errorf("nil Buffer pointer")
			}
			if p.Device != nil {
				return paStreamParameters(p, sampleFmt), (*reflect.SliceHeader)(unsafe.Pointer(arg.Pointer())), nil
			}
		}
		return nil, nil, nil
	}
	var err error
	s.inParams, s.in, err = bufArg(p.Input)
	if err != nil {
		return err
	}
	s.outParams, s.out, err = bufArg(p.Output)
	if err != nil {
		return err
	}
	return nil
}

func sampleFormat(b reflect.Type) (f C.PaSampleFormat) {
	if b.Kind() != reflect.Slice {
		return 0
	}
	b = b.Elem()
	if b.Kind() == reflect.Slice {
		f = C.paNonInterleaved
		b = b.Elem()
	}
	switch b.Kind() {
	case reflect.Float32:
		f |= C.paFloat32
	case reflect.Int32:
		f |= C.paInt32
	default:
		if b == reflect.TypeOf(Int24{}) {
			f |= C.paInt24
		} else {
			return 0
		}
	case reflect.Int16:
		f |= C.paInt16
	case reflect.Int8:
		f |= C.paInt8
	case reflect.Uint8:
		f |= C.paUInt8
	}
	return f
}

func paStreamParameters(p StreamDeviceParameters, fmt C.PaSampleFormat) *C.PaStreamParameters {
	return &C.PaStreamParameters{
		device:           p.Device.index,
		channelCount:     C.int(p.Channels),
		sampleFormat:     fmt,
		suggestedLatency: C.PaTime(p.Latency.Seconds()),
	}
}

func (s *Stream) Close() error {
	scm.Untrack(s.id)
	if !s.closed {
		s.closed = true
		return newError(C.Pa_CloseStream(s.paStream))
	}
	return nil
}

func (s *Stream) Start() error {
	return newError(C.Pa_StartStream(s.paStream))
}

//export streamCallback
func streamCallback(inputBuffer, outputBuffer unsafe.Pointer, frames C.ulong, timeInfo *C.PaStreamCallbackTimeInfo, statusFlags C.PaStreamCallbackFlags, userData unsafe.Pointer) {
	defer func() {
		// Don't let PortAudio silently swallow panics.
		if x := recover(); x != nil {
			buf := make([]byte, 1<<10)
			for runtime.Stack(buf, true) == len(buf) {
				buf = make([]byte, 2*len(buf))
			}
			fmt.Fprintf(os.Stderr, "panic in portaudio stream callback: %s\n\n%s", x, buf)
			os.Exit(2)
		}
	}()

	s := scm.Get(uintptr(userData))
	s.timeInfo = StreamCallbackTimeInfo{duration(timeInfo.inputBufferAdcTime), duration(timeInfo.currentTime), duration(timeInfo.outputBufferDacTime)}
	s.flags = StreamCallbackFlags(statusFlags)
	updateBuffer(s.in, uintptr(inputBuffer), s.inParams, int(frames))
	updateBuffer(s.out, uintptr(outputBuffer), s.outParams, int(frames))
	s.callback.Call(s.args)
}

func updateBuffer(buf *reflect.SliceHeader, p uintptr, params *C.PaStreamParameters, frames int) {
	if p == 0 {
		return
	}
	if params.sampleFormat&C.paNonInterleaved == 0 {
		setSlice(buf, p, frames*int(params.channelCount))
	} else {
		setChannels(buf, p, frames)
	}
}

func setChannels(s *reflect.SliceHeader, p uintptr, frames int) {
	sp := s.Data
	for i := 0; i < s.Len; i++ {
		setSlice((*reflect.SliceHeader)(unsafe.Pointer(sp)), *(*uintptr)(unsafe.Pointer(p)), frames)
		sp += unsafe.Sizeof(reflect.SliceHeader{})
		p += unsafe.Sizeof(uintptr(0))
	}
}

func setSlice(s *reflect.SliceHeader, data uintptr, n int) {
	s.Data = data
	s.Len = n
	s.Cap = n
}

func (s *Stream) Stop() error {
	return newError(C.Pa_StopStream(s.paStream))
}

func (s *Stream) Abort() error {
	return newError(C.Pa_AbortStream(s.paStream))
}

func (s *Stream) Info() *StreamInfo {
	i := C.Pa_GetStreamInfo(s.paStream)
	if i == nil {
		return nil
	}
	return &StreamInfo{duration(i.inputLatency), duration(i.outputLatency), float64(i.sampleRate)}
}

type StreamInfo struct {
	InputLatency, OutputLatency time.Duration
	SampleRate                  float64
}

func (s *Stream) Time() time.Duration {
	return duration(C.Pa_GetStreamTime(s.paStream))
}

func (s *Stream) CpuLoad() float64 {
	return float64(C.Pa_GetStreamCpuLoad(s.paStream))
}

func (s *Stream) AvailableToRead() (int, error) {
	n := C.Pa_GetStreamReadAvailable(s.paStream)
	if n < 0 {
		return 0, newError(C.PaError(n))
	}
	return int(n), nil
}

func (s *Stream) AvailableToWrite() (int, error) {
	n := C.Pa_GetStreamWriteAvailable(s.paStream)
	if n < 0 {
		return 0, newError(C.PaError(n))
	}
	return int(n), nil
}

/*
Read uses the buffer provided to OpenStream.  The number of samples to read is determined by the size of the buffer.
*/
func (s *Stream) Read() error {
	if s.callback.IsValid() {
		return CanNotReadFromACallbackStream
	}
	if s.in == nil {
		return CanNotReadFromAnOutputOnlyStream
	}
	buf, frames, err := getBuffer(s.in, s.inParams)
	if err != nil {
		return err
	}
	return newError(C.Pa_ReadStream(s.paStream, buf, C.ulong(frames)))
}

/*
Write uses the buffer provided to OpenStream.  The number of samples to write is determined by the size of the buffer.
*/
func (s *Stream) Write() error {
	if s.callback.IsValid() {
		return CanNotWriteToACallbackStream
	}
	if s.out == nil {
		return CanNotWriteToAnInputOnlyStream
	}
	buf, frames, err := getBuffer(s.out, s.outParams)
	if err != nil {
		return err
	}
	return newError(C.Pa_WriteStream(s.paStream, buf, C.ulong(frames)))
}

func getBuffer(s *reflect.SliceHeader, p *C.PaStreamParameters) (unsafe.Pointer, int, error) {
	if p.sampleFormat&C.paNonInterleaved == 0 {
		n := int(p.channelCount)
		if s.Len%n != 0 {
			return nil, 0, fmt.Errorf("length of interleaved buffer not divisible by number of channels")
		}
		return unsafe.Pointer(s.Data), s.Len / n, nil
	} else {
		if s.Len != int(p.channelCount) {
			return nil, 0, fmt.Errorf("buffer has wrong number of channels")
		}
		buf := make([]uintptr, s.Len)
		frames := -1
		sp := s.Data
		for i := range buf {
			ch := (*reflect.SliceHeader)(unsafe.Pointer(sp))
			if frames == -1 {
				frames = ch.Len
			} else if ch.Len != frames {
				return nil, 0, fmt.Errorf("channels have different lengths")
			}
			buf[i] = ch.Data
			sp += unsafe.Sizeof(reflect.SliceHeader{})
		}
		return unsafe.Pointer(&buf[0]), frames, nil
	}
}

// StreamCMap tracks the pointers of the Streams between Go and CGO and was required as of the release of Go 1.6
// "panic: runtime error: cgo argument has Go pointer to Go pointer"
type StreamCMap struct {
	sync.RWMutex
	streams map[uintptr]*Stream
	nextId  uintptr
}

var scm = &StreamCMap{streams: make(map[uintptr]*Stream), nextId: 0}

func (scm *StreamCMap) Get(id uintptr) *Stream {
	scm.RLock()
	defer scm.RUnlock()
	s := scm.streams[id]
	if s == nil {
		panic("unregistered stream")
	}
	return s
}

func (scm *StreamCMap) Track(s *Stream) uintptr {
	scm.Lock()
	defer scm.Unlock()
	s.id = scm.nextId
	scm.nextId++
	scm.streams[s.id] = s
	return s.id
}

func (scm *StreamCMap) Untrack(id uintptr) {
	scm.Lock()
	defer scm.Unlock()
	delete(scm.streams, id)
}

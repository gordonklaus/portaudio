# Implementation #

In the spirit of Go, this package aims to provide a friendly and type safe interface to the PortAudio library.  The underlying C library, however, exhibits some lack of type safety.  In particular, it passes audio data around via pointers-to-void that the user has to correctly interpret as buffers of the appropriate format.

The buffer format appears twice in a C program: once to open an audio stream and again to interpret the buffers as they are received and sent.  This duplication is mildly inconvenient; and even potentially, although easily avoidably, problematic -- misinterpreted buffers lead to garbled data and memory corruption.  It's not a bad interface, but we can do better.

When using the Go package, the buffer format is specified in exactly one place: the callback function signature (or, for blocking streams, in the buffer types).  The user is thus relieved of interpreting buffers; she need only declare their type at the point of their use.  The callback function can be a regular function value or a method value, the latter making it easy to attach contextual data to the stream (which, again, the C library achieves using pointer-to-void).

In `Stream.init`, portaudio-go uses reflection to determine whether a `StreamCallback` or `Buffer` arguments were supplied, calling `Stream.initCallback` or `Stream.initBuffers` as appropriate.

`Stream.initCallback` identifies the sample format by reflecting on the type of the callback function.  The buffers are created as slices of the particular format and stored generically as `unsafe.Pointer`s in `Stream.in` and `Stream.out`.  A slice of `reflect.Value`s stores as many arguments as the callback needs, allowing one to omit the timeInfo and flags parameters.

When the underlying callback function `streamCallback` is called, the underlying data of the buffer slices is set generically (in `updateBuffer`, `setChannels`, and `setSlice`) by manipulating them through `reflect.SliceHeader`; their `Data` fields are simply set to point to the raw data provided by PortAudio, and lengths and capacities adjusted accordingly.  In `setChannels()`, `uintptr` arithmetic is used to generically iterate through non-interleaved buffers.  Finally, the `StreamCallback` is called via reflection.

For blocking I/O, `Stream.initBuffers` identifies the sample format by reflecting on the types of the buffer arguments.  If they are pointers to buffers, those are stored, otherwise pointers are created.  When `Stream.Read` or `Stream.Write` is called, the relevant pointer is followed, the buffer size is checked, and the data is read or written.
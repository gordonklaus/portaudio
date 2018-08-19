/*
Package portaudio applies Go bindings to the PortAudio library.

For the most part, these bindings parallel the underlying PortAudio API; please refer to http://www.portaudio.com/docs.html for details.  Differences introduced by the bindings are documented here:

Instead of passing a flag to OpenStream, audio sample formats are inferred from the signature of the stream callback or, for a blocking stream, from the types of the buffers.  See the StreamCallback and Buffer types for details.

Blocking I/O:  Read and Write do not accept buffer arguments; instead they use the buffers (or pointers to buffers) provided to OpenStream.  The number of samples to read or write is determined by the size of the buffers.

The StreamParameters struct combines parameters for both the input and the output device as well as the sample rate, buffer size, and flags.
*/
package portaudio

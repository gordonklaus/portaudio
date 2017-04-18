# portaudio

This package provides an interface to the [PortAudio](http://www.portaudio.com/) audio I/O library.  See the [package documentation](http://godoc.org/github.com/gordonklaus/portaudio) for details.

To build this package you must first have the PortAudio development headers and libraries installed.  Some systems provide a package for this; e.g., on Ubuntu you would want to run `apt-get install portaudio19-dev`.  On other systems you might have to install from source.

Thanks to sqweek for motivating and contributing to host API and device enumeration.

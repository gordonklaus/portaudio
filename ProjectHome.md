# portaudio #
This package provides an interface to the [PortAudio](http://www.portaudio.com/) audio I/O library.

To build portaudio-go, you must first have the PortAudio development headers and libraries installed.  Some systems provide a package for this; e.g., on Ubuntu you would want to run `apt-get install portaudio19-dev`.  On other systems you might have to install from source.  Then, you can
```
go get code.google.com/p/portaudio-go/portaudio
```

[examples](http://code.google.com/p/portaudio-go/source/browse/portaudio/examples)

A previous version of OpenDefaultStream automatically called Initialize (and Stream.Close called Terminate).  This behavior no longer exists; clients must explicitly call Initialize before making any other calls (and finally Terminate).

Thanks to sqweek for motivating and contributing to host API and device enumeration.
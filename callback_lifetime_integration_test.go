//go:build portaudio_integration

package portaudio

import (
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

// Run with go test -tags portaudio_integration -race -run TestStreamUserDataLifetime.
// Requires a usable default output device; lifecycle calls are intentionally serial.
func TestStreamUserDataLifetime(t *testing.T) {
	if err := Initialize(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := Terminate(); err != nil {
			t.Error(err)
		}
	}()
	out, err := DefaultOutputDevice()
	if err != nil {
		t.Skipf("no default output device: %v", err)
	}
	if out == nil || out.MaxOutputChannels < 1 {
		t.Skip("no usable default output device")
	}
	p := HighLatencyParameters(nil, out)
	p.Output.Channels = 1
	p.FramesPerBuffer = 256
	if err := IsFormatSupported(p, func([]float32) {}); err != nil {
		t.Skipf("default output format unavailable: %v", err)
	}
	t.Logf("PortAudio=%q device index=%d name=%q host=%s", VersionText(), out.Index, out.Name, out.HostApi.Name)

	// ID zero can hide the invalid integer-to-Go-pointer conversion.
	dummy := newStream()
	delStream(dummy)
	before := len(streams)
	// Fail before C and inside Pa_OpenStream; both must remove the registered ID.
	if s, err := OpenStream(p); err == nil || s != nil {
		t.Fatal("expected init failure")
	}
	if len(streams) != before {
		t.Fatal("init failure leaked registry entry")
	}
	bad := p
	bad.SampleRate = -1
	if s, err := OpenStream(bad, func([]float32) {}); err == nil || s != nil {
		t.Fatal("expected native open failure")
	}
	if len(streams) != before {
		t.Fatal("native open failure leaked registry entry")
	}
	for cycle := 0; cycle < 5; cycle++ {
		var calls int64
		s, err := OpenStream(p, func(b []float32) {
			for i := range b {
				b[i] = 0
			}
			atomic.AddInt64(&calls, 1)
		})
		if err != nil {
			t.Fatal(err)
		}
		id := s.id
		closed := false
		defer func() {
			if !closed {
				if live := getStream(id); live != nil {
					if err := live.Close(); err != nil {
						t.Error(err)
					}
				}
			}
		}()
		if id == 0 || getStream(id) != s {
			t.Fatal("nonzero ID not registered")
		}
		if err = s.Start(); err != nil {
			t.Fatal(err)
		}
		s = nil // Only the registry must root the stream during GC.
		for i := 0; i < 8; i++ {
			runtime.GC()
			_ = make([]byte, 1<<18)
			time.Sleep(12 * time.Millisecond)
		}
		s = getStream(id)
		if s == nil {
			t.Fatal("GC lost registry root")
		}
		if n := atomic.LoadInt64(&calls); n < 2 {
			t.Fatalf("missing real callback: %d", n)
		}
		// Exercise synchronous close both after Stop and while active.
		if cycle%2 == 0 {
			if err = s.Stop(); err != nil {
				t.Fatal(err)
			}
		}
		if err = s.Close(); err != nil {
			t.Fatal(err)
		}
		closed = true
		if err = s.Close(); err != nil {
			t.Fatal(err)
		}
		n := atomic.LoadInt64(&calls)
		runtime.GC()
		time.Sleep(30 * time.Millisecond)
		if atomic.LoadInt64(&calls) != n {
			t.Fatal("callback after close")
		}
		if getStream(id) != nil || len(streams) != before {
			t.Fatal("close leaked registry entry")
		}
		t.Logf("cycle=%d opaque_id=%d callbacks=%d registry_after=%d", cycle, id, n, len(streams))
	}
}

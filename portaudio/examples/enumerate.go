package main

import (
	"code.google.com/p/portaudio-go/portaudio"
	"os"
	"text/template"
)

const tmpl = `{{. | len}} host APIs: {{range .}}
	Name:                   {{.Name}}
	Default input device:   {{.DefaultInputDevice.Name}}
	Default output device:  {{.DefaultOutputDevice.Name}}
	Devices: {{range .Devices}}
		Name:                      {{.Name}}
		MaxInputChannels:          {{.MaxInputChannels}}
		MaxOutputChannels:         {{.MaxOutputChannels}}
		DefaultLowInputLatency:    {{.DefaultLowInputLatency}}
		DefaultLowOutputLatency:   {{.DefaultLowOutputLatency}}
		DefaultHighInputLatency:   {{.DefaultHighInputLatency}}
		DefaultHighOutputLatency:  {{.DefaultHighOutputLatency}}
		DefaultSampleRate:         {{.DefaultSampleRate}}
	{{end}}
{{end}}`

func main() {
	portaudio.Initialize()
	defer portaudio.Terminate()
	hs, err := portaudio.HostApis()
	chk(err)
	t, err := template.New("").Parse(tmpl)
	chk(err)
	err = t.Execute(os.Stdout, hs)
	chk(err)
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}

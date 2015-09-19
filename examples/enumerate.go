package main

import (
	"github.com/gordonklaus/portaudio"
	"os"
	"text/template"
)

var tmpl = template.Must(template.New("").Parse(
	`{{. | len}} host APIs: {{range .}}
	Name:                   {{.Name}}
	{{if .DefaultInputDevice}}Default input device:   {{.DefaultInputDevice.Name}}{{end}}
	{{if .DefaultOutputDevice}}Default output device:  {{.DefaultOutputDevice.Name}}{{end}}
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
{{end}}`,
))

func main() {
	portaudio.Initialize()
	defer portaudio.Terminate()
	hs, err := portaudio.HostApis()
	chk(err)
	err = tmpl.Execute(os.Stdout, hs)
	chk(err)
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}

package main

import (
	"net/http"

	"github.com/manabr0w/golang-GUI/painter"
	"github.com/manabr0w/golang-GUI/painter/lang"
	"github.com/manabr0w/golang-GUI/ui"
)

func main() {
	var (
		pv ui.Visualizer

		opLoop painter.Loop
		parser lang.Parser
	)

	pv.Title = "Simple painter"

	pv.OnScreenReady = opLoop.Start
	opLoop.Receiver = &pv

	go func() {
		http.Handle("/", lang.HttpHandler(&opLoop, &parser))
		_ = http.ListenAndServe("localhost:17000", nil)
	}()

	pv.Main()
	opLoop.StopAndWait()
}

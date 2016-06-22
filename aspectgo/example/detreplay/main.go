package main

import (
	"golang.org/x/exp/aspectgo/example/detreplay/worker"
)

func main() {
	n := 16
	chans := make([]chan struct{}, n)
	for i := 0; i < n; i++ {
		w := worker.W{i}
		chans[i] = w.DoWork()
	}
	for i := 0; i < n; i++ {
		<-chans[i]
	}
}

package worker

import (
	"fmt"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type W struct {
	X int
}

func (w *W) doActualWork(workAmount int) {
	time.Sleep(time.Duration(workAmount) * time.Millisecond)
	fmt.Printf("hello from %d (r=%d)\n", w.X, workAmount)
}

func (w *W) DoWork() chan struct{} {
	ch := make(chan struct{})
	go func() {
		workAmount := int(rand.Int31n(8))
		w.doActualWork(workAmount)
		ch <- struct{}{}
	}()
	return ch
}

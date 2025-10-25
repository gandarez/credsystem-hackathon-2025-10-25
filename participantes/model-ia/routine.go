package modelia

import (
	"fmt"
	"math/rand"
	"time"
)

// Resultado de cada goroutine
type Result struct {
	Name  string
	Score float64
}

func goroutine1(ch chan Result) {
	time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
	ch <- Result{"Goroutine1", rand.Float64() * 1}
}

func goroutine2(ch chan Result) {
	time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
	ch <- Result{"Goroutine2", rand.Float64() * 1}
}

func Rotina() {
	rand.Seed(time.Now().UnixNano())

	ch1 := make(chan Result)
	ch2 := make(chan Result)

	go goroutine1(ch1)
	go goroutine2(ch2)

	var res1, res2 Result
	done := false

	for !done {
		select {
		case res1 = <-ch1:
			if res2.Name != "" {
				done = true
			}
		case res2 = <-ch2:
			if res2.Score > 0.8 {
				fmt.Println("Retornando Goroutine2 por score alto:", res2)
				return
			}
			if res1.Name != "" {
				done = true
			}
		}
	}

	if res2.Score > 0.8 {
		fmt.Println("Retornando Goroutine2:", res2)
	} else {
		fmt.Println("Retornando Goroutine1:", res1)
	}
}

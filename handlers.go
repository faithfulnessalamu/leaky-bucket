package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	RequestsCap = 5 //number at which we start dropping requests
	LeakRate    = 3 * time.Second
)

//handleAsAMeter implements Leaky Bucket as a Meter.
//See https://en.wikipedia.org/wiki/Leaky_bucket#As_a_meter
func handleAsAMeter() http.HandlerFunc {
	var (
		meter int
		mu    sync.Mutex
	)

	//constant leak rate
	ticker := time.NewTicker(LeakRate)
	go func(ticker *time.Ticker) {
		for range ticker.C {
			mu.Lock()
			if meter == 0 {
				mu.Unlock()
				continue
			}
			meter--
			mu.Unlock()
		}
	}(ticker)

	return func(w http.ResponseWriter, r *http.Request) {
		//attend to request if meter won't overflow, otherwise, drop request
		mu.Lock()
		if meter == RequestsCap {
			w.WriteHeader(http.StatusTooManyRequests)
			mu.Unlock()
			fmt.Fprintln(w, "Request dropped, overflowing.......")
			return
		}
		meter++
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Pong")
	}
}

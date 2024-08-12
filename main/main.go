package main

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requestCounter = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "aiml_infrastructure",
		Subsystem: "server",
		Name:      "request_counter",
		Help:      "the number of the API request",
	})
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		requestCounter.Inc()
		w.Write([]byte("Hello World!"))
	})

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		for {
			time.Sleep(10 * time.Second)
			requestCounter.Inc()
		}
	}()

	http.ListenAndServe(":8081", nil)
}

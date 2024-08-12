package prometheus

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusServer struct {
	Host string
}

func NewPrometheusServer(host string) PrometheusServer {
	return PrometheusServer{
		Host: host,
	}
}

func (p PrometheusServer) Start() error {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		http.ListenAndServe(p.Host, nil)
	}()

	return nil
}

func (p PrometheusServer) Stop() {
	return
}

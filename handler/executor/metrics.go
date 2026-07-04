package executor

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	CounterError = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "aiml-infrastructure",
		Subsystem: "executor",
		Name:      "counter_error",
		Help:      "measure the error of executing an executor job",
	})

	HistogramExecutionTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "aiml-infrastructure",
		Subsystem: "executor",
		Name:      "execution_time_second",
		Help:      "measure the execution time of an executor",
	}, []string{"task"})
)

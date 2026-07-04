package executor

import (
	"aiml-infrastructure/handler/payload"
	"aiml-infrastructure/package/log"
	"fmt"
	"sync"
	"time"
)

type ExecutorChannel[T any] struct {
	executor    IExecutor[T]
	nWorker     int
	payloadChan chan payload.Payload[T]
	doneChan    chan bool
	wgCounter   *sync.WaitGroup
}

func NewExecutorChannel[T any](executor IExecutor[T],
	nWorker int,
	payloadChan chan payload.Payload[T]) *ExecutorChannel[T] {

	counter := new(sync.WaitGroup)
	return &ExecutorChannel[T]{
		executor:    executor,
		nWorker:     nWorker,
		payloadChan: payloadChan,
		doneChan:    make(chan bool, nWorker),
		wgCounter:   counter,
	}
}

func (m *ExecutorChannel[T]) error(err error, methodName string, params ...interface{}) error {
	return fmt.Errorf("ExecutorChannel.(%s)(%s) %w", methodName, params, err)
}

func (m *ExecutorChannel[T]) Start() error {

	for i := 0; i < m.nWorker; i++ {
		m.wgCounter.Add(1)
		go func(i int) {
			// ensure wgCounter decremented before handler finish
			defer m.wgCounter.Done()

			// infinite loop to execute the payload
			for {
				select {

				// exit go routine if there is payload in doneChan
				case <-m.doneChan:
					return

				// execute if there is payload in payloadChan
				case payload := <-m.payloadChan:
					startTime := time.Now()
					err := m.executor.Execute(payload)
					executionTime := time.Since(startTime).Seconds()
					HistogramExecutionTime.WithLabelValues("executor_time").Observe(executionTime)

					if err != nil {
						CounterError.Inc()
						log.Error(m.error(err, "Start").Error())
					}
				}
			}
		}(i)
	}

	return nil
}

func (m *ExecutorChannel[T]) Stop() {
	for i := 0; i < m.nWorker; i++ {
		m.doneChan <- true
	}
	m.wgCounter.Wait()
	close(m.doneChan)
}

package executor

import "aiml-infrastructure/handler/payload"

type IExecutor[T any] interface {
	Execute(payload payload.Payload[T]) error
}

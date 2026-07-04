package payload

type Payload[T any] struct {
	Data      T
	Attribute map[string]string
}

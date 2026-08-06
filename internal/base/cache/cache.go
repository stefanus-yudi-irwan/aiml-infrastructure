package cache

type BackendCache string
type Value struct {
	Data  string
	Score float64
}

type KeyValue struct {
	Key   string
	Value Value
}

type Config struct {
	Backend                 BackendCache
	Address                 string
	Username                string
	Password                string
	Db                      int64
	DefaultSecondExpiration int64
}

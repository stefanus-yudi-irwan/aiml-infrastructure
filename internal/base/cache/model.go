package cache

import (
	"errors"
	"fmt"
)

type Value struct {
	Data  string
	Score float64
}

type KeyValue struct {
	Key   string
	Value Value
}

func (k *KeyValue) ValidateKeyValue() error {

	if k.Key == "" {
		return errors.New("Key cannot be empty")
	}

	if k.Value.Data == "" {
		return errors.New("Value.Data cannot be empty")
	}

	return nil
}

func (k *KeyValue) ValidateScoredKeyValue() error {
	if err := k.ValidateKeyValue(); err != nil {
		return err
	}

	if k.Value.Score <= 0 {
		return errors.New("Value.Score cannot be equal or less than 0")
	}

	return nil
}

type BackendCache string

const (
	BackendRedis     BackendCache = "redis"
	BackendValkey    BackendCache = "valkey"
	BackendDragonfly BackendCache = "dragonfly"
	BackendMemcached BackendCache = "memcached"
)

var RegisteredBackend = map[BackendCache]struct{}{
	BackendRedis:     {},
	BackendValkey:    {},
	BackendDragonfly: {},
	BackendMemcached: {},
}

type Config struct {
	Backend                 BackendCache
	Address                 string
	Username                string
	Password                string
	Db                      int64
	DefaultSecondExpiration int64
}

func (c *Config) Validate() error {
	if _, isExists := RegisteredBackend[c.Backend]; !isExists {
		return fmt.Errorf("unsuported cache: %s", c.Backend)
	}

	if c.Address == "" {
		return errors.New("Address cannot be empty")
	}

	if c.Db < 0 {
		return errors.New("Db num cannot be less than 0")
	}

	if c.DefaultSecondExpiration < 0 {
		return errors.New("DefaultSecondExpiration cannot be less than 0")
	}

	return nil
}

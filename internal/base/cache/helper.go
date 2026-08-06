package cache

import (
	"fmt"
	"reflect"
	"strings"
)

func CreateCacheKey(keys ...string) (string, error) {
	if len(keys) == 0 {
		return "", fmt.Errorf("%s: %s", "createCacheKey", "keys argument found empty")
	}
	var cacheKeys strings.Builder
	cacheKeys.WriteString(keys[0])
	for _, key := range keys[1:] {
		if key == "" {
			return "", fmt.Errorf("%s: %s", "createCacheKey", "keys argument consist of empty string")
		}
		cacheKeys.WriteRune('/')
		cacheKeys.WriteString(key)
	}
	return cacheKeys.String(), nil
}

func CreateCacheValue(data string) (Value, error) {
	if data == "" {
		return Value{}, fmt.Errorf("%s: %s", "createCacheValue", "empty data argument")
	}
	return Value{
		Data:  data,
		Score: 0,
	}, nil
}

func CreateCacheValueWithScore(data string, score float64) (Value, error) {
	if data == "" {
		return Value{}, fmt.Errorf("%s: %s", "createCacheValue", "empty data argument")
	}
	return Value{
		Data:  data,
		Score: score,
	}, nil
}

func CreateCacheKeyValue(key string, value Value) (KeyValue, error) {
	if key == "" {
		return KeyValue{}, fmt.Errorf("%s: %s", "createCacheKeyValue", "empty key argument")
	}

	if reflect.DeepEqual(value, Value{}) {
		return KeyValue{}, fmt.Errorf("%s: %s", "createCacheKeyValue", "empty value struct argument")
	}

	return KeyValue{
		Key:   key,
		Value: value,
	}, nil
}

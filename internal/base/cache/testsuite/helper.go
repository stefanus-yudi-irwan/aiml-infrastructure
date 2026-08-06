package testsuite

import (
	"aiml-infrastructure/internal/base/cache"
	"strconv"
)

func createKeyValueTest(keyName string, index int) cache.KeyValue {
	key, _ := cache.CreateCacheKey("test-key", keyName, strconv.Itoa(index))
	value, _ := cache.CreateCacheValue("test-value-" + key)
	keyValue, _ := cache.CreateCacheKeyValue(key, value)
	return keyValue
}

func createBatchKeyValueTest(keyName string, quantity int) []cache.KeyValue {

	var keyValues []cache.KeyValue

	for index := 0; index < quantity; index++ {
		keyValue := createKeyValueTest(keyName, index)
		keyValues = append(keyValues, keyValue)
	}

	return keyValues
}

func extractKeys(keyValues ...cache.KeyValue) []string {
	var keys []string
	for _, keyValue := range keyValues {
		keys = append(keys, keyValue.Key)
	}
	return keys
}

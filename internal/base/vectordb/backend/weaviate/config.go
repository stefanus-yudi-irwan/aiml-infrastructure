package weaviate

import (
	"errors"
	"unicode"
)

type WeaviateConnectorConfig struct {
	Host   string
	Scheme string
}

func (w *WeaviateConnectorConfig) Validate() error {
	if w.Host == "" {
		return errors.New("Host cannot be empty")
	}

	if w.Scheme == "" {
		return errors.New("Scheme cannot be empty")
	}

	return nil
}

type CollectionConfig struct {
	Name        string
	Description string
	Properties  []PropertyConfig
	Vectorizer  string
}

type PropertyConfig struct {
	Name         string
	Description  string
	DataType     []string
	Tokenization string
}

func (w *CollectionConfig) Validate() error {
	if w.Name == "" {
		return errors.New("Name cannot be empty")
	}

	if !unicode.IsUpper([]rune(w.Name)[0]) {
		return errors.New("collection name must start with an uppercase letter")
	}

	return nil
}

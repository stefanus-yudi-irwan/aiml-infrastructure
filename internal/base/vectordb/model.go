package vectordb

import (
	"errors"
	"reflect"
)

type Data struct {
	ID     string    `json:"id"`
	Vector []float32 `json:"vector,omitempty"`
	Fields Metadata  `json:"metadata,omitempty"`
}

type Metadata map[string]interface{}

func (d *Data) ValidateVector() error {
	if d.ID == "" {
		return errors.New("ID cannot be empty")
	}

	if d.Vector == nil {
		return errors.New("Vector cannot be empty")
	}
	return nil
}

func (d *Data) ValidateMetadata() error {
	if d.ID == "" {
		return errors.New("ID cannot be empty")
	}

	if reflect.ValueOf(d.Fields).IsZero() {
		return errors.New("Metadata cannot be empty struct")
	}
	return nil
}

// type Filter interface{}
// type SearchResult struct {
// 	Record Hash
// 	Score  float32
// }

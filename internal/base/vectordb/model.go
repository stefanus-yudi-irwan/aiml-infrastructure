package vectordb

import (
	"errors"
	"reflect"
)

type Data struct {
	ID         string      `json:"id"`
	DocumentID string      `json:"document_id"`
	ChunkID    string      `json:"chunk_id"`
	Content    string      `json:"content"`
	Embedding  []float32   `json:"vector,omitempty"`
	Metadata   MetadataMap `json:"metadata,omitempty"`
}

type MetadataMap map[string]interface{}

func (d *Data) ValidateIdentifier() error {
	if d.ID == "" {
		return errors.New("id cannot be empty")
	}

	if d.DocumentID == "" {
		return errors.New("document id cannot be empty")
	}

	if d.ChunkID == "" {
		return errors.New("chunk id cannot be empty")
	}

	return nil
}

func (d *Data) ValidateContent() error {
	if d.Content == "" {
		return errors.New("content cannot be empty")
	}
	return nil
}

func (d *Data) ValidateEmbedding() error {
	if len(d.Embedding) == 0 {
		return errors.New("embedding cannot be empty")
	}
	return nil
}

func (d *Data) ValidateMetadata() error {
	if reflect.ValueOf(d.Metadata).IsZero() {
		return errors.New("metadata cannot be empty")
	}
	return nil
}

type SearchResult struct {
	Data  Data
	Score float32
}

type SearchRequest struct {
	Embedding []float32
	TopK      int
	Filter    *Filter
}

type Filter struct {
	Must    []Condition `json:"must,omitempty"`
	Should  []Condition `json:"should,omitempty"`
	MustNot []Condition `json:"must_not,omitempty"`
}

type Condition struct {
	Field    string   `json:"field"`
	Operator Operator `json:"operator"`
	Value    any      `json:"value"`
}

type Operator string

const (
	OpEqual        Operator = "eq"
	OpNotEqual     Operator = "neq"
	OpIn           Operator = "in"
	OpNotIn        Operator = "not_in"
	OpGreaterThan  Operator = "gt"
	OpGreaterEqual Operator = "gte"
	OpLessThan     Operator = "lt"
	OpLessEqual    Operator = "lte"
)

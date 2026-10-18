package vectordb

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	// main data
	IDColumnName              string = "id"
	DocumentIDColumnName      string = "document_id"
	ChunkIDColumnName         string = "chunk_id"
	ContentColumnName         string = "content"
	SparseEmbeddingColumnName string = "sparse_embedding"
	DenseEmbeddingColumnName  string = "dense_embedding"
	MetadataColumnName        string = "metadata"

	//embeddings data
	SparseEmbedding1Name string = "sparse_1"
	SparseEmbedding2Name string = "sparse_2"
	DenseEmbedding1Name  string = "dense_1"
	DenseEmbedding2Name  string = "dense_2"
)

type SparseEmbedding struct {
	Indices []uint32  `json:"indices"`
	Values  []float32 `json:"values"`
}

func (d *SparseEmbedding) Validate() error {
	if len(d.Indices) == 0 {
		return errors.New("value of indices cannot be empty")
	}

	if len(d.Values) == 0 {
		return errors.New("value of sparse embedding cannot be empty")
	}

	if len(d.Indices) != len(d.Values) {
		return fmt.Errorf("length of indices and values differ, indices length: %d value length: %d",
			len(d.Indices),
			len(d.Values))
	}

	return nil
}

type DenseEmbedding struct {
	Values []float32 `json:"values"`
}

func (d *DenseEmbedding) Validate() error {
	if len(d.Values) == 0 {
		return errors.New("value of dense embedding cannot be empty")
	}
	return nil
}

type Data struct {
	ID               string                     `json:"id"`
	DocumentID       string                     `json:"document_id"`
	ChunkID          string                     `json:"chunk_id"`
	Content          string                     `json:"content"`
	SparseEmbeddings map[string]SparseEmbedding `json:"sparse_embeddings,omitempty"`
	DenseEmbeddings  map[string]DenseEmbedding  `json:"dense_embeddings,omitempty"`
	Metadata         map[string]interface{}     `json:"metadata,omitempty"`
}

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
	if len(d.Content) == 0 {
		return errors.New("content cannot be empty")
	}
	return nil
}

func (d *Data) ValidateSparseEmbeddings() error {
	if len(d.SparseEmbeddings) == 0 {
		return errors.New("spare embeddings cannot be empty")
	}

	for _, value := range d.SparseEmbeddings {
		if err := value.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (d *Data) ValidateDenseEmbeddings() error {
	if len(d.DenseEmbeddings) == 0 {
		return errors.New("dense embeddings cannot be empty")
	}

	for _, value := range d.DenseEmbeddings {
		if err := value.Validate(); err != nil {
			return err
		}
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

var (
	DataTypeVarchar    = "varchar"
	DataTypeText       = "text"
	DataTypeUUID       = "uuid"
	DataTypeInteger    = "integer"
	DataTypeBigInteger = "bigint"
	DataTypeFloat      = "float"
	DataTypeDouble     = "double"
	DataTypeBoolean    = "bool"
	DataTypeJSON       = "json"
	DataTypeTimestamp  = "timestamp"
	DataTypeTimestampz = "timestampz"
	DataTypeVector     = "vector"
)

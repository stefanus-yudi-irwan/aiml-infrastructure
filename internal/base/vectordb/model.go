package vectordb

import (
	"errors"
	"fmt"
)

var (
	// main data
	IDColumnName               string = "id"
	DocumentIDColumnName       string = "document_id"
	ChunkIDColumnName          string = "chunk_id"
	ContentColumnName          string = "content"
	SparseEmbeddingsColumnName string = "sparse_embeddings"
	DenseEmbeddingsColumnName  string = "dense_embeddings"
	MetadataColumnName         string = "metadata"
)

type SparseEmbeddingName string

var (
	SparseEmbedding1Name SparseEmbeddingName = "sparse_1"
	SparseEmbedding2Name SparseEmbeddingName = "sparse_2"
)

var RegisteredSparseEmbedding = map[SparseEmbeddingName]struct{}{
	SparseEmbedding1Name: {},
	SparseEmbedding2Name: {},
}

type DenseEmbeddingName string

var (
	DenseEmbedding1Name DenseEmbeddingName = "dense_1"
	DenseEmbedding2Name DenseEmbeddingName = "dense_2"
)

var RegisteredDenseEmbedding = map[DenseEmbeddingName]struct{}{
	DenseEmbedding1Name: {},
	DenseEmbedding2Name: {},
}

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

type Data struct {
	ID               string                                  `json:"id"`
	DocumentID       string                                  `json:"document_id"`
	ChunkID          string                                  `json:"chunk_id"`
	Content          string                                  `json:"content"`
	SparseEmbeddings map[SparseEmbeddingName]SparseEmbedding `json:"sparse_embeddings,omitempty"`
	DenseEmbeddings  map[DenseEmbeddingName][]float32        `json:"dense_embeddings,omitempty"`
	Metadata         map[string]interface{}                  `json:"metadata,omitempty"`
}

func (d *Data) Validate() error {
	if err := d.ValidateIdentifier(); err != nil {
		return err
	}

	if err := d.ValidateContent(); err != nil {
		return err
	}

	if d.SparseEmbeddings != nil {
		if err := d.ValidateSparseEmbeddings(); err != nil {
			return err
		}
	}

	if d.DenseEmbeddings != nil {
		if err := d.ValidateDenseEmbeddings(); err != nil {
			return err
		}
	}

	if d.Metadata != nil {
		if err := d.ValidateMetadata(); err != nil {
			return err
		}
	}

	return nil
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

	for name, embedding := range d.SparseEmbeddings {
		if _, isExists := RegisteredSparseEmbedding[name]; !isExists {
			return errors.New("unsuported sparse embedding name")
		}

		if err := embedding.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (d *Data) ValidateDenseEmbeddings() error {
	if len(d.DenseEmbeddings) == 0 {
		return errors.New("dense embeddings cannot be empty")
	}

	for name, embedding := range d.DenseEmbeddings {
		if _, isExists := RegisteredDenseEmbedding[name]; !isExists {
			return errors.New("unsuported dense embedding name")
		}

		if len(embedding) == 0 {
			return errors.New("dense vector length cannot be empty")
		}

	}
	return nil
}

func (d *Data) ValidateMetadata() error {
	if len(d.Metadata) == 0 {
		return errors.New("metadata cannot be empty")
	}

	for key, value := range d.Metadata {
		switch value.(type) {
		case bool, []bool, string, []string, float64, []float64:
			//valid metadata type
		default:
			return fmt.Errorf("metadata field %q has unsupported type %T",
				key,
				value)
		}
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

package weaviate

import (
	"errors"
	"unicode"
)

type VectorDBConnectorConfig struct {
	Host   string
	Scheme string
}

func (w *VectorDBConnectorConfig) Validate() error {
	if w.Host == "" {
		return errors.New("host cannot be empty")
	}

	if w.Scheme == "" {
		return errors.New("scheme cannot be empty")
	}

	return nil
}

type VectorIndexType string

var (
	HNSWIndexType    VectorIndexType = "hnsw"
	FlatIndexType    VectorIndexType = "flat"
	DynamicIndexType VectorIndexType = "dynamic"
	HFreshIndexType  VectorIndexType = "hfresh"
)

type DistanceType string

var (
	DistanceCosine   DistanceType = "cosine"
	DistanceDot      DistanceType = "dot"
	DistanceEuclid   DistanceType = "l2-squared"
	DistanceHamming  DistanceType = "hamming"
	DistanceMahattan DistanceType = "manhattan"
)

const (
	EF_SMALL  = 100
	EF_MEDIUM = 200
	EF_LARGE  = 300
	EF_JUMBO  = 400
)

const (
	MAX_CONNECTIONS_SMALL  = 8
	MAX_CONNECTIONS_MEDIUM = 16
	MAX_CONNECTIONS_LARGE  = 32
	MAX_CONNECTIONS_JUMBO  = 64
)

const (
	EF_CONSTRUCTION_SMALL  = 100
	EF_CONSTRUCTION_MEDIUM = 200
	EF_CONSTRUCTION_LARGE  = 300
	EF_CONSTRUCTION_JUMBO  = 400
)

type CollectionConfig struct {
	Name        string
	Description string
	Dimension   int64
	VectorIndex VectorIndexType
	Distance    DistanceType
	HNSW        *HNSWConfig
	Fields      []FieldConfig
	Vectorizer  string
}

type FieldConfig struct {
	Name         string
	Description  string
	DataType     []string
	Tokenization string
	Fields       []FieldConfig
}

type HNSWConfig struct {
	EF             int
	EFConstruction int
	MaxConnections int
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

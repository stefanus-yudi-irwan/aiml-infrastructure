package qdrant

import (
	"aiml-infrastructure/internal/base/vectordb"
	"errors"
	"fmt"

	"github.com/qdrant/go-client/qdrant"
)

type VectorDBConnectorConfig struct {
	Host   string
	Port   int
	APIKey string
}

func (q *VectorDBConnectorConfig) Validate() error {
	if q.Host == "" {
		return errors.New("host cannot be empty")
	}

	if q.Port == 0 {
		return errors.New("port cannot be empty")
	}

	if q.APIKey == "" {
		return errors.New("api key cannot be empty")
	}

	return nil
}

type CollectionConfig struct {
	Name                   string
	DenseEmbeddingFields   []*DenseEmbeddingFieldConfig
	SparseEmbeddingFields  []*SparseEmbeddingFieldConfig
	ShardNumber            uint32
	ReplicationFactor      uint32
	WriteConsistencyFactor uint32
}

func (v *CollectionConfig) Validate() error {
	if v.Name == "" {
		return errors.New("field name cannot be empty")
	}

	if v.DenseEmbeddingFields == nil {
		return errors.New("dense embedding field need to have configuration")
	}

	for _, embeddingField := range v.DenseEmbeddingFields {
		if err := embeddingField.Validate(); err != nil {
			return err
		}
	}

	return nil
}

type DenseEmbeddingFieldConfig struct {
	Name      vectordb.DenseEmbeddingName
	Dimension int
	Distance  DistanceType
	Index     IndexType
	HNSW      *HNSWConfig
}

func (v *DenseEmbeddingFieldConfig) Validate() error {
	if v.Name == "" {
		return errors.New("field name cannot be empty")
	}

	if v.Dimension < 0 {
		return errors.New("dense embedding dimension cannot be negative")
	}

	switch v.Distance {
	case COSINE, L2, IP, MANHATTAN:
	default:
		return errors.New("invalid distance type")
	}

	switch v.Index {
	case HNSW:
		if err := v.HNSW.Validate(); err != nil {
			return err
		}
	default:
		return errors.New("invalid index type")
	}

	return nil
}

type SparseEmbeddingFieldConfig struct {
	Name vectordb.SparseEmbeddingName
}

func (v *SparseEmbeddingFieldConfig) Validate() error {
	if v.Name == "" {
		return errors.New("field name cannot be empty")
	}

	return nil
}

type IndexType string

var (
	HNSW IndexType = "hnsw"
)

type DistanceType string

var (
	COSINE    DistanceType = "cosine"
	L2        DistanceType = "euclidean"
	IP        DistanceType = "inner_product"
	MANHATTAN DistanceType = "manhattan"
)

func toQdrantDistance(distanceType DistanceType) qdrant.Distance {
	switch distanceType {
	case COSINE:
		return qdrant.Distance_Cosine
	case L2:
		return qdrant.Distance_Euclid
	case IP:
		return qdrant.Distance_Dot
	case MANHATTAN:
		return qdrant.Distance_Manhattan
	default:
		return qdrant.Distance_UnknownDistance
	}
}

// HNSW graph connectivity
const (
	HNSW_M_SMALL  = 16
	HNSW_M_MEDIUM = 24
	HNSW_M_LARGE  = 32
	HNSW_M_JUMBO  = 40
)

// HNSW index construction
const (
	HNSW_EF_CONSTRUCTION_SMALL  = 100
	HNSW_EF_CONSTRUCTION_MEDIUM = 200
	HNSW_EF_CONSTRUCTION_LARGE  = 300
	HNSW_EF_CONSTRUCTION_JUMBO  = 400
)

// HNSW qdrant full scan threshold
const (
	HNSW_FULL_SCAN_THRESHOLD_SMALL  = 1000
	HNSW_FULL_SCAN_THRESHOLD_MEDIUM = 10000
	HNSW_FULL_SCAN_THRESHOLD_LARGE  = 100000
	HNSW_FULL_SCAN_THRESHOLD_JUMBO  = 1000000
)

type HNSWConfig struct {
	M                 int
	EFConstruction    int
	FullScanThreshold int
}

func (v *HNSWConfig) Validate() error {
	if v == nil {
		return errors.New("hnsw config cannot be nil")
	}

	if v.M < HNSW_M_SMALL || v.M > HNSW_M_JUMBO {
		return fmt.Errorf("hnsw m value must be between %d and %d",
			HNSW_M_SMALL,
			HNSW_M_JUMBO)
	}

	if v.EFConstruction < HNSW_EF_CONSTRUCTION_SMALL || v.EFConstruction > HNSW_EF_CONSTRUCTION_JUMBO {
		return fmt.Errorf("hnsw efconstruction value must be between %d and %d",
			HNSW_EF_CONSTRUCTION_SMALL,
			HNSW_EF_CONSTRUCTION_JUMBO)
	}

	if v.FullScanThreshold <= 0 {
		return fmt.Errorf("full scan threshold  value must be between %d and %d",
			HNSW_FULL_SCAN_THRESHOLD_SMALL,
			HNSW_FULL_SCAN_THRESHOLD_JUMBO)
	}

	return nil
}

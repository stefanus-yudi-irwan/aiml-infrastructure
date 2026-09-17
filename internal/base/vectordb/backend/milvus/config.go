package milvus

import (
	"errors"

	"github.com/milvus-io/milvus/client/v2/entity"
)

type MilvusConnectorConfig struct {
	Address  string
	Username string
	Password string
	Db       string
}

func (v *MilvusConnectorConfig) Validate() error {
	if v.Address == "" {
		return errors.New("Address cannot be empty")
	}

	if v.Username == "" {
		return errors.New("Username cannot be empty")
	}

	if v.Password == "" {
		return errors.New("Password cannot be empty")
	}

	if v.Db == "" {
		return errors.New("Db cannot be empty")
	}
	return nil
}

type CollectionConfig struct {
	Name        string
	Description string
	Fields      []FieldConfig
	ShardNum    int32
}

func (v *CollectionConfig) Validate() error {
	if v.Name == "" {
		return errors.New("Collection name cannot be empty")
	}

	if v.Description == "" {
		return errors.New("Collection description cannot be empty")
	}

	if len(v.Fields) == 0 {
		return errors.New("Collection fields cannot be empty")
	}

	for _, field := range v.Fields {
		if err := field.Validate(); err != nil {
			return err
		}
	}

	if v.ShardNum < 0 {
		return errors.New("Collection shard number cannot be negative")
	}

	return nil
}

type FieldConfig struct {
	Name         string
	DataType     entity.FieldType
	PrimaryKey   bool
	AutoID       bool
	Dimension    int64
	MaxLength    int64
	VectorConfig *VectorIndexConfig
}

func (v *FieldConfig) Validate() error {
	if v.Name == "" {
		return errors.New("Field name cannot be empty")
	}

	if v.DataType == entity.FieldTypeNone {
		return errors.New("Field data type cannot be empty")
	}

	if v.Dimension < 0 {
		return errors.New("Field dimension cannot be negative")
	}

	if v.MaxLength < 0 {
		return errors.New("Field max length cannot be negative")
	}

	return nil
}

type VectorIndexAlgorithm string

const (
	HNSW       VectorIndexAlgorithm = "HNSW"
	FLAT       VectorIndexAlgorithm = "FLAT"
	DISKANN    VectorIndexAlgorithm = "DISKANN"
	IVFFLAT    VectorIndexAlgorithm = "IVF_FLAT"
	IVFPQ      VectorIndexAlgorithm = "IVF_PQ"
	IVFSQ8     VectorIndexAlgorithm = "IVF_SQ8"
	GPUIVFFLAT VectorIndexAlgorithm = "GPU_IVF_FLAT"
	GPUIVFPQ   VectorIndexAlgorithm = "GPU_IVF_PQ"
)

const (
	COSINE = entity.COSINE
	L2     = entity.L2
	IP     = entity.IP
)

const (
	M_SMALL  = 16
	M_MEDIUM = 24
	M_LARGE  = 32
	M_JUMBO  = 40
)

const (
	EF_CONSTRUCTION_SMALL  = 100
	EF_CONSTRUCTION_MEDIUM = 200
	EF_CONSTRUCTION_LARGE  = 300
	EF_CONSTRUCTION_JUMBO  = 400
)

type VectorIndexConfig struct {
	Algorithm VectorIndexAlgorithm
	Metric    entity.MetricType
	HNSW      *HNSWConfig
}

func (v *VectorIndexConfig) Validate() error {
	if v.Algorithm == "" {
		return errors.New("Vector algorithm cannot be empty")
	}

	if v.Metric == "" {
		return errors.New("Vector metric cannot be empty")
	}

	if v.Algorithm == HNSW && v.HNSW == nil {
		return errors.New("HNSW config cannot be nil for HNSW algorithm")
	}

	if v.HNSW != nil {
		if err := v.HNSW.Validate(); err != nil {
			return err
		}
	}

	return nil
}

type HNSWConfig struct {
	M              int
	EFConstruction int
}

func (v *HNSWConfig) Validate() error {
	if v.M < M_SMALL || v.M > M_JUMBO {
		return errors.New("HNSW M value must be between 16 and 40")
	}

	if v.EFConstruction < EF_CONSTRUCTION_SMALL || v.EFConstruction > EF_CONSTRUCTION_JUMBO {
		return errors.New("HNSW EFConstruction value must be between 100 and 400")
	}

	return nil
}

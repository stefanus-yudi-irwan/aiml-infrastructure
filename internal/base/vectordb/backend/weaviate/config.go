package weaviate

import (
	"aiml-infrastructure/internal/base/vectordb"
	"errors"
	"fmt"
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

const (
	HNSW_EF_RUNTIME_SMALL  = 100
	HNSW_EF_RUNTIME_MEDIUM = 200
	HNSW_EF_RUNTIME_LARGE  = 300
	HNSW_EF_RUNTIME_JUMBO  = 400
)

const (
	HNSW_EF_CONSTRUCTION_SMALL  = 100
	HNSW_EF_CONSTRUCTION_MEDIUM = 200
	HNSW_EF_CONSTRUCTION_LARGE  = 300
	HNSW_EF_CONSTRUCTION_JUMBO  = 400
)

const (
	HNSW_M_SMALL  = 8
	HNSW_M_MEDIUM = 16
	HNSW_M_LARGE  = 32
	HNSW_M_JUMBO  = 64
)

type CollectionConfig struct {
	Name                  string
	Description           string
	DenseEmbeddingFields  []*DenseEmbeddingFieldConfig
	SparseEmbeddingFields []*SparseEmbeddingFieldConfig
	DataFields            []*DataFieldConfig
}

func (v *CollectionConfig) Validate() error {
	if v.Name == "" {
		return errors.New("Name cannot be empty")
	}

	if !unicode.IsUpper([]rune(v.Name)[0]) {
		return errors.New("collection name must start with an uppercase letter")
	}

	if v.DenseEmbeddingFields == nil {
		return errors.New("embedding field need to have configuration")
	}

	for _, embeddingField := range v.DenseEmbeddingFields {
		if err := embeddingField.Validate(); err != nil {
			return err
		}
	}

	if len(v.DataFields) == 0 {
		return errors.New("collection fields cannot be empty")
	}

	for _, dataField := range v.DataFields {
		if err := dataField.Validate(); err != nil {
			return err
		}
	}

	return nil
}

type DenseEmbeddingFieldConfig struct {
	Name       vectordb.DenseEmbeddingName
	Dimension  int
	Distance   DistanceType
	Index      IndexType
	HNSW       *HNSWConfig
	Vectorizer string
}

func (v *DenseEmbeddingFieldConfig) Validate() error {
	if v.Dimension < 0 {
		return errors.New("dimension cannot be less than 0")
	}

	if _, isExists := RegisteredDistanceType[v.Distance]; !isExists {
		return fmt.Errorf("unsuported Metric: %s", v.Distance)
	}

	if _, isExists := RegisteredIndexType[v.Index]; !isExists {
		return fmt.Errorf("unsuported Algorithm: %s", v.Index)
	}

	if v.Index == HNSW {
		if err := v.HNSW.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type HNSWConfig struct {
	M              int
	EFConstruction int
	EFRuntime      int
}

func (v *HNSWConfig) Validate() error {
	if v == nil {
		return errors.New("HNSW config is empty when attempting to create HNSW index")
	}

	if v.M < HNSW_M_SMALL || v.M > HNSW_M_JUMBO {
		return fmt.Errorf("unsuported graph connection value, need to be between %d and %d: %d",
			HNSW_M_SMALL,
			HNSW_M_JUMBO,
			v.M)
	}

	if v.EFConstruction < HNSW_EF_CONSTRUCTION_SMALL || v.EFConstruction > HNSW_EF_CONSTRUCTION_JUMBO {
		return fmt.Errorf("unsuported ef construction size, need to be between %d and %d: %d",
			HNSW_EF_CONSTRUCTION_SMALL,
			HNSW_EF_CONSTRUCTION_JUMBO,
			v.EFConstruction)
	}

	if v.EFRuntime < HNSW_EF_RUNTIME_SMALL || v.EFRuntime > HNSW_EF_RUNTIME_JUMBO {
		return fmt.Errorf("unsuported ef runtime size, need to be between %d and %d: %d",
			HNSW_EF_RUNTIME_SMALL,
			HNSW_EF_RUNTIME_JUMBO,
			v.EFRuntime)
	}
	return nil
}

type SparseEmbeddingFieldConfig struct {
	Name         vectordb.SparseEmbeddingName
	Description  string
	DataType     DataFieldType
	Tokenization string
}

func (v *SparseEmbeddingFieldConfig) Validate() error {
	if v.Name == "" {
		return errors.New("field name cannot be nil")
	}

	if len(v.DataType) == 0 {
		return errors.New("data type cannot be empty")
	}

	return nil
}

type DataFieldConfig struct {
	Name         string
	Description  string
	DataType     DataFieldType
	Tokenization string
	Fields       []DataFieldConfig
}

func (v *DataFieldConfig) Validate() error {
	if v.Name == "" {
		return errors.New("field name cannot be nil")
	}

	if len(v.DataType) == 0 {
		return errors.New("data type cannot be empty")
	}

	if len(v.Fields) > 0 {
		for _, field := range v.Fields {
			if err := field.Validate(); err != nil {
				return err
			}
		}
	}
	return nil
}

type IndexType string

var (
	HNSW    IndexType = "HNSW"
	FLAT    IndexType = "FLAT"
	DYNAMIC IndexType = "DYNAMIC"
	HFRESH  IndexType = "HFRESH"
)

var RegisteredIndexType = map[IndexType]struct{}{
	HNSW:    {},
	FLAT:    {},
	DYNAMIC: {},
	HFRESH:  {},
}

func toWeaviateIndexType(index IndexType) (string, error) {
	switch index {
	case HNSW:
		return "hnsw", nil
	case FLAT:
		return "flat", nil
	case DYNAMIC:
		return "dynamic", nil
	case HFRESH:
		return "hfresh", nil
	default:
		return "", errors.New("index not supported")
	}
}

type DistanceType string

var (
	COSINE    DistanceType = "COSINE"
	IP        DistanceType = "DOT"
	L2        DistanceType = "L2"
	HAMMING   DistanceType = "HAMMING"
	MANHATTAN DistanceType = "MANHATTAN"
)

var RegisteredDistanceType = map[DistanceType]struct{}{
	COSINE:    {},
	L2:        {},
	IP:        {},
	HAMMING:   {},
	MANHATTAN: {},
}

func toWeaviateDistanceType(distance DistanceType) (string, error) {
	switch distance {
	case COSINE:
		return "cosine", nil
	case IP:
		return "dot", nil
	case L2:
		return "l2-squared", nil
	case HAMMING:
		return "hamming", nil
	case MANHATTAN:
		return "manhattan", nil
	default:
		return "", errors.New("distance type not supported")
	}
}

type DataFieldType string

var (
	DataFieldFloat      DataFieldType = "float"
	DataFieldFloatArray DataFieldType = "[]float"
	DataFieldBool       DataFieldType = "bool"
	DataFieldBoolArray  DataFieldType = "[]bool"
	DataFieldText       DataFieldType = "text"
	DataFieldTextArray  DataFieldType = "[]text"
	DataFieldObject     DataFieldType = "object"
)

func toWeaviateDataType(dataType DataFieldType) ([]string, error) {
	switch dataType {
	case DataFieldFloat:
		return []string{"number"}, nil
	case DataFieldFloatArray:
		return []string{"number[]"}, nil
	case DataFieldBool:
		return []string{"boolean"}, nil
	case DataFieldBoolArray:
		return []string{"boolean[]"}, nil
	case DataFieldText:
		return []string{"text"}, nil
	case DataFieldTextArray:
		return []string{"text[]"}, nil
	case DataFieldObject:
		return []string{"object"}, nil
	default:
		return nil, errors.New("dataType not supported")
	}
}

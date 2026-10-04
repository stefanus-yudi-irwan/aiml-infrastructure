package qdrant

import (
	"errors"

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

var (
	DistanceCosine = qdrant.Distance_Cosine
	DistanceDot    = qdrant.Distance_Dot
	DistanceEuclid = qdrant.Distance_Euclid
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

const (
	FULL_SCAN_THRESHOLD_SMALL  = 1000
	FULL_SCAN_THRESHOLD_MEDIUM = 10000
	FULL_SCAN_THRESHOLD_LARGE  = 100000
	FULL_SCAN_THRESHOLD_JUMBO  = 1000000
)

type CollectionConfig struct {
	Name                   string
	Dimension              uint64
	Distance               qdrant.Distance
	HNSW                   *HNSWConfig
	ShardNumber            uint32
	ReplicationFactor      uint32
	WriteConsistencyFactor uint32
}

type HNSWConfig struct {
	M                 int
	EFConstruction    int
	FullScanThreshold int
}

func (q *CollectionConfig) Validate() error {
	if q.Name == "" {
		return errors.New("name cannot be empty")
	}

	if q.Dimension == 0 {
		return errors.New("dimension cannot be empty")
	}

	switch q.Distance {
	case DistanceCosine, DistanceDot, DistanceEuclid:
	default:
		return errors.New("invalid distance")
	}

	if q.HNSW != nil {
		if q.HNSW.M <= 0 {
			return errors.New("hnsw m cannot be empty")
		}

		if q.HNSW.EFConstruction <= 0 {
			return errors.New("hnsw ef construction cannot be empty")
		}

		if q.HNSW.FullScanThreshold <= 0 {
			return errors.New("hnsw full scan threshold cannot be empty")
		}
	}

	return nil
}

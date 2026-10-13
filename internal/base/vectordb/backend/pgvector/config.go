package pgvector

import (
	"errors"
	"time"
)

type VectorDBConnectorConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

func (p *VectorDBConnectorConfig) Validate() error {
	if p.Host == "" {
		return errors.New("host cannot be empty")
	}

	if p.Port == 0 {
		return errors.New("port cannot be zero")
	}

	if p.User == "" {
		return errors.New("user cannot be empty")
	}

	if p.Password == "" {
		return errors.New("password cannot be empty")
	}

	if p.Database == "" {
		return errors.New("database cannot be empty")
	}

	if p.SSLMode != "enable" || p.SSLMode != "disable" || p.SSLMode == "" {
		return errors.New("undefined or empty ssl mode")
	}

	if p.MaxConns == 0 {
		return errors.New("max conns cannot be zero")
	}

	if p.MinConns == 0 {
		return errors.New("min conns cannot be zero")
	}

	if p.MaxConnLifetime == 0 {
		return errors.New("max conns lifetime cannot be zero")
	}

	if p.MaxConnIdleTime == 0 {
		return errors.New("max conns idletime cannot be zero")
	}

	return nil
}

type IndexType string

var (
	HNSWIndex    IndexType = "hnsw"
	IVFFlatIndex IndexType = "ivfflat"
)

type DistanceType string

var (
	DistanceCosine       DistanceType = "cosine"
	DistanceEuclidean    DistanceType = "euclidean"
	DistanceInnerProduct DistanceType = "inner_product"
	DistanceManhattan    DistanceType = "manhattan"
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

type CollectionConfig struct {
	SchemaName         string
	CollectionName     string
	EmbeddingDimension int
	Index              IndexType
	Distance           DistanceType
	HNSW               *HNSWConfig
	IVF                *IVFConfig
}

type HNSWConfig struct {
	M              int
	EFConstruction int
}

type IVFConfig struct {
	IVFFlatLists int
}

func (p *CollectionConfig) Validate() error {
	if p.SchemaName == "" {
		return errors.New("schema name cannot be empty")
	}

	if p.CollectionName == "" {
		return errors.New("collection name cannot be empty")
	}

	if p.EmbeddingDimension == 0 {
		return errors.New("vector dimension cannot be zero")
	}

	switch p.Index {
	case HNSWIndex:
		if p.HNSW == nil {
			return errors.New("hnsw config cannot be nil")
		}
	case IVFFlatIndex:
		if p.IVF == nil {
			return errors.New("ivfflat config cannot be nil")
		}
	default:
		return errors.New("invalid index type")
	}

	switch p.Distance {
	case DistanceCosine, DistanceEuclidean, DistanceInnerProduct, DistanceManhattan:
	default:
		return errors.New("invalid distance type")
	}

	return nil
}

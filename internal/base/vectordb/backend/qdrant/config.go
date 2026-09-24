package qdrant

import "errors"

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

type Distance string

var (
	DistanceCosine Distance = "cosine"
	DistanceDot    Distance = "dot"
	DistanceEuclid Distance = "euclid"
)

type CollectionConfig struct {
	Name                   string
	Dimension              uint64
	Distance               Distance
	ShardNumber            uint32
	ReplicationFactor      uint32
	WriteConsistencyFactor uint32
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

	return nil
}

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

type CollectionConfig struct {
	Name               string
	EmbeddingDimension int
}

func (p *CollectionConfig) Validate() error {
	if p.Name == "" {
		return errors.New("collection name cannot be empty")
	}

	if p.EmbeddingDimension == 0 {
		return errors.New("vector dimension cannot be zero")
	}

	return nil
}

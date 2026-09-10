package milvus

import "errors"

type MilvusConnectorConfig struct {
	Address     string
	Username    string
	Password    string
	Db          string
	Collections []CollectionConfig
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
	Name   string
	Fields []FieldConfig
	Index  IndexConfig
}

type FieldConfig struct {
	Name       string
	Type       FieldType
	PrimaryKey bool

	// Vector-specific
	Dimension int
	Metric    DistanceMetric

	// String-specific
	MaxLength int

	// Array-specific
	ElementType FieldType
}

type IndexConfig struct {
	Type   IndexType
	Metric DistanceMetric
}

type FieldType int

const (
	FieldTypeBool FieldType = iota
	FieldTypeInt8
	FieldTypeInt16
	FieldTypeInt32
	FieldTypeInt64
	FieldTypeFloat
	FieldTypeDouble
	FieldTypeString
	FieldTypeJSON
	FieldTypeArray
	FieldTypeFloatVector
	FieldTypeBinaryVector
	FieldTypeFloat16Vector
	FieldTypeBFloat16Vector
	FieldTypeInt8Vector
	FieldTypeSparseFloatVector
	FieldTypeGeometry
)

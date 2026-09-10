package redisvector

import (
	"errors"
	"fmt"
	"math"
)

type VectorDBClientConfig struct {
	Address     string
	Username    string
	Password    string
	Db          int
	Collections []CollectionConfig
}

func (v *VectorDBClientConfig) Validate() error {
	if v.Address == "" {
		return errors.New("Address cannot be empty")
	}
	if v.Db < 0 {
		return errors.New("Db number cannot be less than 0")
	}
	return nil
}

type IndexAlgorithm string

const (
	HNSW IndexAlgorithm = "HNSW"
	FLAT IndexAlgorithm = "FLAT"
)

var RegisteredIndexAlgorithm = map[IndexAlgorithm]struct{}{
	HNSW: {},
	FLAT: {},
}

type DistanceMetric string

const (
	COSINE DistanceMetric = "COSINE"
	L2     DistanceMetric = "L2"
	IP     DistanceMetric = "IP"
)

var RegisteredDistanceMetric = map[DistanceMetric]struct{}{
	COSINE: {},
	L2:     {},
	IP:     {},
}

type VectorType string

const (
	VECTOR_FLOAT32 VectorType = "FLOAT32"
	VECTOR_FLOAT64 VectorType = "FLOAT64"
)

var RegisteredVectorType = map[VectorType]struct{}{
	VECTOR_FLOAT32: {},
	VECTOR_FLOAT64: {},
}

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
	EF_RUNTIME_SMALL  = 10
	EF_RUNTIME_MEDIUM = 20
	EF_RUNTIME_LARGE  = 50
	EF_RUNTIME_JUMBO  = 100
)

type FieldType string

const (
	FieldTypeVector  FieldType = "VECTOR"
	FieldTypeTag     FieldType = "TAG"
	FieldTypeText    FieldType = "TEXT"
	FieldTypeNumeric FieldType = "NUMERIC"
	FieldTypeGeo     FieldType = "GEO"
)

type IndexType string

const (
	HASH_TYPE IndexType = "HASH"
	JSON_TYPE IndexType = "JSON"
)

type CollectionConfig struct {
	Name   string
	On     IndexType
	Fields []FieldConfig
}

func (c *CollectionConfig) Validate() error {

	if c.Name == "" {
		return errors.New("collection name cannot be empty")
	}

	if c.On == "" {
		return errors.New("collection index type cannot be empty")
	}

	if c.Fields == nil || len(c.Fields) == 0 {
		return errors.New("collection fields cannot be empty")
	}

	if c.Fields != nil {
		for _, field := range c.Fields {
			switch field.Type {
			case FieldTypeVector:
				if err := field.VectorField.Validate(); err != nil {
					return err
				}
			case FieldTypeTag:
				if err := field.TagField.Validate(); err != nil {
					return err
				}
			case FieldTypeText:
				if err := field.TextField.Validate(); err != nil {
					return err
				}
			case FieldTypeNumeric:
				if err := field.NumericField.Validate(); err != nil {
					return err
				}
			case FieldTypeGeo:
				if err := field.GeoField.Validate(); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (c CollectionConfig) CreateCollectionQuery() []interface{} {
	collectionQuery := []interface{}{
		"FT.CREATE", c.Name,
		"ON", string(c.On),
		"PREFIX", 1, createIndexPrefix(c.Name),
		"SCHEMA",
	}

	for _, field := range c.Fields {
		switch field.Type {
		case FieldTypeVector:
			collectionQuery = append(collectionQuery, field.VectorField.CreateVectorFieldQuery(c.On)...)
		case FieldTypeTag:
			collectionQuery = append(collectionQuery, field.TagField.CreateTagFieldQuery(c.On)...)
		case FieldTypeText:
			collectionQuery = append(collectionQuery, field.TextField.CreateTextFieldQuery(c.On)...)
		case FieldTypeNumeric:
			collectionQuery = append(collectionQuery, field.NumericField.CreateNumericFieldQuery(c.On)...)
		case FieldTypeGeo:
			collectionQuery = append(collectionQuery, field.GeoField.CreateGeoFieldQuery(c.On)...)
		}
	}

	return collectionQuery
}

type FieldConfig struct {
	Type         FieldType
	VectorField  *VectorFieldConfig
	TagField     *TagFieldConfig
	TextField    *TextFieldConfig
	NumericField *NumericFieldConfig
	GeoField     *GeoFieldConfig
}

type VectorFieldConfig struct {
	Name       string
	Dimension  int
	Algorithm  IndexAlgorithm
	Metric     DistanceMetric
	Type       VectorType
	InitialCap int
	HNSW       *HNSWConfig
}

type HNSWConfig struct {
	M              int
	EFConstruction int
	EFRuntime      int
}

func (v *VectorFieldConfig) Validate() error {
	if v.Dimension < 0 {
		return errors.New("Dimension cannot be less than 0")
	}

	if _, isExists := RegisteredIndexAlgorithm[v.Algorithm]; !isExists {
		return fmt.Errorf("unsuported Algorithm: %s", v.Algorithm)
	}

	if _, isExists := RegisteredDistanceMetric[v.Metric]; !isExists {
		return fmt.Errorf("unsuported Metric: %s", v.Metric)
	}

	if _, isExists := RegisteredVectorType[v.Type]; !isExists {
		return fmt.Errorf("unsuported vector type: %s", v.Type)
	}

	if v.Algorithm == HNSW && v.HNSW == nil {
		return errors.New("HNSW config is empty when attempting to create HNSW index")
	}

	if v.HNSW != nil {
		if v.HNSW.M < M_SMALL || v.HNSW.M > M_JUMBO {
			return fmt.Errorf("unsuported graph connection value, need to be between %d and %d: %d", M_SMALL, M_JUMBO, v.HNSW.M)
		}

		if v.HNSW.EFConstruction < EF_CONSTRUCTION_SMALL || v.HNSW.EFConstruction > EF_CONSTRUCTION_JUMBO {
			return fmt.Errorf("unsuported candidate size, need to be between %d and %d: %d", EF_CONSTRUCTION_SMALL, EF_CONSTRUCTION_JUMBO, v.HNSW.EFConstruction)
		}

		if v.HNSW.EFRuntime < EF_RUNTIME_SMALL || v.HNSW.EFRuntime > EF_RUNTIME_JUMBO {
			return fmt.Errorf("unsuported candidate size, need to be between %d and %d: %d", EF_RUNTIME_SMALL, EF_RUNTIME_JUMBO, v.HNSW.EFRuntime)
		}
	}

	return nil
}

func (v *VectorFieldConfig) CreateVectorFieldQuery(indexType IndexType) []interface{} {
	var vectorFieldQuery []interface{}

	switch indexType {
	case HASH_TYPE:
		vectorFieldQuery = append(vectorFieldQuery,
			v.Name,
			string(FieldTypeVector),
			string(v.Algorithm),
		)
	case JSON_TYPE:
		vectorFieldQuery = append(vectorFieldQuery,
			"$."+v.Name, "AS", v.Name,
			string(FieldTypeVector),
			string(v.Algorithm),
		)

	}

	switch v.Algorithm {
	case HNSW:
		vectorFieldQuery = append(vectorFieldQuery, 12)
	case FLAT:
		vectorFieldQuery = append(vectorFieldQuery, 8)
	}

	vectorFieldQuery = append(vectorFieldQuery,
		"TYPE", string(v.Type),
		"DIM", v.Dimension,
		"DISTANCE_METRIC", string(v.Metric),
		"INITIAL_CAP", v.InitialCap,
	)

	if v.Algorithm == HNSW {
		vectorFieldQuery = append(vectorFieldQuery,
			"M", v.HNSW.M,
			"EF_CONSTRUCTION", v.HNSW.EFConstruction,
		)
	}

	return vectorFieldQuery
}

type TagFieldConfig struct {
	Name      string
	JSONName  string
	Separator string
	Sortable  bool
	NoIndex   bool
}

func (t *TagFieldConfig) Validate() error {
	if t.Name == "" {
		return errors.New("Name cannot be empty")
	}
	return nil
}

func (t *TagFieldConfig) CreateTagFieldQuery(indexType IndexType) []interface{} {
	var tagFieldQuery []interface{}

	switch indexType {
	case HASH_TYPE:
		tagFieldQuery = append(tagFieldQuery, t.Name, string(FieldTypeTag))
	case JSON_TYPE:
		tagFieldQuery = append(tagFieldQuery, "$.metadata."+t.Name, "AS", t.Name, string(FieldTypeTag))
	}

	if t.Separator != "" {
		tagFieldQuery = append(tagFieldQuery, "SEPARATOR", t.Separator)
	}

	if t.Sortable {
		tagFieldQuery = append(tagFieldQuery, "SORTABLE")
	}

	if t.NoIndex {
		tagFieldQuery = append(tagFieldQuery, "NOINDEX")
	}

	return tagFieldQuery
}

type TextFieldConfig struct {
	Name     string
	Weight   float64
	Sortable bool
	NoIndex  bool
	NoStem   bool
	Phonetic string
}

func (t *TextFieldConfig) Validate() error {
	if t.Name == "" {
		return errors.New("Name cannot be empty")
	}

	if math.IsNaN(t.Weight) || math.IsInf(t.Weight, 0) {
		return errors.New("Weight must be a finite number")
	}

	if t.Weight < 0 {
		return errors.New("Weight cannot be less than 0")
	}

	return nil
}

func (t *TextFieldConfig) CreateTextFieldQuery(indexType IndexType) []interface{} {
	var textFieldQuery []interface{}

	switch indexType {
	case HASH_TYPE:
		textFieldQuery = append(textFieldQuery, t.Name, string(FieldTypeText))
	case JSON_TYPE:
		textFieldQuery = append(textFieldQuery, "$.metadata."+t.Name, "AS", t.Name, string(FieldTypeText))
	}

	if t.Weight != 0 {
		textFieldQuery = append(textFieldQuery, "WEIGHT", t.Weight)
	}

	if t.Sortable {
		textFieldQuery = append(textFieldQuery, "SORTABLE")
	}

	if t.NoIndex {
		textFieldQuery = append(textFieldQuery, "NOINDEX")
	}

	if t.NoStem {
		textFieldQuery = append(textFieldQuery, "NOSTEM")
	}

	if t.Phonetic != "" {
		textFieldQuery = append(textFieldQuery, "PHONETIC", t.Phonetic)
	}

	return textFieldQuery
}

type NumericFieldConfig struct {
	Name     string
	Sortable bool
	NoIndex  bool
}

func (t *NumericFieldConfig) Validate() error {
	if t.Name == "" {
		return errors.New("Name cannot be empty")
	}
	return nil
}

func (t *NumericFieldConfig) CreateNumericFieldQuery(indexType IndexType) []interface{} {
	var numericFieldQuery []interface{}

	switch indexType {
	case HASH_TYPE:
		numericFieldQuery = append(numericFieldQuery, t.Name, string(FieldTypeNumeric))
	case JSON_TYPE:
		numericFieldQuery = append(numericFieldQuery, "$.metadata."+t.Name, "AS", t.Name, string(FieldTypeNumeric))
	}

	if t.Sortable {
		numericFieldQuery = append(numericFieldQuery, "SORTABLE")
	}

	if t.NoIndex {
		numericFieldQuery = append(numericFieldQuery, "NOINDEX")
	}

	return numericFieldQuery
}

type GeoFieldConfig struct {
	Name     string
	Sortable bool
	NoIndex  bool
}

func (t *GeoFieldConfig) Validate() error {
	if t.Name == "" {
		return errors.New("Name cannot be empty")
	}
	return nil
}

func (t *GeoFieldConfig) CreateGeoFieldQuery(indexType IndexType) []interface{} {
	var geoFieldQuery []interface{}

	switch indexType {
	case HASH_TYPE:
		geoFieldQuery = append(geoFieldQuery, t.Name, string(FieldTypeGeo))
	case JSON_TYPE:
		geoFieldQuery = append(geoFieldQuery, "$.metadata."+t.Name, "AS", t.Name, string(FieldTypeGeo))
	}

	if t.Sortable {
		geoFieldQuery = append(geoFieldQuery, "SORTABLE")
	}

	if t.NoIndex {
		geoFieldQuery = append(geoFieldQuery, "NOINDEX")
	}

	return geoFieldQuery
}

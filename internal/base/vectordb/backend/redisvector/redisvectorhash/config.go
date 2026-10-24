package redisvectorhash

import (
	"errors"
	"fmt"
	"math"
)

type VectorDBConnectorConfig struct {
	Address  string
	Username string
	Password string
	Db       int
}

func (v *VectorDBConnectorConfig) Validate() error {
	if v.Address == "" {
		return errors.New("Address cannot be empty")
	}
	if v.Db < 0 {
		return errors.New("Db number cannot be less than 0")
	}
	return nil
}

type IndexType string

const (
	HNSW IndexType = "HNSW"
	FLAT IndexType = "FLAT"
)

var RegisteredIndexType = map[IndexType]struct{}{
	HNSW: {},
	FLAT: {},
}

type DistanceType string

const (
	COSINE DistanceType = "COSINE"
	L2     DistanceType = "L2"
	IP     DistanceType = "IP"
)

var RegisteredDistanceType = map[DistanceType]struct{}{
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
	HNSW_M_SMALL  = 16
	HNSW_M_MEDIUM = 24
	HNSW_M_LARGE  = 32
	HNSW_M_JUMBO  = 40
)

const (
	HNSW_EF_CONSTRUCTION_SMALL  = 100
	HNSW_EF_CONSTRUCTION_MEDIUM = 200
	HNSW_EF_CONSTRUCTION_LARGE  = 300
	HNSW_EF_CONSTRUCTION_JUMBO  = 400
)

const (
	HNSW_EF_RUNTIME_SMALL  = 10
	HNSW_EF_RUNTIME_MEDIUM = 20
	HNSW_EF_RUNTIME_LARGE  = 50
	HNSW_EF_RUNTIME_JUMBO  = 100
)

type FieldType string

const (
	FieldTypeVector  FieldType = "VECTOR"
	FieldTypeTag     FieldType = "TAG"
	FieldTypeText    FieldType = "TEXT"
	FieldTypeNumeric FieldType = "NUMERIC"
	FieldTypeGeo     FieldType = "GEO"
)

type CollectionConfig struct {
	Name            string
	EmbeddingFields []*EmbeddingFieldConfig
	DataFields      []*DataFieldConfig
}

func (c *CollectionConfig) Validate() error {

	if c.Name == "" {
		return errors.New("collection name cannot be empty")
	}

	if c.EmbeddingFields == nil {
		return errors.New("embedding field need to have configuration")
	}

	for _, embeddingField := range c.EmbeddingFields {
		if err := embeddingField.Validate(); err != nil {
			return err
		}
	}

	if len(c.DataFields) == 0 {
		return errors.New("collection fields cannot be empty")
	}

	if c.DataFields != nil {
		for _, field := range c.DataFields {
			switch field.Type {
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
		"ON", "HASH",
		"PREFIX", 1, createIndexPrefix(c.Name),
		"SCHEMA",
	}

	for _, field := range c.EmbeddingFields {
		collectionQuery = append(collectionQuery, field.CreateVectorFieldQuery()...)
	}

	for _, field := range c.DataFields {
		switch field.Type {
		case FieldTypeTag:
			collectionQuery = append(collectionQuery, field.TagField.CreateTagFieldQuery()...)
		case FieldTypeText:
			collectionQuery = append(collectionQuery, field.TextField.CreateTextFieldQuery()...)
		case FieldTypeNumeric:
			collectionQuery = append(collectionQuery, field.NumericField.CreateNumericFieldQuery()...)
		case FieldTypeGeo:
			collectionQuery = append(collectionQuery, field.GeoField.CreateGeoFieldQuery()...)
		}
	}

	return collectionQuery
}

type EmbeddingFieldConfig struct {
	Name       string
	DataType   VectorType
	Dimension  int
	Distance   DistanceType
	Index      IndexType
	InitialCap int
	HNSW       *HNSWConfig
}

func (v *EmbeddingFieldConfig) Validate() error {
	if v.Dimension < 0 {
		return errors.New("dimension cannot be less than 0")
	}

	if _, isExists := RegisteredIndexType[v.Index]; !isExists {
		return fmt.Errorf("unsuported Algorithm: %s", v.Index)
	}

	if _, isExists := RegisteredDistanceType[v.Distance]; !isExists {
		return fmt.Errorf("unsuported Metric: %s", v.Distance)
	}

	if _, isExists := RegisteredVectorType[v.DataType]; !isExists {
		return fmt.Errorf("unsuported vector type: %s", v.DataType)
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

func (c *HNSWConfig) Validate() error {

	if c == nil {
		return errors.New("HNSW config is empty when attempting to create HNSW index")
	}

	if c.M < HNSW_M_SMALL || c.M > HNSW_M_JUMBO {
		return fmt.Errorf("unsuported graph connection value, need to be between %d and %d: %d",
			HNSW_M_SMALL,
			HNSW_M_JUMBO,
			c.M)
	}

	if c.EFConstruction < HNSW_EF_CONSTRUCTION_SMALL || c.EFConstruction > HNSW_EF_CONSTRUCTION_JUMBO {
		return fmt.Errorf("unsuported ef construction size, need to be between %d and %d: %d",
			HNSW_EF_CONSTRUCTION_SMALL,
			HNSW_EF_CONSTRUCTION_JUMBO,
			c.EFConstruction)
	}

	if c.EFRuntime < HNSW_EF_RUNTIME_SMALL || c.EFRuntime > HNSW_EF_RUNTIME_JUMBO {
		return fmt.Errorf("unsuported ef runtime size, need to be between %d and %d: %d",
			HNSW_EF_RUNTIME_SMALL,
			HNSW_EF_RUNTIME_JUMBO,
			c.EFRuntime)
	}

	return nil
}

func (v *EmbeddingFieldConfig) CreateVectorFieldQuery() []interface{} {
	var vectorFieldQuery []interface{}

	vectorFieldQuery = append(vectorFieldQuery,
		v.Name,
		string(FieldTypeVector),
		string(v.Index),
	)

	switch v.Index {
	case HNSW:
		vectorFieldQuery = append(vectorFieldQuery, 12)
	case FLAT:
		vectorFieldQuery = append(vectorFieldQuery, 8)
	}

	vectorFieldQuery = append(vectorFieldQuery,
		"TYPE", string(v.DataType),
		"DIM", v.Dimension,
		"DISTANCE_METRIC", string(v.Distance),
		"INITIAL_CAP", v.InitialCap,
	)

	if v.Index == HNSW {
		vectorFieldQuery = append(vectorFieldQuery,
			"M", v.HNSW.M,
			"EF_CONSTRUCTION", v.HNSW.EFConstruction,
		)
	}

	return vectorFieldQuery
}

type DataFieldConfig struct {
	Type         FieldType
	TagField     *TagFieldConfig
	TextField    *TextFieldConfig
	NumericField *NumericFieldConfig
	GeoField     *GeoFieldConfig
}

type TagFieldConfig struct {
	Name      string
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

func (t *TagFieldConfig) CreateTagFieldQuery() []interface{} {
	var tagFieldQuery []interface{}

	tagFieldQuery = append(tagFieldQuery, t.Name, string(FieldTypeTag))

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

func (t *TextFieldConfig) CreateTextFieldQuery() []interface{} {
	var textFieldQuery []interface{}

	textFieldQuery = append(textFieldQuery, t.Name, string(FieldTypeText))

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

func (t *NumericFieldConfig) CreateNumericFieldQuery() []interface{} {
	var numericFieldQuery []interface{}

	numericFieldQuery = append(numericFieldQuery, t.Name, string(FieldTypeNumeric))

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

func (t *GeoFieldConfig) CreateGeoFieldQuery() []interface{} {
	var geoFieldQuery []interface{}

	geoFieldQuery = append(geoFieldQuery, t.Name, string(FieldTypeGeo))

	if t.Sortable {
		geoFieldQuery = append(geoFieldQuery, "SORTABLE")
	}

	if t.NoIndex {
		geoFieldQuery = append(geoFieldQuery, "NOINDEX")
	}

	return geoFieldQuery
}

package database

import (
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type DAO struct {
	Schema       *schema.Schema
	ReflectValue reflect.Value
}

func parseDAO(db *gorm.DB, structPointer interface{}) (*DAO, error) {

	statement := &gorm.Statement{
		DB: db,
	}

	if err := statement.Parse(structPointer); err != nil {
		return nil, err
	}

	reflectValue := reflect.ValueOf(structPointer)
	for reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}

	return &DAO{
		Schema:       statement.Schema,
		ReflectValue: reflectValue,
	}, nil
}

func (d *DAO) PrimaryKeysColumnName() ([]string, error) {

	primaryKeys := make([]string, 0, len(d.Schema.PrimaryFields))
	for _, field := range d.Schema.PrimaryFields {
		primaryKeys = append(primaryKeys, field.DBName)
	}

	return primaryKeys, nil
}

func (d *DAO) PrimaryKeysValues() (map[string]interface{}, error) {

	values := make(map[string]interface{}, len(d.Schema.PrimaryFields))

	for _, field := range d.Schema.PrimaryFields {
		value := d.ReflectValue.FieldByIndex(field.StructField.Index).Interface()
		values[field.DBName] = value
	}

	return values, nil
}

func (d *DAO) GetSelectedColumnName(structFields ...string) ([]string, error) {

	dbColumns := make([]string, 0, len(structFields))
	for _, field := range structFields {
		columnDB := d.Schema.LookUpField(field)
		if columnDB == nil {
			return nil, fmt.Errorf("field %q does not exist in model %s", field, d.Schema.Name)
		}

		dbColumns = append(dbColumns, columnDB.DBName)
	}

	return dbColumns, nil
}

func (d *DAO) GetUpdateableColumnsName() ([]string, error) {

	const EMPTYCOLUMNNAME = ""

	updateableColumns := make([]string, 0)
	for _, field := range d.Schema.Fields {
		if field.PrimaryKey {
			continue
		}

		if _, ok := field.TagSettings["AUTOCREATETIME"]; ok {
			continue
		}

		if strings.TrimSpace(field.DBName) == EMPTYCOLUMNNAME {
			continue
		}

		updateableColumns = append(updateableColumns, field.DBName)
	}

	return updateableColumns, nil
}

func formatConflictColumns(primaryKeys []string) []clause.Column {
	columns := make([]clause.Column, 0, len(primaryKeys))

	for _, key := range primaryKeys {
		columns = append(columns, clause.Column{
			Name: key,
		})
	}

	return columns
}

func touchTimestamp(structPointer interface{}, fieldName string, unixTime *int64) error {
	if structPointer == nil {
		return fmt.Errorf("structPointer is nil")
	}

	v := reflect.ValueOf(structPointer)

	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("structPointer must be a pointer")
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("structPointer must point to a struct")
	}

	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return fmt.Errorf("field %s does not exist", fieldName)
	}

	if !field.CanSet() {
		return fmt.Errorf("field %s cannot be set", fieldName)
	}

	switch field.Kind() {
	case reflect.Int64:
		field.SetInt(*unixTime)
	case reflect.Ptr:
		if field.Type().Elem().Kind() != reflect.Int64 {
			return fmt.Errorf("field %s must be *int64", fieldName)
		}
		if unixTime == nil {
			field.Set(reflect.Zero(field.Type()))
		} else {
			field.Set(reflect.ValueOf(unixTime))
		}

	default:
		return fmt.Errorf("field %s must be int64 or *int64", fieldName)
	}

	return nil
}

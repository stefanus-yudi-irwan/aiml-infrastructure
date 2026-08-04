package database

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
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

func (d *DAO) TableName() string {
	return d.Schema.Table
}

func (d *DAO) ColumnNames() []string {
	columns := make([]string, 0, len(d.Schema.DBNames))
	columns = append(columns, d.Schema.DBNames...)
	return columns
}

func (d *DAO) ColumnValues() []interface{} {
	values := make([]interface{}, 0, len(d.Schema.Fields))

	for _, field := range d.Schema.Fields {
		value, _ := field.ValueOf(context.Background(), d.ReflectValue)
		values = append(values, value)
	}

	return values
}

func (d *DAO) PrimaryKeysColumnName() []string {

	primaryKeys := make([]string, 0, len(d.Schema.PrimaryFields))
	for _, field := range d.Schema.PrimaryFields {
		primaryKeys = append(primaryKeys, field.DBName)
	}

	return primaryKeys
}

func (d *DAO) PrimaryKeysValues() map[string]interface{} {

	values := make(map[string]interface{}, len(d.Schema.PrimaryFields))

	for _, field := range d.Schema.PrimaryFields {
		value := d.ReflectValue.FieldByIndex(field.StructField.Index).Interface()
		values[field.DBName] = value
	}

	return values
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

func (d *DAO) GetUpdateableColumnsName() []string {

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

	return updateableColumns
}

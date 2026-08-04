package database

import (
	"fmt"
	"reflect"

	"gorm.io/gorm/clause"
)

func FormatConflictColumns(primaryKeys []string) []clause.Column {
	columns := make([]clause.Column, 0, len(primaryKeys))

	for _, key := range primaryKeys {
		columns = append(columns, clause.Column{
			Name: key,
		})
	}

	return columns
}

func TouchTimestamp(structPointer interface{}, fieldName string, unixTime *int64) error {
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

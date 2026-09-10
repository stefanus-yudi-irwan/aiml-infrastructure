package database

import (
	"context"
	"errors"
	"reflect"
)

type IDBConnector interface {
	Insert(ctx context.Context, structPointer interface{}) error
	Update(ctx context.Context, structPointer interface{}, structFields ...string) error
	Upsert(ctx context.Context, structPointer interface{}) error
	HardDelete(ctx context.Context, structPointer interface{}) error
	SoftDelete(ctx context.Context, structPointer interface{}) error
	Restore(ctx context.Context, structPointer interface{}) error
	GetByPrimaryKeys(ctx context.Context, structPointer interface{}) error
	Exists(ctx context.Context, structPointer interface{}) (bool, error)
	ExecuteSQL(ctx context.Context, sql string) error
	Ping(ctx context.Context, timeLimitSecond int) error
	Close() error
}

func ValidateStructPointer(value interface{}) error {
	if value == nil {
		return errors.New("struct pointer cannot be nil")
	}

	v := reflect.ValueOf(value)

	if v.Kind() != reflect.Ptr {
		return errors.New("struct pointer must be a pointer")
	}

	if v.IsNil() {
		return errors.New("struct pointer cannot be nil")
	}

	if v.Elem().Kind() != reflect.Struct {
		return errors.New("struct pointer must be a pointer")
	}

	return nil
}

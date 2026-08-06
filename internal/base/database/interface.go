package database

import "context"

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

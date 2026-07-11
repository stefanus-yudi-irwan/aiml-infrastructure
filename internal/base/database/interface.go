package database

type DataStruct interface{}
type IsDataExists bool

type IDBConnector interface {
	Insert(structPointer DataStruct) error
	Update(structPointer DataStruct, structFields ...string) error
	Upsert(structPointer DataStruct) error
	HardDelete(structPointer DataStruct) error
	SoftDelete(structPointer DataStruct) error
	Restore(structPointer DataStruct) error
	GetByPrimaryKeys(structPointer DataStruct) error
	Exists(structPointer DataStruct) (IsDataExists, error)
	Ping(timeLimitSecond int) error
	Close() error
}

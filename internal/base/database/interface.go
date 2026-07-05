package database

type IDBConnector interface {
	Insert(structPointer interface{}) error
	Update(structPointer interface{}, structFields ...string) error
	Upsert(structPointer interface{}) error
	HardDelete(structPointer interface{}) error
	SoftDelete(structPointer interface{}) error
	Restore(structPointer interface{}) error
}

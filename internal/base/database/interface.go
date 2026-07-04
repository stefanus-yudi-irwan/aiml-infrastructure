package database

type IDBConnector interface {
	InsertData(structData interface{}) error
	DeleteData(structData interface{}) error
	UpdateData(structData interface{}, structFields ...string) error
	UpsertData(structData interface{}) error
	GetData(structID interface{}) error
}

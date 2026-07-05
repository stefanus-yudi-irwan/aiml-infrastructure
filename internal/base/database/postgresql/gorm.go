package postgresql

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

var gormConfig = &gorm.Config{
	CreateBatchSize:        1000,
	SkipDefaultTransaction: true,
	NamingStrategy: schema.NamingStrategy{
		SingularTable: true,
		TablePrefix:   "",
	},
}

type postgresDBConnector struct {
	GormDB   *gorm.DB
	ClientDB *sql.DB
}

func NewPostgresDBConnector(connectionPath string, numberOfConnections int) (*postgresDBConnector, error) {
	var err error

	GormDB, err := gorm.Open(postgres.Open(connectionPath), gormConfig)
	if err != nil {
		return nil, err
	}

	ClientDB, err := GormDB.DB()
	if err != nil {
		return nil, err
	}

	ClientDB.SetMaxIdleConns(numberOfConnections)
	ClientDB.SetMaxOpenConns(numberOfConnections)
	ClientDB.SetConnMaxLifetime(time.Hour)
	ClientDB.SetConnMaxIdleTime(10 * time.Minute)

	return &postgresDBConnector{
		GormDB:   GormDB,
		ClientDB: ClientDB,
	}, nil
}

func (p *postgresDBConnector) error(err error, method string, params ...interface{}) error {
	return fmt.Errorf("postgresDBConnector.(%v)(%v) %w", method, params, err)
}

func (p *postgresDBConnector) InsertData(structPointer interface{}) error {
	if err := p.GormDB.Create(structPointer).Error; err != nil {
		primaryKeys, err := getPrimaryKeys(p.GormDB, structPointer)
		if err != nil {
			return p.error(err, "InsertData-001", "fail to get primary keys")
		}
		return p.error(err, "InsertData-002", primaryKeys)
	}
	return nil
}

func (p *postgresDBConnector) UpdateData(structPointer interface{}, structFields ...string) error {

	dbColumnName, err := getColumnName(p.GormDB, structPointer, structFields...)
	if err != nil {
		primaryKeys, err := getPrimaryKeys(p.GormDB, structPointer)
		if err != nil {
			return p.error(err, "UpdateData-001", "fail to get primary keys")
		}
		return p.error(err, "UpdateData-002", primaryKeys)
	}

	tx := p.GormDB.Select(dbColumnName).Updates(structPointer)

	if tx.Error != nil {
		primaryKeys, err := getPrimaryKeys(p.GormDB, structPointer)
		if err != nil {
			return p.error(err, "UpdateData-003", "fail to get primary keys")
		}
		return p.error(tx.Error, "UpdateData-004", primaryKeys)
	}

	if tx.RowsAffected == 0 {
		primaryKeys, err := getPrimaryKeys(p.GormDB, structPointer)
		if err != nil {
			return p.error(err, "UpdateData-005", "fail to get primary keys")
		}
		return p.error(gorm.ErrRecordNotFound, "UpdateData-006", primaryKeys)
	}

	return nil
}

func (p *postgresDBConnector) UpsertData(structPointer interface{}) error {

	if err := touchUpdatedAt(structPointer); err != nil {
		return p.error(err, "UpsertData-001", "failed to update UpdatedAt")
	}

	primaryKeys, err := getPrimaryKeys(p.GormDB, structPointer)
	if err != nil {
		return p.error(err, "UpsertData-002", "fail to get primary keys")
	}

	updateableColumns, err := getUpdateableColumns(p.GormDB, structPointer)
	if err != nil {
		return p.error(err, "UpsertData-003", primaryKeys)
	}

	if err = p.GormDB.Clauses(clause.OnConflict{
		Columns:   formatConflictColumns(primaryKeys),
		DoUpdates: clause.AssignmentColumns(updateableColumns),
	}).Create(structPointer).Error; err != nil {
		return p.error(err, "UpsertData-004", primaryKeys)
	}

	return nil
}

func (p *postgresDBConnector) DeleteData(structPointer interface{}) error {
	err := p.GormDB.Delete(structPointer).Error
	if err != nil {
		primaryKeys, err := getPrimaryKeys(p.GormDB, structPointer)
		if err != nil {
			return p.error(err, "DeleteData-001", "fail to get primary keys")
		}
		return p.error(err, "DeleteData-002", primaryKeys)
	}

	return nil
}

// func (p *postgresDBConnector) GetDataByID(id interface{}, data interface{}) error {
// 	return p.GormDB.First(data, id).Error
// }

func getSchema(db *gorm.DB, structPointer interface{}) (*schema.Schema, error) {
	statement := &gorm.Statement{
		DB: db,
	}

	if err := statement.Parse(structPointer); err != nil {
		return nil, err
	}

	return statement.Schema, nil
}

func getColumnName(db *gorm.DB, structPointer interface{}, structFields ...string) ([]string, error) {
	dataSchema, err := getSchema(db, structPointer)
	if err != nil {
		return nil, err
	}

	dbColumns := make([]string, 0, len(structFields))
	for _, field := range structFields {
		columnDB := dataSchema.LookUpField(field)
		if columnDB == nil {
			return nil, fmt.Errorf("field %q does not exist in model %s", field, dataSchema.Name)
		}

		dbColumns = append(dbColumns, columnDB.DBName)
	}

	return dbColumns, nil
}

func getPrimaryKeys(db *gorm.DB, structPointer interface{}) ([]string, error) {
	schema, err := getSchema(db, structPointer)
	if err != nil {
		return nil, err
	}

	primaryKeys := make([]string, 0, len(schema.PrimaryFields))
	for _, field := range schema.PrimaryFields {
		primaryKeys = append(primaryKeys, field.DBName)
	}

	return primaryKeys, nil
}

func getUpdateableColumns(db *gorm.DB, structPointer interface{}) ([]string, error) {

	const EMPTYCOLUMNNAME = ""

	schema, err := getSchema(db, structPointer)
	if err != nil {
		return nil, err
	}

	updateableColumns := make([]string, 0)
	for _, field := range schema.Fields {
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

func touchUpdatedAt(structPointer interface{}) error {
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

	field := v.FieldByName("UpdatedAt")
	if !field.IsValid() {
		return fmt.Errorf("field UpdatedAt does not exist")
	}

	if !field.CanSet() {
		return fmt.Errorf("field UpdatedAt cannot be set")
	}

	if field.Kind() != reflect.Int64 {
		return fmt.Errorf("field UpdatedAt must be int64")
	}

	field.SetInt(time.Now().Unix())

	return nil
}

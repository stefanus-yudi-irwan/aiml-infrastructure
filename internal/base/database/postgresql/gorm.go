package postgresql

import (
	"database/sql"
	"fmt"
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

func (p *postgresDBConnector) InsertData(structData interface{}) error {
	if err := p.GormDB.Create(structData).Error; err != nil {
		primaryKeys, err := getPrimaryKeys(p.GormDB, structData)
		if err != nil {
			return p.error(err, "InsertData", "fail to get primary key")
		}
		return p.error(err, "InsertData", primaryKeys)
	}
	return nil
}

func (p *postgresDBConnector) UpdateData(structData interface{}, structFields ...string) error {

	dbColumnName, err := getColumnName(p.GormDB, structData, structFields...)
	if err != nil {
		return p.error(err, "UpdateData", "fail to get column name")
	}

	tx := p.GormDB.Select(dbColumnName).Updates(structData)

	if tx.Error != nil {
		return p.error(tx.Error, "UpdateData")
	}
	if tx.RowsAffected == 0 {
		return p.error(gorm.ErrRecordNotFound, "UpdateData")
	}

	return nil
}

func (p *postgresDBConnector) UpsertData(structData interface{}) error {
	return p.GormDB.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(structData).Error
}

// func (p *postgresDBConnector) DeleteData(structData interface{}) error {
// 	return p.GormDB.Delete(structData).Error
// }

// func (p *postgresDBConnector) GetDataByID(id interface{}, data interface{}) error {
// 	return p.GormDB.First(data, id).Error
// }

func getSchema(db *gorm.DB, structData interface{}) (*schema.Schema, error) {
	statement := &gorm.Statement{
		DB: db,
	}

	if err := statement.Parse(structData); err != nil {
		return nil, err
	}

	return statement.Schema, nil
}

func getColumnName(db *gorm.DB, structData interface{}, structFields ...string) ([]string, error) {
	dataSchema, err := getSchema(db, structData)
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

func getPrimaryKeys(db *gorm.DB, structData interface{}) ([]string, error) {
	schema, err := getSchema(db, structData)
	if err != nil {
		return nil, err
	}

	primaryKeys := make([]string, 0, len(schema.PrimaryFields))
	for _, field := range schema.PrimaryFields {
		primaryKeys = append(primaryKeys, field.DBName)
	}

	return primaryKeys, nil
}

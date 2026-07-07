package postgresql

import (
	"context"
	"database/sql"
	"errors"
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

func (p *postgresDBConnector) Insert(structPointer interface{}) error {

	if err := p.GormDB.Create(structPointer).Error; err != nil {
		DAOMetadata, err := parseDAO(p.GormDB, structPointer)
		if err != nil {
			return p.error(err, "Insert-001", "failed to parse DAO")
		}
		primaryKeysValue, err := DAOMetadata.PrimaryKeysValues()
		if err != nil {
			return p.error(err, "Insert-001", "fail to get primary keys")
		}
		return p.error(err, "Insert-002", primaryKeysValue)
	}
	return nil
}

func (p *postgresDBConnector) Update(structPointer interface{}, structFields ...string) error {

	DAOMetadata, err := parseDAO(p.GormDB, structPointer)
	if err != nil {
		return p.error(err, "Update-001", "failed to parse DAO")
	}

	columns, err := DAOMetadata.GetSelectedColumnName(structFields...)
	if err != nil {
		return p.error(err, "Update-002", "failed to get column names")
	}
	tx := p.GormDB.
		Select(columns).
		Updates(structPointer)

	if tx.Error != nil {
		primaryKeysValue, err := DAOMetadata.PrimaryKeysValues()
		if err != nil {
			return p.error(err, "Update-002", "fail to get primary keys")
		}
		return p.error(tx.Error, "Update-003", primaryKeysValue)
	}

	if tx.RowsAffected == 0 {
		primaryKeysValue, err := DAOMetadata.PrimaryKeysValues()
		if err != nil {
			return p.error(err, "Update-004", "fail to get primary keys")
		}
		return p.error(gorm.ErrRecordNotFound, "Update-005", primaryKeysValue)
	}

	return nil
}

func (p *postgresDBConnector) Upsert(structPointer interface{}) error {

	DAOMetadata, err := parseDAO(p.GormDB, structPointer)
	if err != nil {
		return p.error(err, "Upsert-001", "failed to parse DAO")
	}

	timeNow := time.Now().Unix()
	if err := touchTimestamp(structPointer, "UpdatedAt", &timeNow); err != nil {
		return p.error(err, "Upsert-002", "failed to update UpdatedAt")
	}

	primaryKeys, err := DAOMetadata.PrimaryKeysColumnName()
	if err != nil {
		return p.error(err, "Upsert-003", "fail to get primary keys")
	}

	updateableColumns, err := DAOMetadata.GetUpdateableColumnsName()
	if err != nil {
		return p.error(err, "Upsert-004", primaryKeys)
	}

	if err = p.GormDB.Clauses(clause.OnConflict{
		Columns:   formatConflictColumns(primaryKeys),
		DoUpdates: clause.AssignmentColumns(updateableColumns),
	}).Create(structPointer).Error; err != nil {
		return p.error(err, "Upsert-004", primaryKeys)
	}

	return nil
}

func (p *postgresDBConnector) HardDelete(structPointer interface{}) error {

	tx := p.GormDB.Delete(structPointer)

	if tx.Error != nil {
		DAOMetadata, err := parseDAO(p.GormDB, structPointer)
		if err != nil {
			return p.error(err, "HardDelete-001", "failed to parse DAO")
		}
		primaryKeysValues, err := DAOMetadata.PrimaryKeysValues()
		if err != nil {
			return p.error(err, "HardDelete-002", "fail to get primary keys")
		}
		return p.error(err, "HardDelete-003", primaryKeysValues)
	}

	if tx.RowsAffected == 0 {
		return p.error(gorm.ErrRecordNotFound, "HardDelete-004")
	}

	return nil
}

func (p *postgresDBConnector) SoftDelete(structPointer interface{}) error {

	timeNow := time.Now().Unix()
	if err := touchTimestamp(structPointer, "DeletedAt", &timeNow); err != nil {
		DAOMetadata, err := parseDAO(p.GormDB, structPointer)
		if err != nil {
			return p.error(err, "HardDelete-001", "failed to parse DAO")
		}
		primaryKeysValues, err := DAOMetadata.PrimaryKeysValues()
		if err != nil {
			return p.error(err, "SoftDelete-001", "fail to get primary keys")
		}
		return p.error(err, "SoftDelete-002", primaryKeysValues)
	}

	return p.Update(structPointer, "DeletedAt")
}

func (p *postgresDBConnector) Restore(structPointer interface{}) error {

	if err := touchTimestamp(structPointer, "DeletedAt", nil); err != nil {
		DAOMetadata, err := parseDAO(p.GormDB, structPointer)
		if err != nil {
			return p.error(err, "HardDelete-001", "failed to parse DAO")
		}
		primaryKeysValues, err := DAOMetadata.PrimaryKeysValues()
		if err != nil {
			return p.error(err, "Restore-001", "fail to get primary keys")
		}
		return p.error(err, "Restore-002", primaryKeysValues)
	}

	return p.Update(structPointer, "DeletedAt")

}

// func GetByPrimaryKey(structPointer interface{}) error {

// }

func (p *postgresDBConnector) Exists(structPointer interface{}) (bool, error) {
	tx := p.GormDB.Take(structPointer)

	if tx.Error == nil {
		return true, nil
	}

	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return false, nil
	}

	return false, p.error(tx.Error, "Exists-001")
}

func (p *postgresDBConnector) Ping(timeLimitSecond int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeLimitSecond)*time.Second)
	defer cancel()
	if err := p.ClientDB.PingContext(ctx); err != nil {
		return p.error(err, "Ping-001", fmt.Sprintf("failed to connect to database within %d second", timeLimitSecond))
	}
	return nil
}

func (p *postgresDBConnector) Close() error {
	if err := p.ClientDB.Close(); err != nil {
		return p.error(err, "Close-001", "failed to close database connection")
	}
	return nil
}

func (p *postgresDBConnector) GetByPrimaryKeys(structPointer interface{}) error {
	metadata, err := parseDAO(p.GormDB, structPointer)
	if err != nil {
		return p.error(err, "GetByPrimaryKeys-001", "failed to parse DAO")
	}

	primaryKeys, err := metadata.PrimaryKeysValues()
	if err != nil {
		return p.error(err, "GetByPrimaryKeys-002", "failed to get primary keys")
	}

	tx := p.GormDB.Where(primaryKeys).First(structPointer)

	if tx.Error != nil {
		return p.error(tx.Error, "GetByPrimaryKeys-003", primaryKeys)
	}

	return nil
}

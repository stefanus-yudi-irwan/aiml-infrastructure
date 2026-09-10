package backend

import (
	"aiml-infrastructure/internal/base/database"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

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

type DBConnector struct {
	GormDB   *gorm.DB
	ClientDB *sql.DB
}

func NewDBConnector(dialector gorm.Dialector, numberOfConnections int) (*DBConnector, error) {
	GormDB, err := gorm.Open(dialector, gormConfig)
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

	return &DBConnector{
		GormDB:   GormDB,
		ClientDB: ClientDB,
	}, nil
}

func (p *DBConnector) error(err error, method string, params ...interface{}) error {
	return fmt.Errorf("DBConnector.(%v)(%v) %w", method, params, err)
}

func (p *DBConnector) Insert(ctx context.Context, structPointer interface{}) error {
	if err := p.GormDB.WithContext(ctx).Create(structPointer).Error; err != nil {
		DAOMetadata, parseErr := parseDAO(p.GormDB, structPointer)
		if parseErr != nil {
			return p.error(parseErr, "Insert-001", "failed to parse DAO")
		}
		return p.error(err, "Insert-002", DAOMetadata.PrimaryKeysValues())
	}
	return nil
}

func (p *DBConnector) Update(ctx context.Context, structPointer interface{}, structFields ...string) error {
	DAOMetadata, err := parseDAO(p.GormDB, structPointer)
	if err != nil {
		return p.error(err, "Update-001", "failed to parse DAO")
	}

	columns, err := DAOMetadata.GetSelectedColumnName(structFields...)
	if err != nil {
		return p.error(err, "Update-002", "failed to get column names")
	}

	tx := p.GormDB.WithContext(ctx).Select(columns).Updates(structPointer)

	if tx.Error != nil {
		return p.error(tx.Error, "Update-003", DAOMetadata.PrimaryKeysValues())
	}

	if tx.RowsAffected == 0 {
		return p.error(gorm.ErrRecordNotFound, "Update-004", DAOMetadata.PrimaryKeysValues())
	}

	return nil
}

func (p *DBConnector) Upsert(ctx context.Context, structPointer interface{}) error {
	DAOMetadata, err := parseDAO(p.GormDB, structPointer)
	if err != nil {
		return p.error(err, "Upsert-001", "failed to parse DAO")
	}

	timeNow := time.Now().Unix()
	if err := database.TouchTimestamp(structPointer, "UpdatedAt", &timeNow); err != nil {
		return p.error(err, "Upsert-002", "failed to update UpdatedAt")
	}

	primaryKeys := DAOMetadata.PrimaryKeysColumnName()
	updateableColumns := DAOMetadata.GetUpdateableColumnsName()

	switch p.GormDB.Dialector.Name() {

	case "sqlserver": // due to sqlserver gorm library bug

		mergeQueryBuilder := SQLServerMergeBuilder{
			Table:         DAOMetadata.TableName(),
			Columns:       DAOMetadata.ColumnNames(),
			PrimaryKeys:   primaryKeys,
			UpdateColumns: updateableColumns,
		}

		mergeQuery := mergeQueryBuilder.Build()
		columnValues := DAOMetadata.ColumnValues()

		if err = p.GormDB.WithContext(ctx).Exec(mergeQuery, columnValues...).Error; err != nil {
			return p.error(err, "Upsert-003", primaryKeys)
		}

	default:
		if err = p.GormDB.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   database.FormatConflictColumns(primaryKeys),
			DoUpdates: clause.AssignmentColumns(updateableColumns),
		}).Create(structPointer).Error; err != nil {
			return p.error(err, "Upsert-004", primaryKeys)
		}
	}

	return nil
}

func (p *DBConnector) HardDelete(ctx context.Context, structPointer interface{}) error {
	tx := p.GormDB.WithContext(ctx).Delete(structPointer)

	if tx.Error != nil {
		DAOMetadata, err := parseDAO(p.GormDB, structPointer)
		if err != nil {
			return p.error(err, "HardDelete-001", "failed to parse DAO")
		}
		return p.error(err, "HardDelete-002", DAOMetadata.PrimaryKeysValues())
	}

	if tx.RowsAffected == 0 {
		return p.error(gorm.ErrRecordNotFound, "HardDelete-003")
	}

	return nil
}

func (p *DBConnector) SoftDelete(ctx context.Context, structPointer interface{}) error {
	timeNow := time.Now().Unix()
	if err := database.TouchTimestamp(structPointer, "DeletedAt", &timeNow); err != nil {
		DAOMetadata, err := parseDAO(p.GormDB, structPointer)
		if err != nil {
			return p.error(err, "SoftDelete-001", "failed to parse DAO")
		}
		return p.error(err, "SoftDelete-002", DAOMetadata.PrimaryKeysValues())
	}

	return p.Update(ctx, structPointer, "DeletedAt")
}

func (p *DBConnector) Restore(ctx context.Context, structPointer interface{}) error {
	if err := database.TouchTimestamp(structPointer, "DeletedAt", nil); err != nil {
		DAOMetadata, err := parseDAO(p.GormDB, structPointer)
		if err != nil {
			return p.error(err, "Restore-001", "failed to parse DAO")
		}
		return p.error(err, "Restore-002", DAOMetadata.PrimaryKeysValues())
	}

	return p.Update(ctx, structPointer, "DeletedAt")
}

func (p *DBConnector) Exists(ctx context.Context, structPointer interface{}) (bool, error) {
	tx := p.GormDB.WithContext(ctx).Take(structPointer)

	if tx.Error == nil {
		return true, nil
	}

	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return false, nil
	}

	return false, p.error(tx.Error, "Exists-001")
}

func (p *DBConnector) Ping(ctx context.Context, timeLimitSecond int) error {
	if timeLimitSecond <= 0 {
		return p.error(errors.New("timeLimit second cannot be less than or equal to 0"), "Ping-001")
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeLimitSecond)*time.Second)
	defer cancel()
	if err := p.ClientDB.PingContext(ctx); err != nil {
		return p.error(err, "Ping-002", fmt.Sprintf("failed to connect to database within %d second", timeLimitSecond))
	}
	return nil
}

func (p *DBConnector) Close() error {
	if err := p.ClientDB.Close(); err != nil {
		return p.error(err, "Close-001", "failed to close database connection")
	}
	return nil
}

func (p *DBConnector) GetByPrimaryKeys(ctx context.Context, structPointer interface{}) error {
	metadata, err := parseDAO(p.GormDB, structPointer)
	if err != nil {
		return p.error(err, "GetByPrimaryKeys-001", "failed to parse DAO")
	}

	primaryKeys := metadata.PrimaryKeysValues()
	tx := p.GormDB.WithContext(ctx).Where(primaryKeys).First(structPointer)

	if tx.Error != nil {
		return p.error(tx.Error, "GetByPrimaryKeys-002", primaryKeys)
	}

	return nil
}

func (p *DBConnector) ExecuteSQL(ctx context.Context, sql string) error {
	if err := p.GormDB.WithContext(ctx).Exec(sql).Error; err != nil {
		return p.error(err, "ExecuteSQL-001", "failed to execute SQL")
	}
	return nil
}

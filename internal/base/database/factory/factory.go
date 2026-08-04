package factory

import (
	"aiml-infrastructure/internal/base/database"
	"aiml-infrastructure/internal/base/database/backend"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
)

type BackendDB string

const (
	BackendMySQL      database.BackendDB = "mysql"
	BackendPostgreSQL database.BackendDB = "postgresql"
	BackendMSSQL      database.BackendDB = "mssql"
)

func NewDBConnector(config database.Config) (database.IDBConnector, error) {

	switch config.BackendDB {
	case BackendMySQL:
		return backend.NewDBConnector(mysql.Open(config.ConnectionPath), config.NumberOfConnections)
	case BackendPostgreSQL:
		return backend.NewDBConnector(postgres.Open(config.ConnectionPath), config.NumberOfConnections)
	case BackendMSSQL:
		return backend.NewDBConnector(sqlserver.Open(config.ConnectionPath), config.NumberOfConnections)
	default:
		return nil, fmt.Errorf("unsupported database backend: %s", config.BackendDB)
	}

}

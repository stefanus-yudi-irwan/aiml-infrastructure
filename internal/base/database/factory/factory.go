package factory

import (
	"aiml-infrastructure/internal/base/database"
	"aiml-infrastructure/internal/base/database/backend"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
)

func NewDBConnector(config database.Config) (database.IDBConnector, error) {
	switch config.Database {
	case database.BackendMySQL:
		return backend.NewDBConnector(mysql.Open(config.ConnectionPath), config.NumberOfConnections)
	case database.BackendPostgreSQL:
		return backend.NewDBConnector(postgres.Open(config.ConnectionPath), config.NumberOfConnections)
	case database.BackendMSSQL:
		return backend.NewDBConnector(sqlserver.Open(config.ConnectionPath), config.NumberOfConnections)
	default:
		return nil, fmt.Errorf("unsupported database backend: %s", config.Database)
	}
}

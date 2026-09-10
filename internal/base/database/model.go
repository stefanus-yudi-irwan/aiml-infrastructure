package database

import (
	"errors"
	"fmt"
)

type BackendDB string

const (
	BackendMySQL      BackendDB = "mysql"
	BackendPostgreSQL BackendDB = "postgresql"
	BackendMSSQL      BackendDB = "mssql"
)

var RegisteredBackend = map[BackendDB]struct{}{
	BackendMySQL:      {},
	BackendPostgreSQL: {},
	BackendMSSQL:      {},
}

type Config struct {
	Database            BackendDB
	ConnectionPath      string
	NumberOfConnections int
}

func (c *Config) Validate() error {
	if _, isExists := RegisteredBackend[c.Database]; !isExists {
		return fmt.Errorf("unsuported database: %s", c.Database)
	}

	if c.ConnectionPath == "" {
		return fmt.Errorf("database connection path cannot be empty")
	}

	if c.NumberOfConnections <= 0 {
		return errors.New("database number of connections must be greater than zero")
	}

	return nil
}

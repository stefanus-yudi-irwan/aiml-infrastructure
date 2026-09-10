package mssql

import (
	"aiml-infrastructure/internal/base/database"
	"aiml-infrastructure/internal/base/database/testsuite"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MSSQLTestSuite struct {
	testsuite.DBTestSuite
}

func (p *MSSQLTestSuite) SetupSuite() {
	err := godotenv.Load("init/.env")
	assert.NoError(p.T(), err)

	mssqlUser := os.Getenv("MSSQL_USER")
	mssqlPassword := os.Getenv("MSSQL_PASSWORD")
	mssqlHost := os.Getenv("MSSQL_HOST")
	mssqlPort := os.Getenv("MSSQL_PORT")
	mssqlDB := os.Getenv("MSSQL_DB")
	MaxConnections, err := strconv.Atoi(os.Getenv("DB_MAX_CONNECTIONS"))
	assert.NoError(p.T(), err)
	connectionPath := fmt.Sprintf(
		"sqlserver://%s:%s@%s:%s?database=%s&encrypt=disable",
		mssqlUser,
		mssqlPassword,
		mssqlHost,
		mssqlPort,
		mssqlDB,
	)

	config := database.Config{
		Database:            database.BackendMSSQL,
		ConnectionPath:      connectionPath,
		NumberOfConnections: MaxConnections,
	}

	err = config.Validate()
	assert.NoError(p.T(), err)
	p.DBTestSuite.SetupDB(config, "init/init.up.sql")

	testsuite.SetCustomerTableName("test.customer")
}

func (p *MSSQLTestSuite) TearDownSuite() {
	p.DBTestSuite.TearDownDB("init/init.down.sql")
}

func TestMSSQLSuite(t *testing.T) {
	suite.Run(t, new(MSSQLTestSuite))
}

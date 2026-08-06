package mysql

import (
	"aiml-infrastructure/internal/base/database"
	"aiml-infrastructure/internal/base/database/factory"
	"aiml-infrastructure/internal/base/database/testsuite"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MySQLTestSuite struct {
	testsuite.DBTestSuite
}

func (p *MySQLTestSuite) SetupSuite() {
	err := godotenv.Load("init/.env")
	assert.NoError(p.T(), err)

	mysqlUser := os.Getenv("MYSQL_USER")
	mysqlPassword := os.Getenv("MYSQL_PASSWORD")
	mysqlHost := os.Getenv("MYSQL_HOST")
	mysqlPort := os.Getenv("MYSQL_PORT")
	mysqlDB := os.Getenv("MYSQL_DB")
	MaxConnections, err := strconv.Atoi(os.Getenv("DB_MAX_CONNECTIONS"))
	assert.NoError(p.T(), err)
	connectionPath := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		mysqlUser,
		mysqlPassword,
		mysqlHost,
		mysqlPort,
		mysqlDB,
	)

	p.DBTestSuite.SetupDB(database.Config{
		BackendDB:           factory.BackendMySQL,
		ConnectionPath:      connectionPath,
		NumberOfConnections: MaxConnections,
	}, "init/init.up.sql")

	testsuite.SetCustomerTableName("customer")
}

func (p *MySQLTestSuite) TearDownSuite() {
	p.DBTestSuite.TearDownDB("init/init.down.sql")
}

func TestMySQLSuite(t *testing.T) {
	suite.Run(t, new(MySQLTestSuite))
}

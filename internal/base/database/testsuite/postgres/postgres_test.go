package postgres

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

type PostgresTestSuite struct {
	testsuite.DBTestSuite
}

func (p *PostgresTestSuite) SetupSuite() {
	err := godotenv.Load("init/.env")
	assert.NoError(p.T(), err)

	PgUser := os.Getenv("POSTGRES_USER")
	PgPassword := os.Getenv("POSTGRES_PASSWORD")
	PgHost := os.Getenv("POSTGRES_HOST")
	PgPort := os.Getenv("POSTGRES_PORT")
	PgDB := os.Getenv("POSTGRES_DB")
	PgSSLMode := os.Getenv("POSTGRES_SSL_MODE")
	MaxConnections, err := strconv.Atoi(os.Getenv("DB_MAX_CONNECTIONS"))
	assert.NoError(p.T(), err)
	connectionPath := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		PgUser,
		PgPassword,
		PgHost,
		PgPort,
		PgDB,
		PgSSLMode,
	)

	p.DBTestSuite.SetupDB(database.Config{
		BackendDB:           factory.BackendPostgreSQL,
		ConnectionPath:      connectionPath,
		NumberOfConnections: MaxConnections,
	}, "init/init.up.sql")

	testsuite.SetCustomerTableName("test.customer")
}

func (p *PostgresTestSuite) TearDownSuite() {

	p.DBTestSuite.TearDownDB("init/init.down.sql")

}

func TestPostgresSuite(t *testing.T) {
	suite.Run(t, new(PostgresTestSuite))
}

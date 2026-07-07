package postgresql

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TimestampDAO struct {
	CreatedAt int64  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt int64  `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt *int64 `gorm:"column:deleted_at"`
}

type CustomerDAO struct {
	ID        string `gorm:"type:uuid;primaryKey"`
	FirstName string `gorm:"column:first_name"`
	LastName  string `gorm:"column:last_name"`
	TimestampDAO
}

func (CustomerDAO) TableName() string {
	return "test.customer"
}

type PostgresSuite struct {
	suite.Suite
	DBConnector *postgresDBConnector
}

func (p *PostgresSuite) SetupSuite() {
	err := godotenv.Load("test/.env")
	assert.NoError(p.T(), err)

	PgUser := os.Getenv("POSTGRES_USER")
	PgPassword := os.Getenv("POSTGRES_PASSWORD")
	PgHost := os.Getenv("POSTGRES_HOST")
	PgPort := os.Getenv("POSTGRES_PORT")
	PgDB := os.Getenv("POSTGRES_DB")
	PgSSLMode := os.Getenv("POSTGRES_SSL_MODE")
	PgMaxConnections, err := strconv.Atoi(os.Getenv("POSTGRES_MAX_CONNECTIONS"))
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

	p.DBConnector, err = NewPostgresDBConnector(connectionPath, PgMaxConnections)
	assert.NoError(p.T(), err)

	err = p.DBConnector.Ping(1)
	assert.NoError(p.T(), err)

	initSchema, err := os.ReadFile("test/test_schema.up.sql")
	assert.NoError(p.T(), err)

	err = p.DBConnector.GormDB.Exec(string(initSchema)).Error
	assert.NoError(p.T(), err)
}

func (p *PostgresSuite) TearDownSuite() {

	endSchema, err := os.ReadFile("test/test_schema.down.sql")
	assert.NoError(p.T(), err)

	err = p.DBConnector.GormDB.Exec(string(endSchema)).Error
	assert.NoError(p.T(), err)

	err = p.DBConnector.Close()
	assert.NoError(p.T(), err)

	err = p.DBConnector.Ping(1)
	assert.Error(p.T(), err)
}

func TestPostgresSuite(t *testing.T) {
	suite.Run(t, new(PostgresSuite))
}

func (p *PostgresSuite) Test001Insert() {

	customer := CustomerDAO{
		ID:        uuid.New().String(),
		FirstName: "John",
		LastName:  "Doe",
	}

	err := p.DBConnector.Insert(&customer)
	assert.NoError(p.T(), err)
}

func (p *PostgresSuite) Test002InsertWithConflict() {

	ID := uuid.New().String()

	customer := CustomerDAO{
		ID:        ID,
		FirstName: "John",
		LastName:  "Doe",
	}

	err := p.DBConnector.Insert(&customer)
	assert.NoError(p.T(), err)

	customer.FirstName = "Jane"
	customer.LastName = "Smith"

	err = p.DBConnector.Insert(&customer)
	assert.Error(p.T(), err)
}

func (p *PostgresSuite) Test003UpdateExistingRecord() {

	ID := uuid.New().String()

	customer := CustomerDAO{
		ID:        ID,
		FirstName: "John",
		LastName:  "Doe",
	}

	err := p.DBConnector.Insert(&customer)
	assert.NoError(p.T(), err)

	customer.FirstName = "Jane"
	customer.LastName = "Smith"

	time.Sleep(2 * time.Second)
	err = p.DBConnector.Update(&customer, "FirstName", "LastName")
	assert.NoError(p.T(), err)
}

func (p *PostgresSuite) Test004UpdateWithoutRecord() {

	customer := CustomerDAO{
		ID:        uuid.New().String(),
		FirstName: "John",
		LastName:  "Doe",
	}

	err := p.DBConnector.Update(&customer)
	assert.Error(p.T(), err)
}

func (p *PostgresSuite) Test005UpsertNewRecord() {

	customer := CustomerDAO{
		ID:        uuid.New().String(),
		FirstName: "John",
		LastName:  "Doe",
	}

	err := p.DBConnector.Upsert(&customer)
	assert.NoError(p.T(), err)
}

func (p *PostgresSuite) Test006UpsertExistingRecord() {

	ID := uuid.New().String()

	customer := CustomerDAO{
		ID:        ID,
		FirstName: "John",
		LastName:  "Smith",
	}

	err := p.DBConnector.Upsert(&customer)
	assert.NoError(p.T(), err)

	time.Sleep(2 * time.Second)

	customer.FirstName = "Julia"
	customer.LastName = "Wiwin"

	err = p.DBConnector.Upsert(&customer)
	assert.NoError(p.T(), err)
}

func (p *PostgresSuite) Test007HardDelete() {

	ID := uuid.New().String()

	customer := CustomerDAO{
		ID:        ID,
		FirstName: "John",
		LastName:  "Smith",
	}

	err := p.DBConnector.Insert(&customer)
	assert.NoError(p.T(), err)

	err = p.DBConnector.HardDelete(&customer)
	assert.NoError(p.T(), err)
}

func (p *PostgresSuite) Test008HardDeleteWithoutRecord() {

	customer := CustomerDAO{
		ID:        uuid.NewString(),
		FirstName: "John",
		LastName:  "Smith",
	}

	err := p.DBConnector.HardDelete(&customer)
	assert.Error(p.T(), err)
}

func (p *PostgresSuite) Test009SoftDelete() {

	customer := CustomerDAO{
		ID:        uuid.New().String(),
		FirstName: "Stefanus",
		LastName:  "Yudi",
	}

	err := p.DBConnector.Insert(&customer)
	assert.NoError(p.T(), err)

	time.Sleep(2 * time.Second)

	err = p.DBConnector.SoftDelete(&customer)
	assert.NoError(p.T(), err)
}

func (p *PostgresSuite) Test010SoftDeleteWithoutRecord() {

	customer := CustomerDAO{
		ID:        uuid.NewString(),
		FirstName: "Stefanus",
		LastName:  "Yudi",
	}

	err := p.DBConnector.SoftDelete(&customer)
	assert.Error(p.T(), err)
}

func (p *PostgresSuite) Test011RestoreExistingRecord() {
	customer := CustomerDAO{
		ID:        uuid.New().String(),
		FirstName: "Stefanus",
		LastName:  "Yudi",
	}

	err := p.DBConnector.Insert(&customer)
	assert.NoError(p.T(), err)

	time.Sleep(2 * time.Second)

	err = p.DBConnector.SoftDelete(&customer)
	assert.NoError(p.T(), err)

	time.Sleep(2 * time.Second)

	err = p.DBConnector.Restore(&customer)
	assert.NoError(p.T(), err)
}

func (p *PostgresSuite) Test012RestoreWithoutRecord() {

	customer := CustomerDAO{
		ID:        uuid.New().String(),
		FirstName: "Stefanus",
		LastName:  "Yudi",
	}

	err := p.DBConnector.Restore(&customer)
	assert.Error(p.T(), err)

}

func (p *PostgresSuite) Test013Exists() {

	customer := CustomerDAO{
		ID:        uuid.New().String(),
		FirstName: "Stefanus",
		LastName:  "Yudi",
	}

	err := p.DBConnector.Insert(&customer)
	assert.NoError(p.T(), err)

	isExists, err := p.DBConnector.Exists(&customer)
	assert.NoError(p.T(), err)
	assert.Equal(p.T(), isExists, true)
}

func (p *PostgresSuite) Test014ExistsWithoutRecord() {

	customer := CustomerDAO{
		ID:        uuid.New().String(),
		FirstName: "Stefanus",
		LastName:  "Yudi",
	}

	isExists, err := p.DBConnector.Exists(&customer)
	assert.NoError(p.T(), err)
	assert.Equal(p.T(), isExists, false)

}

func (p *PostgresSuite) Test015GetByPrimaryKeys() {

	ID := uuid.NewString()

	customer := CustomerDAO{
		ID:        ID,
		FirstName: "Stefanus",
		LastName:  "Yudi",
	}

	err := p.DBConnector.Insert(&customer)
	assert.NoError(p.T(), err)

	emptyCustomer := CustomerDAO{
		ID: ID,
	}

	err = p.DBConnector.GetByPrimaryKeys(&emptyCustomer)
	assert.NoError(p.T(), err)
	assert.Equal(p.T(), emptyCustomer.FirstName, "Stefanus")
	assert.Equal(p.T(), emptyCustomer.LastName, "Yudi")
}

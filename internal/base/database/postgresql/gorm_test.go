package postgresql

import (
	"context"
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

	err = p.DBConnector.ClientDB.Ping()
	assert.NoError(p.T(), err)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err = p.DBConnector.ClientDB.PingContext(ctx)
	assert.NoError(p.T(), err)

	initSchema, err := os.ReadFile("test/test_schema.up.sql")
	assert.NoError(p.T(), err)

	err = p.DBConnector.GormDB.Exec(string(initSchema)).Error
	assert.NoError(p.T(), err)
}

func (p *PostgresSuite) TearDownSuite() {

	// endSchema, err := os.ReadFile("test/test_schema.down.sql")
	// assert.NoError(p.T(), err)

	// err = p.DBConnector.GormDB.Exec(string(endSchema)).Error
	// assert.NoError(p.T(), err)

	connection, err := p.DBConnector.GormDB.DB()
	assert.NoError(p.T(), err)
	err = connection.Close()
	assert.NoError(p.T(), err)
}

func TestPostgresSuite(t *testing.T) {
	suite.Run(t, new(PostgresSuite))
}

func (p *PostgresSuite) Test001InsertData() {

	customer := CustomerDAO{
		ID:        uuid.New().String(),
		FirstName: "John",
		LastName:  "Doe",
	}

	err := p.DBConnector.InsertData(&customer)
	assert.NoError(p.T(), err)
}

func (p *PostgresSuite) Test002InsertDataWithConflict() {

	ID := uuid.New().String()

	customer := CustomerDAO{
		ID:        ID,
		FirstName: "John",
		LastName:  "Doe",
	}

	err := p.DBConnector.InsertData(&customer)
	assert.NoError(p.T(), err)

	customer.FirstName = "Jane"
	customer.LastName = "Smith"

	err = p.DBConnector.InsertData(&customer)
	assert.Error(p.T(), err)
}

func (p *PostgresSuite) Test003UpdateData() {

	ID := uuid.New().String()

	customer := CustomerDAO{
		ID:        ID,
		FirstName: "John",
		LastName:  "Doe",
	}

	err := p.DBConnector.InsertData(&customer)
	assert.NoError(p.T(), err)

	fmt.Println(customer.CreatedAt)
	fmt.Println(customer.UpdatedAt)

	customer.FirstName = "Jane"
	customer.LastName = "Smith"

	time.Sleep(2 * time.Second)
	err = p.DBConnector.UpdateData(&customer, "FirstName", "LastName")
	assert.NoError(p.T(), err)

	fmt.Println(customer.CreatedAt)
	fmt.Println(customer.UpdatedAt)
}

func (p *PostgresSuite) Test004UpdateDataWithConflict() {

	customer := CustomerDAO{
		ID:        uuid.New().String(),
		FirstName: "John",
		LastName:  "Doe",
	}

	err := p.DBConnector.UpdateData(&customer)
	assert.Error(p.T(), err)
}

func (p *PostgresSuite) Test005UpsertData() {

	customer := CustomerDAO{
		ID:        uuid.New().String(),
		FirstName: "John",
		LastName:  "Doe",
	}

	err := p.DBConnector.UpsertData(&customer)
	assert.NoError(p.T(), err)
}

func (p *PostgresSuite) Test005UpsertDataWithConflict() {

	ID := uuid.New().String()

	customer := CustomerDAO{
		ID:        ID,
		FirstName: "John",
		LastName:  "Smith",
	}

	err := p.DBConnector.UpsertData(&customer)
	assert.NoError(p.T(), err)

	time.Sleep(2 * time.Second)

	customer.FirstName = "Julia"
	customer.LastName = "Wiwin"

	err = p.DBConnector.UpsertData(&customer)
	assert.NoError(p.T(), err)
}

func (p *PostgresSuite) Test005DeleteData() {

	ID := uuid.New().String()

	customer := CustomerDAO{
		ID:        ID,
		FirstName: "John",
		LastName:  "Smith",
	}

	err := p.DBConnector.InsertData(&customer)
	assert.NoError(p.T(), err)

	err = p.DBConnector.DeleteData(&customer)
	assert.NoError(p.T(), err)
}

func TestFormatConflictColumns(t *testing.T) {
	test := formatConflictColumns([]string{"id"})
	fmt.Println(test)
}

package mssql

import (
	"aiml-infrastructure/internal/base/database"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlserver"
)

type TimestampDAO struct {
	CreatedAt int64  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt int64  `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt *int64 `gorm:"column:deleted_at"`
}

type CustomerDAO struct {
	ID        string `gorm:"type:uniqueidentifier;primaryKey"`
	FirstName string `gorm:"column:first_name"`
	LastName  string `gorm:"column:last_name"`
	TimestampDAO
}

func (CustomerDAO) TableName() string {
	return "test.customer"
}

type DBSuite struct {
	suite.Suite
	DBConnector *database.DBConnector
}

func (p *DBSuite) SetupSuite() {
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

	p.DBConnector, err = database.NewDBConnector(sqlserver.Open(connectionPath), MaxConnections)
	assert.NoError(p.T(), err)

	err = p.DBConnector.Ping(1)
	assert.NoError(p.T(), err)

	initSchema, err := os.ReadFile("init/init.up.sql")
	assert.NoError(p.T(), err)

	err = p.DBConnector.GormDB.Exec(string(initSchema)).Error
	assert.NoError(p.T(), err)
}

func (p *DBSuite) TearDownSuite() {

	endSchema, err := os.ReadFile("init/init.down.sql")
	assert.NoError(p.T(), err)

	err = p.DBConnector.GormDB.Exec(string(endSchema)).Error
	assert.NoError(p.T(), err)

	err = p.DBConnector.Close()
	assert.NoError(p.T(), err)

	err = p.DBConnector.Ping(1)
	assert.Error(p.T(), err)
}

func TestMSSQLSuite(t *testing.T) {
	suite.Run(t, new(DBSuite))
}

func (p *DBSuite) Test001Insert() {

	customer := CustomerDAO{
		ID:        uuid.New().String(),
		FirstName: "John",
		LastName:  "Doe",
	}

	err := p.DBConnector.Insert(&customer)
	assert.NoError(p.T(), err)
}

func (p *DBSuite) Test002InsertWithConflict() {

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

func (p *DBSuite) Test003UpdateExistingRecord() {

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

func (p *DBSuite) Test004UpdateWithoutRecord() {

	customer := CustomerDAO{
		ID:        uuid.New().String(),
		FirstName: "John",
		LastName:  "Doe",
	}

	err := p.DBConnector.Update(&customer)
	assert.Error(p.T(), err)
}

func (p *DBSuite) Test005UpsertNewRecord() {

	customer := CustomerDAO{
		ID:        uuid.New().String(),
		FirstName: "John",
		LastName:  "Doe",
	}

	err := p.DBConnector.Upsert(&customer)
	assert.NoError(p.T(), err)
}

func (p *DBSuite) Test006UpsertExistingRecord() {

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

func (p *DBSuite) Test007HardDelete() {

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

func (p *DBSuite) Test008HardDeleteWithoutRecord() {

	customer := CustomerDAO{
		ID:        uuid.NewString(),
		FirstName: "John",
		LastName:  "Smith",
	}

	err := p.DBConnector.HardDelete(&customer)
	assert.Error(p.T(), err)
}

func (p *DBSuite) Test009SoftDelete() {

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

func (p *DBSuite) Test010SoftDeleteWithoutRecord() {

	customer := CustomerDAO{
		ID:        uuid.NewString(),
		FirstName: "Stefanus",
		LastName:  "Yudi",
	}

	err := p.DBConnector.SoftDelete(&customer)
	assert.Error(p.T(), err)
}

func (p *DBSuite) Test011RestoreExistingRecord() {
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

func (p *DBSuite) Test012RestoreWithoutRecord() {

	customer := CustomerDAO{
		ID:        uuid.New().String(),
		FirstName: "Stefanus",
		LastName:  "Yudi",
	}

	err := p.DBConnector.Restore(&customer)
	assert.Error(p.T(), err)

}

func (p *DBSuite) Test013Exists() {

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

func (p *DBSuite) Test014ExistsWithoutRecord() {

	customer := CustomerDAO{
		ID:        uuid.New().String(),
		FirstName: "Stefanus",
		LastName:  "Yudi",
	}

	isExists, err := p.DBConnector.Exists(&customer)
	assert.NoError(p.T(), err)
	assert.Equal(p.T(), isExists, false)

}

func (p *DBSuite) Test015GetByPrimaryKeys() {

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

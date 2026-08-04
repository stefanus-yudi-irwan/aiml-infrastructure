package testsuite

import (
	"aiml-infrastructure/internal/base/database"
	"aiml-infrastructure/internal/base/database/factory"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TimestampDAO struct {
	CreatedAt int64  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt int64  `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt *int64 `gorm:"column:deleted_at"`
}

type CustomerDAO struct {
	ID        string `gorm:"primaryKey"`
	FirstName string `gorm:"column:first_name"`
	LastName  string `gorm:"column:last_name"`
	TimestampDAO
}

var customerTableName = "test.customer"

func SetCustomerTableName(tableName string) {
	customerTableName = tableName
}

func (CustomerDAO) TableName() string {
	return customerTableName
}

type DBTestSuite struct {
	suite.Suite
	DBConnector database.IDBConnector
}

func (d *DBTestSuite) SetupDB(config database.Config, initUpSQLPath string) {

	var err error

	d.DBConnector, err = factory.NewDBConnector(config)
	assert.NoError(d.T(), err)

	err = d.DBConnector.Ping(1)
	assert.NoError(d.T(), err)

	initSchema, err := os.ReadFile(initUpSQLPath)
	assert.NoError(d.T(), err)

	err = d.DBConnector.ExecuteSQL(string(initSchema))
	assert.NoError(d.T(), err)

}

func (p *DBTestSuite) TearDownDB(initDownSQLPath string) {

	endSchema, err := os.ReadFile(initDownSQLPath)
	assert.NoError(p.T(), err)

	err = p.DBConnector.ExecuteSQL(string(endSchema))
	assert.NoError(p.T(), err)

	err = p.DBConnector.Close()
	assert.NoError(p.T(), err)

	err = p.DBConnector.Ping(1)
	assert.Error(p.T(), err)
}

func (p *DBTestSuite) Test001Insert() {

	customer := CustomerDAO{
		ID:        uuid.New().String(),
		FirstName: "John",
		LastName:  "Doe",
	}

	err := p.DBConnector.Insert(&customer)
	assert.NoError(p.T(), err)
}

func (p *DBTestSuite) Test002InsertWithConflict() {

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

func (p *DBTestSuite) Test003UpdateExistingRecord() {

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

func (p *DBTestSuite) Test004UpdateWithoutRecord() {

	customer := CustomerDAO{
		ID:        uuid.New().String(),
		FirstName: "John",
		LastName:  "Doe",
	}

	err := p.DBConnector.Update(&customer)
	assert.Error(p.T(), err)
}

func (p *DBTestSuite) Test005UpsertNewRecord() {

	customer := CustomerDAO{
		ID:        uuid.New().String(),
		FirstName: "John",
		LastName:  "Doe",
	}

	err := p.DBConnector.Upsert(&customer)
	assert.NoError(p.T(), err)
}

func (p *DBTestSuite) Test006UpsertExistingRecord() {

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

func (p *DBTestSuite) Test007HardDelete() {

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

func (p *DBTestSuite) Test008HardDeleteWithoutRecord() {

	customer := CustomerDAO{
		ID:        uuid.NewString(),
		FirstName: "John",
		LastName:  "Smith",
	}

	err := p.DBConnector.HardDelete(&customer)
	assert.Error(p.T(), err)
}

func (p *DBTestSuite) Test009SoftDelete() {

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

func (p *DBTestSuite) Test010SoftDeleteWithoutRecord() {

	customer := CustomerDAO{
		ID:        uuid.NewString(),
		FirstName: "Stefanus",
		LastName:  "Yudi",
	}

	err := p.DBConnector.SoftDelete(&customer)
	assert.Error(p.T(), err)
}

func (p *DBTestSuite) Test011RestoreExistingRecord() {
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

func (p *DBTestSuite) Test012RestoreWithoutRecord() {

	customer := CustomerDAO{
		ID:        uuid.New().String(),
		FirstName: "Stefanus",
		LastName:  "Yudi",
	}

	err := p.DBConnector.Restore(&customer)
	assert.Error(p.T(), err)

}

func (p *DBTestSuite) Test013Exists() {

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

func (p *DBTestSuite) Test014ExistsWithoutRecord() {

	customer := CustomerDAO{
		ID:        uuid.New().String(),
		FirstName: "Stefanus",
		LastName:  "Yudi",
	}

	isExists, err := p.DBConnector.Exists(&customer)
	assert.NoError(p.T(), err)
	assert.Equal(p.T(), isExists, false)

}

func (p *DBTestSuite) Test015GetByPrimaryKeys() {

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

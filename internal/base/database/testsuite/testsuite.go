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

	err = d.DBConnector.Ping(d.T().Context(), 1)
	assert.NoError(d.T(), err)

	initSchema, err := os.ReadFile(initUpSQLPath)
	assert.NoError(d.T(), err)

	err = d.DBConnector.ExecuteSQL(d.T().Context(), string(initSchema))
	assert.NoError(d.T(), err)

}

func (p *DBTestSuite) TearDownDB(initDownSQLPath string) {

	endSchema, err := os.ReadFile(initDownSQLPath)
	assert.NoError(p.T(), err)

	err = p.DBConnector.ExecuteSQL(p.T().Context(), string(endSchema))
	assert.NoError(p.T(), err)

	err = p.DBConnector.Close()
	assert.NoError(p.T(), err)

	err = p.DBConnector.Ping(p.T().Context(), 1)
	assert.Error(p.T(), err)
}

func (p *DBTestSuite) Test001Insert() {

	customer := CustomerDAO{
		ID:        uuid.New().String(),
		FirstName: "John",
		LastName:  "Doe",
	}

	err := database.ValidateStructPointer(&customer)
	assert.NoError(p.T(), err)

	err = p.DBConnector.Insert(p.T().Context(), &customer)
	assert.NoError(p.T(), err)
}

func (p *DBTestSuite) Test002InsertWithConflict() {

	ID := uuid.New().String()

	customer := CustomerDAO{
		ID:        ID,
		FirstName: "John",
		LastName:  "Doe",
	}

	err := database.ValidateStructPointer(&customer)
	assert.NoError(p.T(), err)

	err = p.DBConnector.Insert(p.T().Context(), &customer)
	assert.NoError(p.T(), err)

	customer.FirstName = "Jane"
	customer.LastName = "Smith"

	err = p.DBConnector.Insert(p.T().Context(), &customer)
	assert.Error(p.T(), err)
}

func (p *DBTestSuite) Test003UpdateExistingRecord() {

	ID := uuid.New().String()

	customer := CustomerDAO{
		ID:        ID,
		FirstName: "John",
		LastName:  "Doe",
	}

	err := database.ValidateStructPointer(&customer)
	assert.NoError(p.T(), err)

	err = p.DBConnector.Insert(p.T().Context(), &customer)
	assert.NoError(p.T(), err)

	customer.FirstName = "Jane"
	customer.LastName = "Smith"

	time.Sleep(2 * time.Second)
	err = p.DBConnector.Update(p.T().Context(), &customer, "FirstName", "LastName")
	assert.NoError(p.T(), err)
}

func (p *DBTestSuite) Test004UpdateWithoutRecord() {

	customer := CustomerDAO{
		ID:        uuid.New().String(),
		FirstName: "John",
		LastName:  "Doe",
	}

	err := database.ValidateStructPointer(&customer)
	assert.NoError(p.T(), err)

	err = p.DBConnector.Update(p.T().Context(), &customer)
	assert.Error(p.T(), err)
}

func (p *DBTestSuite) Test005UpsertNewRecord() {

	customer := CustomerDAO{
		ID:        uuid.New().String(),
		FirstName: "John",
		LastName:  "Doe",
	}

	err := database.ValidateStructPointer(&customer)
	assert.NoError(p.T(), err)

	err = p.DBConnector.Upsert(p.T().Context(), &customer)
	assert.NoError(p.T(), err)
}

func (p *DBTestSuite) Test006UpsertExistingRecord() {

	ID := uuid.New().String()

	customer := CustomerDAO{
		ID:        ID,
		FirstName: "John",
		LastName:  "Smith",
	}

	err := database.ValidateStructPointer(&customer)
	assert.NoError(p.T(), err)

	err = p.DBConnector.Upsert(p.T().Context(), &customer)
	assert.NoError(p.T(), err)

	time.Sleep(2 * time.Second)

	customer.FirstName = "Julia"
	customer.LastName = "Wiwin"

	err = p.DBConnector.Upsert(p.T().Context(), &customer)
	assert.NoError(p.T(), err)
}

func (p *DBTestSuite) Test007HardDelete() {

	ID := uuid.New().String()

	customer := CustomerDAO{
		ID:        ID,
		FirstName: "John",
		LastName:  "Smith",
	}

	err := database.ValidateStructPointer(&customer)
	assert.NoError(p.T(), err)

	err = p.DBConnector.Insert(p.T().Context(), &customer)
	assert.NoError(p.T(), err)

	err = p.DBConnector.HardDelete(p.T().Context(), &customer)
	assert.NoError(p.T(), err)
}

func (p *DBTestSuite) Test008HardDeleteWithoutRecord() {

	customer := CustomerDAO{
		ID:        uuid.NewString(),
		FirstName: "John",
		LastName:  "Smith",
	}

	err := database.ValidateStructPointer(&customer)
	assert.NoError(p.T(), err)

	err = p.DBConnector.HardDelete(p.T().Context(), &customer)
	assert.Error(p.T(), err)
}

func (p *DBTestSuite) Test009SoftDelete() {

	customer := CustomerDAO{
		ID:        uuid.New().String(),
		FirstName: "Stefanus",
		LastName:  "Yudi",
	}

	err := database.ValidateStructPointer(&customer)
	assert.NoError(p.T(), err)

	err = p.DBConnector.Insert(p.T().Context(), &customer)
	assert.NoError(p.T(), err)

	time.Sleep(2 * time.Second)

	err = p.DBConnector.SoftDelete(p.T().Context(), &customer)
	assert.NoError(p.T(), err)
}

func (p *DBTestSuite) Test010SoftDeleteWithoutRecord() {

	customer := CustomerDAO{
		ID:        uuid.NewString(),
		FirstName: "Stefanus",
		LastName:  "Yudi",
	}

	err := database.ValidateStructPointer(&customer)
	assert.NoError(p.T(), err)

	err = p.DBConnector.SoftDelete(p.T().Context(), &customer)
	assert.Error(p.T(), err)
}

func (p *DBTestSuite) Test011RestoreExistingRecord() {
	customer := CustomerDAO{
		ID:        uuid.New().String(),
		FirstName: "Stefanus",
		LastName:  "Yudi",
	}

	err := database.ValidateStructPointer(&customer)
	assert.NoError(p.T(), err)

	err = p.DBConnector.Insert(p.T().Context(), &customer)
	assert.NoError(p.T(), err)

	time.Sleep(2 * time.Second)

	err = p.DBConnector.SoftDelete(p.T().Context(), &customer)
	assert.NoError(p.T(), err)

	time.Sleep(2 * time.Second)

	err = p.DBConnector.Restore(p.T().Context(), &customer)
	assert.NoError(p.T(), err)
}

func (p *DBTestSuite) Test012RestoreWithoutRecord() {

	customer := CustomerDAO{
		ID:        uuid.New().String(),
		FirstName: "Stefanus",
		LastName:  "Yudi",
	}

	err := database.ValidateStructPointer(&customer)
	assert.NoError(p.T(), err)

	err = p.DBConnector.Restore(p.T().Context(), &customer)
	assert.Error(p.T(), err)

}

func (p *DBTestSuite) Test013Exists() {

	customer := CustomerDAO{
		ID:        uuid.New().String(),
		FirstName: "Stefanus",
		LastName:  "Yudi",
	}

	err := database.ValidateStructPointer(&customer)
	assert.NoError(p.T(), err)

	err = p.DBConnector.Insert(p.T().Context(), &customer)
	assert.NoError(p.T(), err)

	isExists, err := p.DBConnector.Exists(p.T().Context(), &customer)
	assert.NoError(p.T(), err)
	assert.Equal(p.T(), isExists, true)
}

func (p *DBTestSuite) Test014ExistsWithoutRecord() {

	customer := CustomerDAO{
		ID:        uuid.New().String(),
		FirstName: "Stefanus",
		LastName:  "Yudi",
	}

	err := database.ValidateStructPointer(&customer)
	assert.NoError(p.T(), err)

	isExists, err := p.DBConnector.Exists(p.T().Context(), &customer)
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

	err := database.ValidateStructPointer(&customer)
	assert.NoError(p.T(), err)

	err = p.DBConnector.Insert(p.T().Context(), &customer)
	assert.NoError(p.T(), err)

	emptyCustomer := CustomerDAO{
		ID: ID,
	}

	err = p.DBConnector.GetByPrimaryKeys(p.T().Context(), &emptyCustomer)
	assert.NoError(p.T(), err)
	assert.Equal(p.T(), emptyCustomer.FirstName, "Stefanus")
	assert.Equal(p.T(), emptyCustomer.LastName, "Yudi")
}

func (p *DBTestSuite) Test016TestStructPointerValidation() {
	customer := CustomerDAO{
		ID:        uuid.New().String(),
		FirstName: "Stefanus",
		LastName:  "Yudi",
	}
	assert.NoError(p.T(), database.ValidateStructPointer(&customer))

	var customer2 *CustomerDAO = nil
	assert.Error(p.T(), database.ValidateStructPointer(customer2))

	var value interface{} = customer2
	assert.Error(p.T(), database.ValidateStructPointer(value))
}

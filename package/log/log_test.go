package log

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type LoggerSuite struct {
	suite.Suite
	logInstance logger
}

func TestLogger(t *testing.T) {
	suite.Run(t, new(LoggerSuite))
}

func (m *LoggerSuite) SetupSuite() {

}

func (m *LoggerSuite) TearDownSuite() {

}

func (m *LoggerSuite) Test001LogInfo() {

}

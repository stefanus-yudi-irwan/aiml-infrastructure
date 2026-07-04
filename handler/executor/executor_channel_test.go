package executor

import (
	"aiml-infrastructure/package/payload"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type mockExecutor int

type mockData struct {
	Id   string
	Name string
}

func (m *mockExecutor) Execute(payload payload.Payload[mockData]) error {
	fmt.Printf("Data: %v ; Attribute %v \n", payload.Data, payload.Attribute)
	return nil
}

type ExecutorChannelSuite struct {
	suite.Suite
	executorChannel *ExecutorChannel[mockData]
}

func TestExecutorChannel(t *testing.T) {
	suite.Run(t, new(ExecutorChannelSuite))
}

func (m *ExecutorChannelSuite) SetupSuite() {

}

func (m *ExecutorChannelSuite) TearDownSuite() {
	log.

}

func (m *ExecutorChannelSuite) Test001Execute() {

}

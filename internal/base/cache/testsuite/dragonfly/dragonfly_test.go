package dragonfly

import (
	"aiml-infrastructure/internal/base/cache"
	"aiml-infrastructure/internal/base/cache/testsuite"
	"testing"

	"github.com/stretchr/testify/suite"
)

type DragonflyTestSuite struct {
	testsuite.CacheTestSuite
}

func (d *DragonflyTestSuite) SetupSuite() {
	d.CacheTestSuite.SetupCache(".env", cache.BackendDragonfly)
}

func (d *DragonflyTestSuite) TearDownSuite() {
	d.CacheTestSuite.TearDownCache()
}

func TestDragonflyTestSuite(t *testing.T) {
	suite.Run(t, new(DragonflyTestSuite))
}

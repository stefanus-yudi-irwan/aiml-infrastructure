package valkey

import (
	"aiml-infrastructure/internal/base/cache"
	"aiml-infrastructure/internal/base/cache/testsuite"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ValkeyTestSuite struct {
	testsuite.CacheTestSuite
}

func (r *ValkeyTestSuite) SetupSuite() {
	r.CacheTestSuite.SetupCache(".env", cache.BackendValkey)
}

func (r *ValkeyTestSuite) TearDownSuite() {
	r.CacheTestSuite.TearDownCache()
}

func TestValkeyTestSuite(t *testing.T) {
	suite.Run(t, new(ValkeyTestSuite))
}

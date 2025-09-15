package allurego

import (
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
)


type SuiteManager struct {
	suite.Suite
}

func (s *SuiteManager) AfterAll(t provider.T) {
	// Add support for multiple assertions within a single test
	if t.Failed() {
		t.Fail()
	}
}
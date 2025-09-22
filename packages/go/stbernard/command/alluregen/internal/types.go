package alluregen

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
)

type Event struct {
	Time        string
	Action      string
	Package     string
	Test        string
	Elapsed     float64
	Output      string
	FailedBuild string
}

type Label struct {
	Name  string
	Value string
}

type StatusDetails struct {
	Message string
	Trace   string
}

type TestCase struct {
	Name          string
	FullName      string
	Status        string
	StatusDetails StatusDetails
	Start         int64
	Stop          int64
	UUID          string
	HistoryId     string
	TestCaseId    string
	Description   string
	Labels        []Label
}

type Package struct {
	Name  string
	Tests []TestCase
}

type TestResults struct {
	Pkgs []Package
}

func (ts *TestCase) create(e Event, eventoutputs map[string][]string) *TestCase {
	switch e.Action {
	case "pass":
		ts.Status = "passed"
	case "fail":
		ts.Status = "failed"
	}

	// Create a hash from test case fullname to generate a stateful reporting
	calTestCaseHash := hashString(e.Test)
	testCaseUUID := fmt.Sprintf("%x", calTestCaseHash)
	calHashHistory := hashString(testCaseUUID)
	historyID := fmt.Sprintf("%x", calHashHistory)

	ts.FullName = e.Test
	ts.Name = e.Test
	ts.UUID = uuid.New().String()
	ts.TestCaseId = testCaseUUID
	ts.HistoryId = historyID
	ts.StatusDetails = StatusDetails{
		Message: strings.Join(eventoutputs[ts.FullName], "\n"),
	}
	pkg := strings.Split(e.Package, "/")

	ts.Labels = []Label{
		{
			Name:  "parentSuite",
			Value: "BHCE - Go Tests",
		},
		{
			Name:  "subSuite",
			Value: pkg[len(pkg)-1],
		},
		{
			Name:  "severity",
			Value: "critical",
		},
	}

	return ts
}

// parseEventsOuput retruns an output per test case
func parseEventsOuput(events []Event) map[string][]string {
	outputs := map[string][]string{}
	for _, e := range events {
		switch e.Action {
		case "output":
			if !strings.Contains(e.Output, "WARN") && !strings.Contains(e.Output, "ok") && !strings.Contains(e.Output, "?") {
				outputs[e.Test] = append(outputs[e.Test], strings.TrimSpace(e.Output))
			}
		}
	}

	return outputs
}

func (pkg *Package) addTestCase(ts TestCase) *Package {
	pkg.Name = ts.Name
	pkg.Tests = append(pkg.Tests, ts)
	return pkg
}

func (r *TestResults) generateTests(envVariable string) {
	allurePath := os.Getenv(envVariable)
	for _, pkg := range r.Pkgs {
		for _, tstCase := range pkg.Tests {
			name := fmt.Sprintf(allurePath+"/allure-results/%v-result.json", tstCase.HistoryId)
			f, err := os.Create(name)
			if err != nil {
				log.Fatal(err)
			}

			defer f.Close()

			data, err := json.MarshalIndent(tstCase, "", "\t")
			if err != nil {
				log.Fatal(err)
			}

			f.Write(data)
		}
	}
}

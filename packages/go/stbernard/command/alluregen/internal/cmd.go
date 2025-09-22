package alluregen

func Run() error {

	events, err := parseEvents("/Users/ykaiboussi/Documents/shared/bloodhound-enterprise/bhce/go_bhce_passed_sliceexttests.out")
	if err != nil {
		return err
	}

	// capture all Go packages and related test cases
	results := new(TestResults)

	eventOutput := parseEventsOuput(events)

	for _, event := range events {
		// instantiate new estCase and Package
		tstCase := new(TestCase)
		pkg := new(Package)

		// Create each test cases and append the results to TestResult struct
		if event.Test != "" && event.Action == "pass" || event.Action == "fail" {
			tstCase.create(event, eventOutput)
			pk := pkg.addTestCase(*tstCase)
			results.Pkgs = append(results.Pkgs, *pk)
		}
	}

	// Create JSON files at test case level
	results.generateTests("ALLURE_OUTPUT_PATH")
	return nil
}

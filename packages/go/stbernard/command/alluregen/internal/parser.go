package alluregen

import (
	"bufio"
	"encoding/json"
	"os"
)

// parseEvents returns a list Events struct of go test results
func parseEvents(fileName string) ([]Event, error) {
	file, err := os.OpenFile(fileName, os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}

	fileReader := bufio.NewReader(file)
	fileScanner := bufio.NewScanner(fileReader)

	var e Event
	var events []Event
	for fileScanner.Scan() {
		if err := json.Unmarshal([]byte(fileScanner.Text()), &e); err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	return events, nil
}

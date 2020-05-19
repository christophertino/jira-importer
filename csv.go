package jiraimporter

import (
	"os"

	"github.com/gocarina/gocsv"
)

// Parse the CSV of exported Jira issues
func parseCSV(csvPath string) ([]*Issue, error) {
	f, err := os.OpenFile(csvPath, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	issues := []*Issue{}

	if err := gocsv.UnmarshalFile(f, &issues); err != nil {
		return nil, err
	}

	return issues, nil
}

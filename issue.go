package jiraimporter

import (
	"fmt"
	"os"

	"github.com/gocarina/gocsv"
)

// Issue is a row from the Jira import file
type Issue struct {
	Key  string `csv:"Issue key"`
	Type string `csv:"Issue Type"`
}

// MigrateIssues is the entry point for the Jira import migration
func (ji *JiraImporter) MigrateIssues() {
	// Load the CSV export
	issues, err := parseCSV(ji.CSVPath)
	if err != nil {
		fmt.Println("Error parsing CSV file:", err)
		return
	}

	// Convert issues to correct types
	for _, issue := range issues {
		// fmt.Println(issue.Key)
	}

	// Epics become stories and subtasks are relationships

	// Handle subtasks by adding them to correct parents

	// Import story points

	// Components become labels
}

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

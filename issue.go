package jiraimporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gocarina/gocsv"
)

// Issue is a row from the Jira import file
type Issue struct {
	IssueID     string `csv:"Issue id"`
	IssueKey    string `csv:"Issue key"`
	IssueType   string `csv:"Issue Type"`
	ParentID    string `csv:"Parent id"`
	ProjectKey  string `csv:"Project key"`
	Labels      string `csv:"Labels"`
	EpicLink    string `csv:"Custom field (Epic Link)"`
	StoryPoints string `csv:"Custom field (Story Points)"`
}

// Creates the PUT body data used to update Jira issues
type issueUpdateData struct {
	Update struct {
		StoryPoints []struct {
			Set string `json:"set,omitempty"`
		} `json:"customfield_10016,omitempty"`
	} `json:"update,omitempty"`
	Fields struct {
		Issuetype struct {
			ID string `json:"id,omitempty"`
		} `json:"issuetype,omitempty"`
	} `json:"fields,omitempty"`
}

// MigrateIssues is the entry point for the Jira import migration
func (ji *JiraImporter) MigrateIssues() {
	// Load the CSV export
	issues, err := parseCSV(ji.CSVPath)
	if err != nil {
		fmt.Println("Error parsing CSV file:", err)
		return
	}

	// Get Project info
	project, err := ji.getProject(issues[0].ProjectKey)
	if err != nil {
		fmt.Println("Error getting Project info:", err)
		return
	}

	// Convert issues to correct types
	for _, issue := range issues {
		var updateData = issueUpdateData{}

		switch issue.IssueType {
		case "Epic":
			// We can't convert existing types to Epics, so they become Stories instead
			issue.IssueType = "Story"
			break
		case "Sub-task":
			// Next-gen removes hyphen in Sub-task
			issue.IssueType = "Subtask"
			break
		}

		typeID, err := getIssueTypeID(issue.IssueType, project)
		if err != nil {
			fmt.Printf("Error getting type for issue %s: %s\n", issue.IssueKey, err)
			continue
		}
		updateData.Fields.Issuetype.ID = typeID

		// Import story points
		if issue.StoryPoints != "" {
			updateData.Update.StoryPoints[0].Set = issue.StoryPoints
		}

		// TODO: handle rate limiting
		if err = ji.updateIssue(issue.IssueKey, &updateData); err != nil {
			fmt.Printf("Error updating issue %s: %s\n", issue.IssueKey, err)
			continue
		}
	}

	// Now that Issue types are correct, we can set relationships

	// Epics become stories and subtasks are relationships

	// Handle subtasks by adding them to correct parents

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

// Send Jira issue update
func (ji *JiraImporter) updateIssue(issueID string, data *issueUpdateData) error {
	bytesMessage, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// fmt.Println(string(bytesMessage))

	_, err = ji.sendJiraRequest(http.MethodPut, fmt.Sprintf("/issue/%s?notifyUsers=false", issueID), bytes.NewBuffer(bytesMessage))
	if err != nil {
		return err
	}

	return nil
}

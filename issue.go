package jiraimporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Issue is a row from the Jira import file
type Issue struct {
	IssueID     string `csv:"Issue id"`  // JiraID
	IssueKey    string `csv:"Issue key"` // Public ID (i.e. GL-27)
	IssueType   string `csv:"Issue Type"`
	ParentID    string `csv:"Parent id"` // JiraID
	ProjectKey  string `csv:"Project key"`
	Labels      string `csv:"Labels"`
	EpicLink    string `csv:"Custom field (Epic Link)"`
	StoryPoints string `csv:"Custom field (Story Points)"`
	IssueSplit  string `csv:"Outward issue link (Issue split)"`
}

// Creates the PUT body data used to update Jira issues
type issueUpdateData struct {
	Update struct {
		StoryPoints []struct {
			Set int `json:"set,omitempty"`
		} `json:"customfield_10016,omitempty"`
	} `json:"update,omitempty"`
	Fields struct {
		IssueType struct {
			ID string `json:"id,omitempty"`
		} `json:"issuetype,omitempty"`
		Parent struct {
			Key string `json:"key,omitempty"`
		} `json:"parent,omitempty"`
	} `json:"fields,omitempty"`
}

// Update a Jira issue IssueKey. Using `NewRequest` to allow `notifyUsers` param
func (ji *JiraImporter) updateIssue(issueKey string, data *issueUpdateData) error {
	bytesMessage, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// fmt.Println(string(bytesMessage))

	_, err = ji.sendJiraRequest(http.MethodPut, fmt.Sprintf("/issue/%s?notifyUsers=false", issueKey), bytes.NewBuffer(bytesMessage))
	if err != nil {
		return err
	}

	fmt.Printf("Successfully updated issue %s\n", issueKey)

	return nil
}

// Find parent IssueKey by its JiraID
func findParentIssueKey(parentID string, issues []*Issue) string {
	for _, issue := range issues {
		if issue.IssueID == parentID {
			return issue.IssueKey
		}
	}
	return ""
}

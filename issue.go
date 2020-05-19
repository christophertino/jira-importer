package jiraimporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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
	IssueSplit  string `csv:"Outward issue link (Issue split)"`
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

// Update a Jira issue by ID. Using `NewRequest` to allow `notifyUsers` param
func (ji *JiraImporter) updateIssue(issueID string, data *issueUpdateData) error {
	bytesMessage, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// fmt.Println(string(bytesMessage))

	_, err = ji.JiraClient.NewRequest(http.MethodPut, fmt.Sprintf("/rest/api/3/issue/%s?notifyUsers=false", issueID), bytes.NewBuffer(bytesMessage))
	if err != nil {
		return err
	}

	return nil
}

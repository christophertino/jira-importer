package jiraimporter

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/andygrunwald/go-jira"
)

// JiraImporter creates a struct of config info
type JiraImporter struct {
	JiraEmail  string
	JiraToken  string
	CSVPath    string
	JiraClient *jira.Client
}

// Custom error type for Jira API requests
type jiraError struct {
	Errors    []string `json:"errors"`
	ErrorType string   `json:"error_type"`
}

// Format Jira errors
func (je jiraError) Error() string {
	return fmt.Sprintf("Jira API Error: %s %s", je.Errors, je.ErrorType)
}

// Make a request to the JIRA API
func (ji *JiraImporter) sendJiraRequest(method string, path string, body io.Reader) (*http.Response, error) {
	// Prepare Jira API request
	req, err := http.NewRequest(method, fmt.Sprintf("https://ghostery.atlassian.net/rest/api/3/%s", strings.TrimPrefix(path, "/")), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(ji.JiraEmail, ji.JiraToken)

	// Send request
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// Handle API errors
	if res.StatusCode >= http.StatusBadRequest {
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		jErr := jiraError{}
		if err := json.Unmarshal(bodyBytes, &jErr); err != nil {
			return nil, err
		}
		return nil, jErr
	}

	return res, nil
}

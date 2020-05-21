package jiraimporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/andygrunwald/go-jira"
)

type jqlBody struct {
	Jql        string `json:"jql"`
	MaxResults int    `json:"maxResults"`
	StartAt    int    `json:"startAt"`
}

// Fetch all project components from the previous Jira instance
func (ji *JiraImporter) getLegacyProjectComponents(projectKey string) ([]*jira.Component, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/rest/api/3/project/%s/components", ji.LegacyDomain, projectKey), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(ji.LegacyEmail, ji.LegacyToken)

	// Send request
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// Handle API errors
	if res.StatusCode >= http.StatusBadRequest {
		jError := &jiraError{}
		if err := json.Unmarshal(bodyBytes, jError); err != nil {
			return nil, err
		}
		return nil, jError
	}

	components := []*jira.Component{}
	if err := json.Unmarshal(bodyBytes, &components); err != nil {
		return nil, err
	}

	return components, nil
}

// Get all of the issues assigned to a component
func (ji *JiraImporter) getLegacyComponentIssues(projectKey string, componentName string) ([]*jira.Issue, error) {
	jqlQuery := jqlBody{
		Jql:        fmt.Sprintf("project = \"%s\" AND component = \"%s\"", projectKey, componentName),
		MaxResults: 100,
		StartAt:    0,
	}
	bytesMessage, err := json.Marshal(jqlQuery)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, "/search", bytes.NewBuffer(bytesMessage))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(ji.LegacyEmail, ji.LegacyToken)

	// Send request
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// Handle API errors
	if res.StatusCode >= http.StatusBadRequest {
		jError := &jiraError{}
		if err := json.Unmarshal(bodyBytes, jError); err != nil {
			return nil, err
		}
		return nil, jError
	}

	issues := []*jira.Issue{}
	if err := json.Unmarshal(bodyBytes, &issues); err != nil {
		return nil, err
	}

	return issues, nil
}

package jiraimporter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/andygrunwald/go-jira"
)

// Fetch a Jira project by its Key
func (ji *JiraImporter) getProject(projectKey string) (*jira.Project, error) {
	res, err := ji.sendJiraRequest(http.MethodGet, fmt.Sprintf("/project/%s", projectKey), nil)
	if err != nil {
		return nil, err
	}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var jProject = jira.Project{}
	if err := json.Unmarshal(bodyBytes, &jProject); err != nil {
		return nil, err
	}

	return &jProject, nil
}

// Find an IssueType ID by name
func getIssueTypeID(typeName string, proj *jira.Project) (string, error) {
	for _, it := range proj.IssueTypes {
		if it.Name == typeName {
			return it.ID, nil
		}
	}
	return "", fmt.Errorf("IssueType name %s not found", typeName)
}

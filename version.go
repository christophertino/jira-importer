package jiraimporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/andygrunwald/go-jira"
)

// Get FixVersion by name
func getVersionInfo(versionName string, versions []*jira.FixVersion) *jira.FixVersion {
	for _, v := range versions {
		if versionName == v.Name {
			return v
		}
	}
	return nil
}

// Update a Jira FixVersion
func (ji *JiraImporter) updateVersion(versionID string, data *jira.FixVersion) error {
	bytesMessage, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// fmt.Println(string(bytesMessage))

	_, err = ji.sendJiraRequest(http.MethodPut, fmt.Sprintf("/version/%s", versionID), bytes.NewBuffer(bytesMessage))
	if err != nil {
		return err
	}

	fmt.Printf("Successfully updated version %s\n", versionID)

	return nil
}

// Fetch all project versions from the previous Jira instance
func (ji *JiraImporter) getLegacyProjectVersions(projectKey string) ([]*jira.FixVersion, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/rest/api/3/project/%s/versions", ji.LegacyDomain, projectKey), nil)
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

	versions := []*jira.FixVersion{}
	if err := json.Unmarshal(bodyBytes, &versions); err != nil {
		return nil, err
	}

	return versions, nil
}

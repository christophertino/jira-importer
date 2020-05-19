package jiraimporter

import (
	"fmt"

	"github.com/andygrunwald/go-jira"
)

// Find an IssueType ID by name. Issue Types are specific to a project
func getIssueTypeID(typeName string, proj *jira.Project) (string, error) {
	for _, it := range proj.IssueTypes {
		if it.Name == typeName {
			return it.ID, nil
		}
	}
	return "", fmt.Errorf("IssueType name %s not found", typeName)
}

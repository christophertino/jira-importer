package jiraimporter

import (
	"github.com/andygrunwald/go-jira"
)

// JiraImporter creates a struct of config info
type JiraImporter struct {
	CSVPath    string
	JiraClient *jira.Client
}

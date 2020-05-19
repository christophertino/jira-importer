package jiraimporter

import (
	"github.com/andygrunwald/go-jira"
)

// JiraImporter creates a struct of config info
type JiraImporter struct {
	JiraEmail  string
	JiraToken  string
	CSVPath    string
	JiraClient *jira.Client
}

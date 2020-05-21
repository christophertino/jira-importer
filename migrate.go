package jiraimporter

import (
	"fmt"
	"strconv"
	"time"

	"github.com/andygrunwald/go-jira"
)

const rateLimit = time.Second / 100 // 100 calls per second

// MigrateIssues is the entry point for the Jira import migration
func (ji *JiraImporter) MigrateIssues() {
	// Load the CSV export
	issues, err := parseCSV(ji.CSVPath)
	if err != nil {
		fmt.Println("Error parsing CSV file:", err)
		return
	}

	// Get Project info
	project, _, err := ji.JiraClient.Project.Get(issues[0].ProjectKey)
	if err != nil {
		fmt.Println("Error getting Project info:", err)
		return
	}

	// Handle rate limiting
	throttle := time.Tick(rateLimit)

	// Convert issues to correct types
	for _, issue := range issues {
		<-throttle
		updateData := issueUpdateData{}
		issueType := issue.IssueType

		switch issue.IssueType {
		case "Epic":
			// We can't convert existing types to Epics, so they become Stories instead
			issueType = "Story"
			break
		case "Sub-task":
			// Next-gen removes hyphen in Sub-task
			issueType = "Subtask"
			break
		}

		// Was this previously the a child of an Epic?
		if issue.EpicLink != "" {
			if issueType == "Subtask" {
				// For subtasks, add the issue as the child of the new Story(epic)
				updateData.Fields.Parent.Key = issue.EpicLink
			} else {
				// Only subtasks can be children. Create a 'Blocks' relationship
				issueLink := jira.IssueLink{
					Type: jira.IssueLinkType{
						Name: "Blocks",
					},
					InwardIssue: &jira.Issue{
						Key: issue.IssueKey,
					},
					OutwardIssue: &jira.Issue{
						Key: issue.EpicLink,
					},
				}
				_, err := ji.JiraClient.Issue.AddLink(&issueLink)
				if err != nil {
					fmt.Printf("Could not add 'Blocks' relationship to %s\n: %s", issue.IssueKey, err)
				}
			}
		}

		// Get the IssueType ID by name
		typeID, err := getIssueTypeID(issueType, project)
		if err != nil {
			fmt.Printf("Error getting type for issue %s: %s\n", issue.IssueKey, err)
			continue
		}
		updateData.Fields.IssueType.ID = typeID

		// Add existing subtasks back to their parent issues
		if issue.ParentID != "" {
			if parentKey := findParentIssueKey(issue.ParentID, issues); parentKey != "" {
				updateData.Fields.Parent.Key = parentKey
			}
		}

		// Import story points
		if issue.StoryPoints != "" {
			points, err := strconv.Atoi(issue.StoryPoints)
			if err == nil {
				updateData.Update.StoryPoints = append(updateData.Update.StoryPoints, struct {
					Set int `json:"set,omitempty"`
				}{
					Set: points,
				})
			}
		}

		// Fix missing Issue Split relationships
		if issue.IssueSplit != "" {
			issueLink := jira.IssueLink{
				Type: jira.IssueLinkType{
					Name: "Issue split",
				},
				InwardIssue: &jira.Issue{
					Key: issue.IssueKey,
				},
				OutwardIssue: &jira.Issue{
					Key: issue.IssueSplit,
				},
			}
			_, err := ji.JiraClient.Issue.AddLink(&issueLink)
			if err != nil {
				fmt.Printf("Could not add 'Splits' relationship to %s\n: %s", issue.IssueKey, err)
			}
		}

		// Update the issue on Jira
		if err = ji.updateIssue(issue.IssueKey, &updateData); err != nil {
			fmt.Printf("Error updating issue %s: %s\n", issue.IssueKey, err)
		}
	}

	// TODO: Components become labels
}

// MigrateVersions migrates project FixVersions between Jira accounts
func (ji *JiraImporter) MigrateVersions() {
	// Load the CSV export
	issues, err := parseCSV(ji.CSVPath)
	if err != nil {
		fmt.Println("Error parsing CSV file:", err)
		return
	}

	// Get Project info
	project, _, err := ji.JiraClient.Project.Get(issues[0].ProjectKey)
	if err != nil {
		fmt.Println("Error getting Project info:", err)
		return
	}

	// Query project info from the legacy account
	legacyVersions, err := ji.getLegacyProjectVersions(project.Key)
	if err != nil {
		fmt.Println("Error getting legacy project versions:", err)
		return
	}

	// Look at all versions currently on the new Jira account
	for _, v := range project.Versions {
		// Lookup version info on the old account
		if legacyVersion := getVersionInfo(v.Name, legacyVersions); legacyVersion != nil {
			updatedVersion := &jira.FixVersion{
				Released:    legacyVersion.Released,
				ReleaseDate: legacyVersion.ReleaseDate,
			}
			// Update version on the new account
			if err := ji.updateVersion(v.ID, updatedVersion); err != nil {
				fmt.Println("Error updating version:", err)
				return
			}
		}
	}
}

// MigrateComponents migrates Components to Labels on a next-gen project
func (ji *JiraImporter) MigrateComponents() {
	// Load the CSV export
	issues, err := parseCSV(ji.CSVPath)
	if err != nil {
		fmt.Println("Error parsing CSV file:", err)
		return
	}

	// Get Project info
	project, _, err := ji.JiraClient.Project.Get(issues[0].ProjectKey)
	if err != nil {
		fmt.Println("Error getting Project info:", err)
		return
	}

	// Query project info from the legacy account
	legacyComponents, err := ji.getLegacyProjectComponents(project.Key)
	if err != nil {
		fmt.Println("Error getting legacy project components:", err)
		return
	}

	// Look at all versions currently on the new Jira account
	for _, c := range legacyComponents {
		// Search old jira account for any issues attached to the component
		issues, err := ji.getLegacyComponentIssues(project.Key, c.Name)
		if err != nil {
			fmt.Printf("Error getting Component %s info: %s\n", c.Name, err)
			continue
		}

		// Add label to issues on new account
		for _, i := range issues {
			var updateData = issueUpdateData{}
			updateData.Update.Labels = append(updateData.Update.Labels, struct {
				Add string `json:"add,omitempty"`
			}{
				Add: c.Name,
			})
			if err = ji.updateIssue(i.Key, &updateData); err != nil {
				fmt.Printf("Error updating issue %s: %s\n", i.Key, err)
			}
		}
	}
}

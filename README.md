# Jira Next-Gen Issue Importer

A set of helper tools for importing issues from a classic Jira project into a next-gen project. This application is meant to be used in conjunction with the JIRA's built-in CSV issue importer. Currently, that tool causes several bugs:

+ All issue types are forced to type 'Story' or 'Task' ([See bug](https://jira.atlassian.com/browse/JRACLOUD-72091))
+ Story points are removed
+ Epics lose their children
+ Sub-tasks lose their parent relationships
+ Component info is lost

This project will:

+ Set issues to their correct type
+ Add story points
+ Add sub-tasks to their correct parent tasks
+ Set Epics as Story type and create "Relates" issue type relationship with former children
	+ Currently there is not way to convert a Task to Epic
+ Add components as ticket labels (Next-gen does not support components)

## Configuration

```sh
# Install dependencies with Go Modules
$ go mod init
```

```sh
# Generate a local configuration file
$ cp .env-example .env
```

## Preparing for Import

#### In the Classic Project:
1. Issues & Filters > All Issues > Advanced Search > Export > Export Excel CSV (all fields)
2. Edit the Permission Scheme for the project (if migrating between Jira Cloud Instances)
	+ Set 'Browse Projects' permission group to 'Public'
	+ This allows the importer to migrate attachments
3. Edit the exported CSV file
	+ Change `Created`, `Updated`, `Last Viewed` and `Resolved` columns to date format `dd/MMM/yy h:mm am/pm`

#### In the Next-Gen Project
1. Project Settings > Issue Types
	+ Create additional issue types to match types in the classic project

## Run Jira's Built-In Import Tool

1. Jira Settings > External System Import > CSV
	+ Upload the CSV and configuration file (optional)
	+ Map CSV fields to Jira fields
	+ Check 'Map field value' for Status to confirm they are migrated correctly
2. Run the import

## Run the Import Helper

```sh
$ go run cmd/importer/main.go
```

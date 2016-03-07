package models

import (
	"sort"
	"time"
)

type StarEvent struct {
	StarredAt time.Time
}

type byStarredAt []StarEvent

func (a byStarredAt) Len() int           { return len(a) }
func (a byStarredAt) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byStarredAt) Less(i, j int) bool { return a[i].StarredAt.Before(a[j].StarredAt) }

type IssueEventType int

const (
	IssueOpened IssueEventType = iota
	IssueClosed
)

type IssueEvent struct {
	EventType IssueEventType
	IsPr      bool
	Timestamp time.Time
}

type byTimestamp []IssueEvent

func (a byTimestamp) Len() int           { return len(a) }
func (a byTimestamp) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byTimestamp) Less(i, j int) bool { return a[i].Timestamp.Before(a[j].Timestamp) }

func StarEventsFromApi(apiObjects []map[string]interface{}) ([]StarEvent, error) {
	starEvents := make([]StarEvent, len(apiObjects))
	for i := 0; i < len(apiObjects); i++ {
		starredAt, err := time.Parse(time.RFC3339, apiObjects[i]["starred_at"].(string))
		if err != nil {
			return nil, err
		}
		starEvents[i].StarredAt = starredAt
	}
	sort.Sort(byStarredAt(starEvents))
	return starEvents, nil
}

func IssueEventsFromApi(apiObjects []map[string]interface{}) ([]IssueEvent, error) {
	var issueEvents []IssueEvent
	for i := 0; i < len(apiObjects); i++ {
		issueOpened := IssueEvent{EventType: IssueOpened}
		_, issueOpened.IsPr = apiObjects[i]["pull_request"]
		createdAt, err := time.Parse(time.RFC3339, apiObjects[i]["created_at"].(string))
		if err != nil {
			return nil, err
		}
		issueOpened.Timestamp = createdAt
		issueEvents = append(issueEvents, issueOpened)

		if closedAt := apiObjects[i]["closed_at"]; closedAt != nil {
			issueClosed := IssueEvent{EventType: IssueClosed}
			issueClosed.IsPr = issueOpened.IsPr
			closedAt, err := time.Parse(time.RFC3339, closedAt.(string))
			if err != nil {
				return nil, err
			}
			issueClosed.Timestamp = closedAt
			issueEvents = append(issueEvents, issueClosed)
		}
	}
	sort.Sort(byTimestamp(issueEvents))
	return issueEvents, nil
}

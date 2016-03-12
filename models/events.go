package models

import (
	"sort"
	"time"

	"github.com/ksheedlo/ghviz/github"
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

func IssueEventsFromApi(issues []github.Issue) []IssueEvent {
	var issueEvents []IssueEvent

	for _, issue := range issues {
		issueEvents = append(issueEvents, IssueEvent{
			EventType: IssueOpened,
			IsPr:      issue.IsPr,
			Timestamp: issue.CreatedAt,
		})

		if issue.IsClosed {
			issueEvents = append(issueEvents, IssueEvent{
				EventType: IssueClosed,
				IsPr:      issue.IsPr,
				Timestamp: issue.ClosedAt,
			})
		}
	}
	sort.Sort(byTimestamp(issueEvents))
	return issueEvents
}

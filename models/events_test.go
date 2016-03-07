package models

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const starsJson string = `[
{"starred_at":"2016-03-07T03:25:41.469Z"},
{"starred_at":"2016-03-07T03:23:53.002Z"},
{"starred_at":"2016-03-07T03:26:14.739Z"}]`

func TestStarEventsFromApi(t *testing.T) {
	t.Parallel()
	stars := make([]map[string]interface{}, 3)
	json.Unmarshal([]byte(starsJson), &stars)
	starEvents, err := StarEventsFromApi(stars)
	assert.NoError(t, err)
	assert.Len(t, starEvents, 3)
	assert.True(t,
		starEvents[0].StarredAt.Before(starEvents[1].StarredAt),
		"Expected starEvents[0] to be before starEvents[1] !",
	)
	assert.True(t,
		starEvents[1].StarredAt.Before(starEvents[2].StarredAt),
		"Expected starEvents[1] to be before starEvents[2] !",
	)
}

const badStarsJson string = `[{"starred_at":"fish"}]`

func TestStarEventsFromApiError(t *testing.T) {
	t.Parallel()
	stars := make([]map[string]interface{}, 1)
	json.Unmarshal([]byte(badStarsJson), &stars)
	_, err := StarEventsFromApi(stars)
	assert.Error(t, err)
}

const issuesJson string = `[
{"created_at":"2016-03-07T03:26:14.739Z"},
{"created_at":"2016-03-07T03:23:53.002Z","closed_at":"2016-03-07T03:25:41.469Z"},
{"created_at":"2016-03-07T03:46:36.717Z","closed_at":"2016-03-07T03:46:55.993Z",
 "pull_request":{}},
{"created_at":"2016-03-07T03:46:46.458Z","pull_request":{}}]`

func TestIssueEventsFromApi(t *testing.T) {
	t.Parallel()
	issues := make([]map[string]interface{}, 4)
	json.Unmarshal([]byte(issuesJson), &issues)
	issueEvents, err := IssueEventsFromApi(issues)
	assert.NoError(t, err)
	assert.Len(t, issueEvents, 6)

	for i := 0; (i + 1) < len(issueEvents); i++ {
		assert.True(t,
			issueEvents[i].Timestamp.Before(issueEvents[i+1].Timestamp),
			fmt.Sprintf("Expected issueEvents[%d] to be before issueEvents[%d]", i, i+1),
		)
	}

	assert.Equal(t, issueEvents[0].EventType, IssueOpened)
	assert.False(t, issueEvents[0].IsPr, "Expected issueEvents[0] not to be a PR")
	assert.Equal(t, issueEvents[1].EventType, IssueClosed)
	assert.False(t, issueEvents[1].IsPr, "Expected issueEvents[1] not to be a PR")
	assert.Equal(t, issueEvents[2].EventType, IssueOpened)
	assert.False(t, issueEvents[2].IsPr, "Expected issueEvents[2] not to be a PR")
	assert.Equal(t, issueEvents[3].EventType, IssueOpened)
	assert.True(t, issueEvents[3].IsPr, "Expected issueEvents[3] to be a PR")
	assert.Equal(t, issueEvents[4].EventType, IssueOpened)
	assert.True(t, issueEvents[4].IsPr, "Expected issueEvents[4] to be a PR")
	assert.Equal(t, issueEvents[5].EventType, IssueClosed)
	assert.True(t, issueEvents[5].IsPr, "Expected issueEvents[5] to be a PR")
}

const issuesBadCreatedJson string = `[{"created_at":"fish"}]`

func TestIssueEventsFromApiBadCreatedAt(t *testing.T) {
	t.Parallel()
	issues := make([]map[string]interface{}, 4)
	json.Unmarshal([]byte(issuesBadCreatedJson), &issues)
	_, err := IssueEventsFromApi(issues)
	assert.Error(t, err)
}

const issuesBadClosedJson string = `[
{"created_at":"2016-03-07T03:26:14.739Z","closed_at":"fish"}]`

func TestIssueEventsFromApiBadClosedAt(t *testing.T) {
	t.Parallel()
	issues := make([]map[string]interface{}, 4)
	json.Unmarshal([]byte(issuesBadClosedJson), &issues)
	_, err := IssueEventsFromApi(issues)
	assert.Error(t, err)
}

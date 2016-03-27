package models

import (
	"fmt"
	"testing"
	"time"

	"github.com/ksheedlo/ghviz/github"
	"github.com/stretchr/testify/assert"
)

func TestIssueEventsFromApi(t *testing.T) {
	t.Parallel()
	issues := []github.Issue{
		github.Issue{CreatedAt: time.Unix(3, 0), IsPr: false, IsClosed: false},
		github.Issue{
			CreatedAt: time.Unix(1, 0),
			IsPr:      false,
			IsClosed:  true,
			ClosedAt:  time.Unix(2, 0),
		},
		github.Issue{
			CreatedAt: time.Unix(4, 0),
			IsPr:      true,
			IsClosed:  true,
			ClosedAt:  time.Unix(6, 0),
		},
		github.Issue{CreatedAt: time.Unix(5, 0), IsPr: true, IsClosed: false},
	}
	issueEvents := IssueEventsFromApi(issues)
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

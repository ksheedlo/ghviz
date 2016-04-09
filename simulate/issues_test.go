package simulate

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/ksheedlo/ghviz/mocks"
	"github.com/ksheedlo/ghviz/models"
	"github.com/stretchr/testify/assert"
)

func TestOpenIssueAndPrCounts(t *testing.T) {
	t.Parallel()

	expectedPrCounts := []int{0, 0, 0, 1, 2, 1}
	expectedIssueCounts := []int{1, 0, 1, 1, 1, 1}

	issueEvents := []models.IssueEvent{
		models.IssueEvent{
			EventType: models.IssueOpened,
			IsPr:      false,
			Timestamp: time.Unix(1, 0),
		},
		models.IssueEvent{
			EventType: models.IssueClosed,
			IsPr:      false,
			Timestamp: time.Unix(2, 0),
		},
		models.IssueEvent{
			EventType: models.IssueOpened,
			IsPr:      false,
			Timestamp: time.Unix(3, 0),
		},
		models.IssueEvent{
			EventType: models.IssueOpened,
			IsPr:      true,
			Timestamp: time.Unix(4, 0),
		},
		models.IssueEvent{
			EventType: models.IssueOpened,
			IsPr:      true,
			Timestamp: time.Unix(5, 0),
		},
		models.IssueEvent{
			EventType: models.IssueClosed,
			IsPr:      true,
			Timestamp: time.Unix(6, 0),
		},
	}

	issueCounts := OpenIssueAndPrCounts(issueEvents)
	assert.Len(t, issueCounts, 6)

	for i := 0; (i + 1) < len(issueCounts); i++ {
		assert.True(t,
			issueCounts[i].Timestamp.Before(issueCounts[i+1].Timestamp),
			fmt.Sprintf("Expected issueCounts[%d] to be before issueCounts[%d]", i, i+1),
		)
	}

	for i := 0; i < len(issueCounts); i++ {
		assert.Equal(t, expectedIssueCounts[i], issueCounts[i].OpenIssues)
		assert.Equal(t, expectedPrCounts[i], issueCounts[i].OpenPrs)
	}
}

func TestOpenIssueToJSON(t *testing.T) {
	t.Parallel()

	issueCount := OpenIssueAndPrCount{
		OpenIssues: 2,
		OpenPrs:    5,
		Timestamp:  time.Unix(1458966366, 892000000).UTC(),
	}
	jsonBytes := mocks.MarshalJSON(t, &issueCount)
	var issueMap map[string]interface{}
	assert.NoError(t, json.Unmarshal(jsonBytes, &issueMap))
	assert.Equal(t, 2.0, issueMap["open_issues"].(float64))
	assert.Equal(t, 5.0, issueMap["open_prs"].(float64))
	assert.Equal(t, "2016-03-26T04:26:06.892Z", issueMap["timestamp"].(string))
}

package simulate

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ksheedlo/ghviz/models"
	"github.com/stretchr/testify/assert"
)

const issuesJson string = `[
{"created_at":"2016-03-07T03:26:14.739Z"},
{"created_at":"2016-03-07T03:23:53.002Z","closed_at":"2016-03-07T03:25:41.469Z"},
{"created_at":"2016-03-07T03:46:36.717Z","closed_at":"2016-03-07T03:46:55.993Z",
 "pull_request":{}},
{"created_at":"2016-03-07T03:46:46.458Z","pull_request":{}}]`

func TestOpenIssueAndPrCounts(t *testing.T) {
	t.Parallel()

	expectedPrCounts := []int{0, 0, 0, 1, 2, 1}
	expectedIssueCounts := []int{1, 0, 1, 1, 1, 1}

	issues := make([]map[string]interface{}, 4)
	json.Unmarshal([]byte(issuesJson), &issues)
	issueEvents, err := models.IssueEventsFromApi(issues)
	assert.NoError(t, err)

	issueCounts := OpenIssueAndPrCounts(issueEvents)
	assert.Len(t, issueCounts, 6)

	for i := 0; (i + 1) < len(issueCounts); i++ {
		assert.True(t,
			issueCounts[i].Timestamp.Before(issueCounts[i+1].Timestamp),
			fmt.Sprintf("Expected issueCounts[%d] to be before issueCounts[%d]", i, i+1),
		)
	}

	for i := 0; i < len(issueCounts); i++ {
		assert.Equal(t, issueCounts[i].OpenIssues, expectedIssueCounts[i])
		assert.Equal(t, issueCounts[i].OpenPrs, expectedPrCounts[i])
	}
}

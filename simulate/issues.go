package simulate

import (
	"encoding/json"
	"time"

	"github.com/ksheedlo/ghviz/models"
)

type OpenIssueAndPrCount struct {
	OpenIssues int
	OpenPrs    int
	Timestamp  time.Time
}

func (ict *OpenIssueAndPrCount) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"open_issues": ict.OpenIssues,
		"open_prs":    ict.OpenPrs,
		"timestamp":   ict.Timestamp,
	})
}

func OpenIssueAndPrCounts(issueEvents []models.IssueEvent) []OpenIssueAndPrCount {
	issueCounts := make([]OpenIssueAndPrCount, len(issueEvents))
	openIssues := 0
	openPrs := 0
	for i := 0; i < len(issueEvents); i++ {
		switch {
		case issueEvents[i].EventType == models.IssueOpened && issueEvents[i].IsPr:
			openPrs++
		case issueEvents[i].EventType == models.IssueClosed && issueEvents[i].IsPr:
			openPrs--
		case issueEvents[i].EventType == models.IssueOpened && (!issueEvents[i].IsPr):
			openIssues++
		default:
			openIssues--
		}
		issueCounts[i].OpenIssues = openIssues
		issueCounts[i].OpenPrs = openPrs
		issueCounts[i].Timestamp = issueEvents[i].Timestamp
	}
	return issueCounts
}

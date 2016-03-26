package simulate

import (
	"encoding/json"
	"time"

	"github.com/ksheedlo/ghviz/github"
)

type StarCount struct {
	Stars     int
	Timestamp time.Time
}

func (sc *StarCount) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"stars":     sc.Stars,
		"timestamp": sc.Timestamp,
	})
}

func StarCounts(starEvents []github.StarEvent) []StarCount {
	starCounts := make([]StarCount, len(starEvents))
	for i := 0; i < len(starEvents); i++ {
		starCounts[i].Stars = i + 1
		starCounts[i].Timestamp = starEvents[i].StarredAt
	}
	return starCounts
}

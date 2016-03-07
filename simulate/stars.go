package simulate

import (
	"time"

	"github.com/ksheedlo/ghviz/models"
)

type StarCount struct {
	Stars     int
	Timestamp time.Time
}

func StarCounts(starEvents []models.StarEvent) []StarCount {
	starCounts := make([]StarCount, len(starEvents))
	for i := 0; i < len(starEvents); i++ {
		starCounts[i].Stars = i + 1
		starCounts[i].Timestamp = starEvents[i].StarredAt
	}
	return starCounts
}

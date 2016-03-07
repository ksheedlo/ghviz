package simulate

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ksheedlo/ghviz/models"
	"github.com/stretchr/testify/assert"
)

const starsJson string = `[
{"starred_at":"2016-03-07T03:25:41.469Z"},
{"starred_at":"2016-03-07T03:23:53.002Z"},
{"starred_at":"2016-03-07T03:26:14.739Z"}]`

func TestStarCounts(t *testing.T) {
	t.Parallel()

	stars := make([]map[string]interface{}, 3)
	json.Unmarshal([]byte(starsJson), &stars)
	starEvents, err := models.StarEventsFromApi(stars)
	assert.NoError(t, err)
	starCounts := StarCounts(starEvents)

	for i := 0; (i + 1) < len(starCounts); i++ {
		assert.True(t,
			starCounts[i].Timestamp.Before(starCounts[i+1].Timestamp),
			fmt.Sprintf("Expected starCounts[%d] to be before starCounts[%d] !", i, i+1),
		)
		assert.Equal(t, starCounts[i].Stars, i+1)
	}
	assert.Equal(t, starCounts[len(starCounts)-1].Stars, len(starCounts))
}
